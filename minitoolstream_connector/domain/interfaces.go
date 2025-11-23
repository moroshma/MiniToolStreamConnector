package domain

import (
	"context"
	"io"
)

// IngressClient represents the interface for communicating with MiniToolStreamIngress
type IngressClient interface {
	Publish(ctx context.Context, msg *PublishMessage) (*PublishResult, error)
	Close() error
}

// EgressClient represents the interface for communicating with MiniToolStreamEgress
type EgressClient interface {
	Subscribe(ctx context.Context, config *SubscriptionConfig) (NotificationStream, error)
	Fetch(ctx context.Context, config *SubscriptionConfig) (MessageStream, error)
	GetLastSequence(ctx context.Context, subject string) (uint64, error)
	Close() error
}

// NotificationStream represents a stream of notifications
type NotificationStream interface {
	Recv() (*Notification, error)
}

// MessageStream represents a stream of messages
type MessageStream interface {
	Recv() (*ReceivedMessage, error)
}

// Publisher represents the interface for publishing messages
type Publisher interface {
	Publish(ctx context.Context, preparer MessagePreparer) error
	PublishAll(ctx context.Context, preparers []MessagePreparer) error
	RegisterHandler(preparer MessagePreparer)
	RegisterHandlers(preparers []MessagePreparer)
	SetResultHandler(handler ResultHandler)
	Close() error
}

// Subscriber represents the interface for subscribing to subjects
type Subscriber interface {
	RegisterHandler(subject string, handler MessageHandler)
	RegisterHandlers(handlers map[string]MessageHandler)
	Start() error
	Stop()
	Wait()
}

// IsEOF checks if error is EOF
func IsEOF(err error) bool {
	return err == io.EOF
}
