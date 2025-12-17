package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SQSClient sendMessage interface based on aws sdk
type SQSClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

// SQSProducer SQSProducer based on aws sqs.
type SQSProducer struct {
	client        SQSClient
	laborQueueURL string
	emailQueueURL string
}

// NewSQSProducer new instance of SQS producer.
func NewSQSProducer(client SQSClient, laborQueueURL string, emailQueueURL string) *SQSProducer {
	return &SQSProducer{
		client:        client,
		laborQueueURL: laborQueueURL,
		emailQueueURL: emailQueueURL,
	}
}

// PublishCheckOutEvent send checkout event to the SQS queue.
func (p *SQSProducer) PublishCheckOutEvent(ctx context.Context, event CheckOutEvent) error {

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.laborQueueURL),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"EventType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("CHECK_OUT"),
			},
		},
	}

	_, err = p.client.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send message to labor queue: %w", err)
	}

	return nil
}

// PublishEmailEvent send email event to the SQS queue.
func (p *SQSProducer) PublishEmailEvent(ctx context.Context, event EmailEvent) error {

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.emailQueueURL),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"EventType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("EMAIL_SENT"),
			},
		},
	}

	_, err = p.client.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send message to email queue: %w", err)
	}
	return nil
}
