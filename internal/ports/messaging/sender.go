package messaging

import (
	"context"

	"checkin.service/pkg/telemetry"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSSender implements MessageSender for AWS SQS.
type SQSSender struct {
	client SQSClient
}

func (s *SQSSender) SendMessage(ctx context.Context, destination string, body []byte) error {
	// Inject trace context into message attributes
	attributes := telemetry.InjectTraceContext(ctx)

	_, err := s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          aws.String(destination),
		MessageBody:       aws.String(string(body)),
		MessageAttributes: attributes,
	})
	return err
}
