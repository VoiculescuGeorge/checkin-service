package ports

import (
	"context"
)

// QeuueProducer defines the output port for publishing domain events.
type QeuueProducer interface {
	PublishLabor(ctx context.Context, body interface{}) error
	PublishEmail(ctx context.Context, body interface{}) error
}

// MessageSender defines the interface for sending raw messages to a messaging system.
type MessageSender interface {
	SendMessage(ctx context.Context, destination string, body []byte) error
}

type Producer struct {
	sender        MessageSender
	laborQueueURL string
	emailQueueURL string
}
