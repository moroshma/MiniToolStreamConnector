package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
	pb "github.com/moroshma/MiniToolStreamConnector/model"
)

// IngressClient implements domain.IngressClient using gRPC
type IngressClient struct {
	conn   *grpc.ClientConn
	client pb.IngressServiceClient
}

// NewIngressClient creates a new gRPC client for MiniToolStreamIngress
func NewIngressClient(serverAddr string, opts ...grpc.DialOption) (*IngressClient, error) {
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

	client := pb.NewIngressServiceClient(conn)

	return &IngressClient{
		conn:   conn,
		client: client,
	}, nil
}

// Publish publishes a message to the specified subject
func (c *IngressClient) Publish(ctx context.Context, msg *domain.PublishMessage) (*domain.PublishResult, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	if msg.Subject == "" {
		return nil, fmt.Errorf("subject cannot be empty")
	}

	req := &pb.PublishRequest{
		Subject: msg.Subject,
		Data:    msg.Data,
		Headers: msg.Headers,
	}

	resp, err := c.client.Publish(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("publish failed: %w", err)
	}

	return &domain.PublishResult{
		Sequence:     resp.Sequence,
		ObjectName:   resp.ObjectName,
		StatusCode:   resp.StatusCode,
		ErrorMessage: resp.ErrorMessage,
	}, nil
}

// Close closes the gRPC connection
func (c *IngressClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
