package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"checkin.service/pkg/telemetry"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type SQSClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type SQSProducer struct {
	client        SQSClient
	laborQueueURL string
	emailQueueURL string
}

func NewSQSProducer(client SQSClient, laborQueueURL, emailQueueURL string) *SQSProducer {
	return &SQSProducer{
		client:        client,
		laborQueueURL: laborQueueURL,
		emailQueueURL: emailQueueURL,
	}
}

func (p *SQSProducer) PublishLabor(ctx context.Context, body interface{}) error {
	return p.publish(ctx, p.laborQueueURL, body)
}

func (p *SQSProducer) PublishEmail(ctx context.Context, body interface{}) error {
	return p.publish(ctx, p.emailQueueURL, body)
}

func (p *SQSProducer) publish(ctx context.Context, queueURL string, body interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	// Enrich the current span with employee_id if available
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		var payload struct {
			EmployeeID string `json:"employeeId"`
		}
		// Attempt to unmarshal to extract employee_id
		if err := json.Unmarshal(b, &payload); err == nil && payload.EmployeeID != "" {
			span.SetAttributes(attribute.String("app.employeeId", payload.EmployeeID))
		}
	}

	// Inject trace context into message attributes
	attributes := telemetry.InjectTraceContext(ctx)

	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          aws.String(queueURL),
		MessageBody:       aws.String(string(b)),
		MessageAttributes: attributes,
	})

	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}
	return nil
}
