package minitoolstream_connector

import (
	"fmt"

	"google.golang.org/grpc"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
	grpcClient "github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/infrastructure/grpc"
	subscriberUsecase "github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/usecase/subscriber"
)

// Subscriber re-exports domain.Subscriber interface
type Subscriber = domain.Subscriber

// ReceivedMessage re-exports domain.ReceivedMessage
type ReceivedMessage = domain.ReceivedMessage

// Notification re-exports domain.Notification
type Notification = domain.Notification

// MessageHandler re-exports domain.MessageHandler
type MessageHandler = domain.MessageHandler

// MessageHandlerFunc re-exports domain.MessageHandlerFunc
type MessageHandlerFunc = domain.MessageHandlerFunc

// NewSubscriber creates a new subscriber with default configuration
func NewSubscriber(serverAddr string, durableName string, opts ...grpc.DialOption) (Subscriber, error) {
	if serverAddr == "" {
		return nil, fmt.Errorf("server address is required")
	}

	if durableName == "" {
		durableName = "default-subscriber"
	}

	// Create gRPC client
	client, err := grpcClient.NewEgressClient(serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create subscriber
	sub, err := subscriberUsecase.New(&subscriberUsecase.Config{
		Client:      client,
		DurableName: durableName,
		BatchSize:   10,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	return sub, nil
}

// SubscriberBuilder provides a fluent interface for building subscribers
type SubscriberBuilder struct {
	serverAddr  string
	durableName string
	batchSize   int32
	dialOpts    []grpc.DialOption
	logger      subscriberUsecase.Logger
	err         error
}

// NewSubscriberBuilder creates a new subscriber builder
func NewSubscriberBuilder(serverAddr string) *SubscriberBuilder {
	return &SubscriberBuilder{
		serverAddr: serverAddr,
		batchSize:  10,
	}
}

// WithDurableName sets the durable name
func (b *SubscriberBuilder) WithDurableName(durableName string) *SubscriberBuilder {
	b.durableName = durableName
	return b
}

// WithBatchSize sets the batch size for fetching messages
func (b *SubscriberBuilder) WithBatchSize(batchSize int32) *SubscriberBuilder {
	b.batchSize = batchSize
	return b
}

// WithDialOptions sets custom dial options
func (b *SubscriberBuilder) WithDialOptions(opts ...grpc.DialOption) *SubscriberBuilder {
	b.dialOpts = opts
	return b
}

// WithLogger sets a custom logger
func (b *SubscriberBuilder) WithLogger(logger subscriberUsecase.Logger) *SubscriberBuilder {
	b.logger = logger
	return b
}

// Build creates the subscriber instance
func (b *SubscriberBuilder) Build() (Subscriber, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.serverAddr == "" {
		return nil, fmt.Errorf("server address is required")
	}

	if b.durableName == "" {
		b.durableName = "default-subscriber"
	}

	// Create gRPC client
	client, err := grpcClient.NewEgressClient(b.serverAddr, b.dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create subscriber
	sub, err := subscriberUsecase.New(&subscriberUsecase.Config{
		Client:      client,
		DurableName: b.durableName,
		BatchSize:   b.batchSize,
		Logger:      b.logger,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	return sub, nil
}
