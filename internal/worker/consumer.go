package worker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSClient interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	ChangeMessageVisibility(ctx context.Context, params *sqs.ChangeMessageVisibilityInput, optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error)
}

// Processor is a generic interface for any type that can process a message from SQS.
// This lets us reuse the main worker logic for different kinds of jobs.
type Processor interface {
	Process(ctx context.Context, msg types.Message) (shouldRetry bool, retryDelay int32, err error)
}

// Worker is our generic SQS message consumer. It polls a queue and passes
// messages off to a Processor.
type Worker struct {
	client    SQSClient
	queueURL  string
	processor Processor
}

// NewWorker creates a new SQS worker, ready to be started.
func NewWorker(client SQSClient, url string, proc Processor) *Worker {
	return &Worker{client: client, queueURL: url, processor: proc}
}

// Start kicks off the worker's main loop for polling the SQS queue.
// It will run until the provided context is canceled.
func (w *Worker) Start(ctx context.Context) {
	log.Println("SQS Worker started. Polling for messages...")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			output, err := w.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            &w.queueURL,
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     20,
			})
			if err != nil {
				log.Printf("Error receiving messages: %v", err)
				continue
			}

			for _, msg := range output.Messages {
				w.handleMessage(ctx, msg)
			}
		}
	}
}

// handleMessage is where the real work happens for a single message. It calls the
// processor and then decides whether to delete the message or make it visible again for a retry.
func (w *Worker) handleMessage(ctx context.Context, msg types.Message) {
	shouldRetry, retryDelay, err := w.processor.Process(ctx, msg)

	if err != nil && shouldRetry {
		log.Printf("Processing failed, will retry in %d seconds. Error: %v", retryDelay, err)

		_, _ = w.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
			QueueUrl:          &w.queueURL,
			ReceiptHandle:     msg.ReceiptHandle,
			VisibilityTimeout: retryDelay,
		})
		return
	}

	if err == nil {
		// Only delete on total success
		w.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      &w.queueURL,
			ReceiptHandle: msg.ReceiptHandle,
		})
	} else {
		// An unrecoverable error occurred (e.g., bad message format).
		log.Printf("Unrecoverable error processing message, will not retry. Error: %v", err)
	}
}
