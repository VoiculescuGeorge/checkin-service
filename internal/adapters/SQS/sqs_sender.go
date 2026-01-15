package sqsadapter

import (
	"context"

	"checkin.service/internal/ports"
	"checkin.service/pkg/telemetry"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSSender implements MessageSender for AWS SQS.
type SQSSender struct {
	client SQSClient
}

// SQSClient defines the interface for the AWS SQS client.
type SQSClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
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

// NewSQSProducer creates a new Producer backed by an AWS SQS sender.
func NewSQSProducer(client SQSClient, laborQueueURL, emailQueueURL string) *ports.Producer {
	return ports.NewProducer(&SQSSender{client: client}, laborQueueURL, emailQueueURL)
}
