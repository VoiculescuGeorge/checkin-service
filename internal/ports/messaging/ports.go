package messaging

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSProducer defines the output port for publishing domain events.
type SQSProducer interface {
	PublishLabor(ctx context.Context, body interface{}) error
	PublishEmail(ctx context.Context, body interface{}) error
}

// MessageSender defines the interface for sending raw messages to a messaging system.
type MessageSender interface {
	SendMessage(ctx context.Context, destination string, body []byte) error
}

// SQSClient defines the interface for the AWS SQS client.
type SQSClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}
