package core

import (
	"context"
	"fmt"

	"checkin.service/pkg/telemetry"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type EmailService interface {
	SendCheckOutSummary(ctx context.Context, to string, hours float64) error
}

type SESEmailService struct {
	client *ses.Client
	sender string
}

func NewSESEmailService(client *ses.Client, sender string) *SESEmailService {
	return &SESEmailService{client: client, sender: sender}
}

func (s *SESEmailService) SendCheckOutSummary(ctx context.Context, to string, hours float64) error {
	tracer := otel.Tracer("ses-email-service")
	ctx, span := tracer.Start(ctx, "send_email", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// Enrich span with employeeId if available in context
	if empID := telemetry.GetEmployeeIDFromContext(ctx); empID != "" {
		span.SetAttributes(attribute.String("app.employeeId", empID))
	}

	input := &ses.SendEmailInput{
		Source: aws.String(s.sender),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String("Work Shift Summary"),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(fmt.Sprintf("Hello,\n\nYou have successfully checked out. Total hours worked: %.2f hours.", hours)),
				},
			},
		},
	}

	_, err := s.client.SendEmail(ctx, input)
	return err
}
