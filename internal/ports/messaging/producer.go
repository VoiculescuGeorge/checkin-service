package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Producer struct {
	sender        MessageSender
	laborQueueURL string
	emailQueueURL string
}

func NewProducer(sender MessageSender, laborQueueURL, emailQueueURL string) *Producer {
	return &Producer{
		sender:        sender,
		laborQueueURL: laborQueueURL,
		emailQueueURL: emailQueueURL,
	}
}

func NewSQSProducer(client SQSClient, laborQueueURL, emailQueueURL string) *Producer {
	return NewProducer(&SQSSender{client: client}, laborQueueURL, emailQueueURL)
}

func (p *Producer) PublishLabor(ctx context.Context, body interface{}) error {
	return p.publish(ctx, p.laborQueueURL, body)
}

func (p *Producer) PublishEmail(ctx context.Context, body interface{}) error {
	return p.publish(ctx, p.emailQueueURL, body)
}

func (p *Producer) publish(ctx context.Context, destination string, body interface{}) error {
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

	if err := p.sender.SendMessage(ctx, destination, b); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}
