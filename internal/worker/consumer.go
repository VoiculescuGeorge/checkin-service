package worker

import (
	"context"

	"checkin.service/pkg/logger"
	"checkin.service/pkg/telemetry"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
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
	processor Processor // The logic to process a single message
	// Concurrency controls how many messages can be processed at the same time.
	Concurrency int
}

// NewWorker creates a new SQS worker, ready to be started.
func NewWorker(client SQSClient, url string, proc Processor) *Worker {
	return &Worker{
		client:      client,
		queueURL:    url,
		processor:   proc,
		Concurrency: 10, // Default to 10 concurrent processors
	}
}

// Start kicks off the worker's main loop for polling the SQS queue.
// It will run until the provided context is canceled.
func (w *Worker) Start(ctx context.Context) {
	log.Info().Int("concurrency", w.Concurrency).Msg("SQS Worker started. Polling for messages...")

	messagesCh := make(chan types.Message, w.Concurrency)

	// Start a pool of processor goroutines
	for i := 0; i < w.Concurrency; i++ {
		go w.processMessages(ctx, messagesCh)
	}

	// Start the poller in the main goroutine
	w.pollMessages(ctx, messagesCh)
}

// pollMessages is the poller loop that fetches messages from SQS and sends them to a channel.
func (w *Worker) pollMessages(ctx context.Context, messagesCh chan<- types.Message) {
	defer close(messagesCh) // Close channel to signal processors to stop

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Poller shutting down...")
			return
		default:
			output, err := w.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:              &w.queueURL,
				MaxNumberOfMessages:   int32(w.Concurrency), // Fetch as many messages as we have processors
				WaitTimeSeconds:       20,
				MessageAttributeNames: []string{"All"}, // Request attributes to get trace context
			})
			if err != nil {
				log.Error().Err(err).Msg("Error receiving messages")
				continue
			}
			log.Info().Int("count", len(output.Messages)).Msg("Received messages")
			for _, msg := range output.Messages {
				messagesCh <- msg
			}
		}
	}
}

// processMessages runs in a goroutine, listening for messages on a channel and processing them.
func (w *Worker) processMessages(ctx context.Context, messagesCh <-chan types.Message) {
	for msg := range messagesCh {
		w.handleSingleMessage(ctx, msg)
	}
}

// handleSingleMessage is where the real work happens for a single message. It calls the
// processor and then decides whether to delete the message or change its visibility for a retry.
func (w *Worker) handleSingleMessage(ctx context.Context, msg types.Message) {
	ctx, span := telemetry.StartSpanFromSQSMessage(ctx, msg)
	defer span.End()

	ctx = logger.EnrichContextWithLogger(ctx)

	shouldRetry, retryDelay, err := w.processor.Process(ctx, msg)

	if err != nil && shouldRetry {
		log.Ctx(ctx).Warn().Err(err).Int32("retry_delay", retryDelay).Msg("Processing failed, will retry")

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
		log.Ctx(ctx).Error().Err(err).Msg("Unrecoverable error processing message, will not retry")
	}
}
