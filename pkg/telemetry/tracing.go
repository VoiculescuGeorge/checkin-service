package telemetry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const EmployeeIDKey contextKey = "employeeId"

// InitTracer initializes the OpenTelemetry tracer provider.
func InitTracer(serviceName string) (func(context.Context) error, error) {
	ctx := context.Background()

	// Configure the OTLP exporter to send traces to Jaeger via gRPC.
	// Ensure "jaeger" is reachable at port 4317 (e.g., defined in docker-compose).
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint("jaeger:4317"))
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp.Shutdown, nil
}

// StartSpanFromSQSMessage extracts trace context from SQS message attributes and starts a new span.
func StartSpanFromSQSMessage(ctx context.Context, msg types.Message) (context.Context, trace.Span) {
	// Extract trace context from SQS message attributes
	carrier := sqsCarrier{attrs: msg.MessageAttributes}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	tracer := otel.Tracer("sqs-worker")
	ctx, span := tracer.Start(ctx, "process_message",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "aws_sqs"),
			attribute.String("messaging.message_id", *msg.MessageId),
		),
	)

	// Attempt to extract employee_id from the message body to enrich the trace
	if msg.Body != nil {
		var payload struct {
			EmployeeID string `json:"employeeId"`
		}
		if err := json.Unmarshal([]byte(*msg.Body), &payload); err == nil && payload.EmployeeID != "" {
			span.SetAttributes(attribute.String("app.employeeId", payload.EmployeeID))
			ctx = context.WithValue(ctx, EmployeeIDKey, payload.EmployeeID)
		}
	}
	return ctx, span
}

// GetEmployeeIDFromContext retrieves the employee ID from the context.
func GetEmployeeIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(EmployeeIDKey).(string); ok {
		return val
	}
	return ""
}

// InjectTraceContext injects the current trace context into SQS message attributes.
func InjectTraceContext(ctx context.Context) map[string]types.MessageAttributeValue {
	attrs := make(map[string]types.MessageAttributeValue)
	carrier := sqsCarrier{attrs: attrs}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return attrs
}

// sqsCarrier implements propagation.TextMapCarrier to extract trace context from SQS attributes.
type sqsCarrier struct {
	attrs map[string]types.MessageAttributeValue
}

func (c sqsCarrier) Get(key string) string {
	if attr, ok := c.attrs[key]; ok && attr.StringValue != nil {
		return *attr.StringValue
	}
	return ""
}

func (c sqsCarrier) Set(key string, value string) {
	c.attrs[key] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(value),
	}
}

func (c sqsCarrier) Keys() []string {
	keys := make([]string, 0, len(c.attrs))
	for k := range c.attrs {
		keys = append(keys, k)
	}
	return keys
}
