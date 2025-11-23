package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
	pb "github.com/moroshma/MiniToolStreamConnector/model"
)

// EgressClient implements domain.EgressClient using gRPC
type EgressClient struct {
	conn   *grpc.ClientConn
	client pb.EgressServiceClient
}

// NewEgressClient creates a new gRPC client for MiniToolStreamEgress
func NewEgressClient(serverAddr string, opts ...grpc.DialOption) (*EgressClient, error) {
	if serverAddr == "" {
		return nil, fmt.Errorf("server address is required")
	}

	// Default dial options if not provided
	if len(opts) == 0 {
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}

	// Connect to the server
	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", serverAddr, err)
	}

	client := pb.NewEgressServiceClient(conn)

	return &EgressClient{
		conn:   conn,
		client: client,
	}, nil
}

// Subscribe subscribes to notifications for a subject
func (c *EgressClient) Subscribe(ctx context.Context, config *domain.SubscriptionConfig) (domain.NotificationStream, error) {
	if config == nil {
		return nil, fmt.Errorf("subscription config cannot be nil")
	}

	if config.Subject == "" {
		return nil, fmt.Errorf("subject cannot be empty")
	}

	req := &pb.SubscribeRequest{
		Subject:     config.Subject,
		DurableName: config.DurableName,
	}

	if config.StartSequence != nil {
		seq := *config.StartSequence
		req.StartSequence = &seq
	}

	stream, err := c.client.Subscribe(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("subscribe failed: %w", err)
	}

	return &notificationStreamAdapter{stream: stream}, nil
}

// Fetch fetches messages from a subject
func (c *EgressClient) Fetch(ctx context.Context, config *domain.SubscriptionConfig) (domain.MessageStream, error) {
	if config == nil {
		return nil, fmt.Errorf("subscription config cannot be nil")
	}

	if config.Subject == "" {
		return nil, fmt.Errorf("subject cannot be empty")
	}

	req := &pb.FetchRequest{
		Subject:     config.Subject,
		DurableName: config.DurableName,
		BatchSize:   config.BatchSize,
	}

	stream, err := c.client.Fetch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	return &messageStreamAdapter{stream: stream}, nil
}

// GetLastSequence gets the last sequence number for a subject
func (c *EgressClient) GetLastSequence(ctx context.Context, subject string) (uint64, error) {
	if subject == "" {
		return 0, fmt.Errorf("subject cannot be empty")
	}

	req := &pb.GetLastSequenceRequest{
		Subject: subject,
	}

	resp, err := c.client.GetLastSequence(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("get last sequence failed: %w", err)
	}

	return resp.LastSequence, nil
}

// Close closes the gRPC connection
func (c *EgressClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// notificationStreamAdapter adapts gRPC stream to domain.NotificationStream
type notificationStreamAdapter struct {
	stream pb.EgressService_SubscribeClient
}

func (a *notificationStreamAdapter) Recv() (*domain.Notification, error) {
	notification, err := a.stream.Recv()
	if err != nil {
		return nil, err
	}

	return &domain.Notification{
		Subject:  notification.Subject,
		Sequence: notification.Sequence,
	}, nil
}

// messageStreamAdapter adapts gRPC stream to domain.MessageStream
type messageStreamAdapter struct {
	stream pb.EgressService_FetchClient
}

func (a *messageStreamAdapter) Recv() (*domain.ReceivedMessage, error) {
	msg, err := a.stream.Recv()
	if err != nil {
		return nil, err
	}

	return &domain.ReceivedMessage{
		Subject:   msg.Subject,
		Sequence:  msg.Sequence,
		Data:      msg.Data,
		Headers:   msg.Headers,
		Timestamp: msg.Timestamp.AsTime(),
	}, nil
}
