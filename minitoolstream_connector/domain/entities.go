package domain

import (
	"context"
	"time"
)

// PublishMessage represents a message to be published to Ingress
type PublishMessage struct {
	Subject string
	Data    []byte
	Headers map[string]string
}

// ReceivedMessage represents a message received from Egress
type ReceivedMessage struct {
	Subject   string
	Sequence  uint64
	Data      []byte
	Headers   map[string]string
	Timestamp time.Time
}

// PublishResult represents the result of a publish operation
type PublishResult struct {
	Sequence     uint64
	ObjectName   string
	StatusCode   int64
	ErrorMessage string
}

// Notification represents a notification about new messages
type Notification struct {
	Subject  string
	Sequence uint64
}

// SubscriptionConfig represents subscription configuration
type SubscriptionConfig struct {
	Subject       string
	DurableName   string
	StartSequence *uint64
	BatchSize     int32
}

// MessagePreparer prepares messages for publishing
type MessagePreparer interface {
	Prepare(ctx context.Context) (*PublishMessage, error)
}

// MessagePreparerFunc is a function adapter for MessagePreparer
type MessagePreparerFunc func(ctx context.Context) (*PublishMessage, error)

// Prepare implements MessagePreparer interface
func (f MessagePreparerFunc) Prepare(ctx context.Context) (*PublishMessage, error) {
	return f(ctx)
}

// ResultHandler processes publish results
type ResultHandler interface {
	Handle(ctx context.Context, result *PublishResult) error
}

// ResultHandlerFunc is a function adapter for ResultHandler
type ResultHandlerFunc func(ctx context.Context, result *PublishResult) error

// Handle implements ResultHandler interface
func (f ResultHandlerFunc) Handle(ctx context.Context, result *PublishResult) error {
	return f(ctx, result)
}

// MessageHandler processes received messages
type MessageHandler interface {
	Handle(ctx context.Context, msg *ReceivedMessage) error
}

// MessageHandlerFunc is a function adapter for MessageHandler
type MessageHandlerFunc func(ctx context.Context, msg *ReceivedMessage) error

// Handle implements MessageHandler interface
func (f MessageHandlerFunc) Handle(ctx context.Context, msg *ReceivedMessage) error {
	return f(ctx, msg)
}
