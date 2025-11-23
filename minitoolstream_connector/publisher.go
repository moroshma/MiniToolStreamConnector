package minitoolstream_connector

import (
	"fmt"

	"google.golang.org/grpc"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
	grpcClient "github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/infrastructure/grpc"
	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/usecase/publisher"
)

// Publisher re-exports domain.Publisher interface
type Publisher = domain.Publisher

// PublishMessage re-exports domain.PublishMessage
type PublishMessage = domain.PublishMessage

// PublishResult re-exports domain.PublishResult
type PublishResult = domain.PublishResult

// MessagePreparer re-exports domain.MessagePreparer
type MessagePreparer = domain.MessagePreparer

// MessagePreparerFunc re-exports domain.MessagePreparerFunc
type MessagePreparerFunc = domain.MessagePreparerFunc

// ResultHandler re-exports domain.ResultHandler
type ResultHandler = domain.ResultHandler

// ResultHandlerFunc re-exports domain.ResultHandlerFunc
type ResultHandlerFunc = domain.ResultHandlerFunc

// NewPublisher creates a new publisher with default configuration
func NewPublisher(serverAddr string, opts ...grpc.DialOption) (Publisher, error) {
	if serverAddr == "" {
		return nil, fmt.Errorf("server address is required")
	}

	// Create gRPC client
	client, err := grpcClient.NewIngressClient(serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create publisher
	pub, err := publisher.New(&publisher.Config{
		Client: client,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return pub, nil
}

// PublisherBuilder provides a fluent interface for building publishers
type PublisherBuilder struct {
	serverAddr    string
	dialOpts      []grpc.DialOption
	resultHandler domain.ResultHandler
	err           error
}

// NewPublisherBuilder creates a new publisher builder
func NewPublisherBuilder(serverAddr string) *PublisherBuilder {
	return &PublisherBuilder{
		serverAddr: serverAddr,
	}
}

// WithDialOptions sets custom dial options
func (b *PublisherBuilder) WithDialOptions(opts ...grpc.DialOption) *PublisherBuilder {
	b.dialOpts = opts
	return b
}

// WithResultHandler sets a custom result handler
func (b *PublisherBuilder) WithResultHandler(handler domain.ResultHandler) *PublisherBuilder {
	b.resultHandler = handler
	return b
}

// Build creates the publisher instance
func (b *PublisherBuilder) Build() (Publisher, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.serverAddr == "" {
		return nil, fmt.Errorf("server address is required")
	}

	// Create gRPC client
	client, err := grpcClient.NewIngressClient(b.serverAddr, b.dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create publisher
	pub, err := publisher.New(&publisher.Config{
		Client:        client,
		ResultHandler: b.resultHandler,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return pub, nil
}
