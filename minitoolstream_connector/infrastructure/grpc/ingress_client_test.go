package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
	pb "github.com/moroshma/MiniToolStreamConnector/model"
	"google.golang.org/grpc"
)

// Mock IngressServiceClient for testing
type mockIngressServiceClient struct {
	pb.IngressServiceClient
	publishFunc func(ctx context.Context, in *pb.PublishRequest, opts ...grpc.CallOption) (*pb.PublishResponse, error)
}

func (m *mockIngressServiceClient) Publish(ctx context.Context, in *pb.PublishRequest, opts ...grpc.CallOption) (*pb.PublishResponse, error) {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, in, opts...)
	}
	return &pb.PublishResponse{
		Sequence:     1,
		ObjectName:   "test-object",
		StatusCode:   0,
		ErrorMessage: "",
	}, nil
}

func TestNewIngressClient(t *testing.T) {
	t.Run("empty server address", func(t *testing.T) {
		client, err := NewIngressClient("")
		if err == nil {
			t.Fatal("expected error for empty server address")
		}
		if client != nil {
			t.Error("expected nil client")
		}
		if err.Error() != "server address is required" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("with custom dial options", func(t *testing.T) {
		// This test will fail to connect, but we're testing that options are passed
		opts := []grpc.DialOption{
			grpc.WithBlock(),
		}
		client, _ := NewIngressClient("invalid:9999", opts...)
		// We expect an error since we can't connect to invalid address
		if client != nil {
			client.Close()
		}
		// We just verify the function accepts options without panicking
	})
}

func TestIngressClient_Publish(t *testing.T) {
	t.Run("successful publish", func(t *testing.T) {
		mockClient := &mockIngressServiceClient{
			publishFunc: func(ctx context.Context, in *pb.PublishRequest, opts ...grpc.CallOption) (*pb.PublishResponse, error) {
				if in.Subject != "test.subject" {
					t.Errorf("expected subject 'test.subject', got %s", in.Subject)
				}
				if string(in.Data) != "test data" {
					t.Errorf("expected data 'test data', got %s", string(in.Data))
				}
				return &pb.PublishResponse{
					Sequence:     42,
					ObjectName:   "object-42",
					StatusCode:   0,
					ErrorMessage: "",
				}, nil
			},
		}

		client := &IngressClient{
			conn:   nil, // not needed for this test
			client: mockClient,
		}

		msg := &domain.PublishMessage{
			Subject: "test.subject",
			Data:    []byte("test data"),
			Headers: map[string]string{"key": "value"},
		}

		result, err := client.Publish(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Sequence != 42 {
			t.Errorf("expected sequence 42, got %d", result.Sequence)
		}
		if result.ObjectName != "object-42" {
			t.Errorf("expected object name 'object-42', got %s", result.ObjectName)
		}
		if result.StatusCode != 0 {
			t.Errorf("expected status code 0, got %d", result.StatusCode)
		}
	})

	t.Run("nil message", func(t *testing.T) {
		client := &IngressClient{
			conn:   nil,
			client: &mockIngressServiceClient{},
		}

		_, err := client.Publish(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil message")
		}
		if err.Error() != "message cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("empty subject", func(t *testing.T) {
		client := &IngressClient{
			conn:   nil,
			client: &mockIngressServiceClient{},
		}

		msg := &domain.PublishMessage{
			Subject: "",
			Data:    []byte("test data"),
		}

		_, err := client.Publish(context.Background(), msg)
		if err == nil {
			t.Fatal("expected error for empty subject")
		}
		if err.Error() != "subject cannot be empty" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("grpc error", func(t *testing.T) {
		expectedErr := errors.New("grpc error")
		mockClient := &mockIngressServiceClient{
			publishFunc: func(ctx context.Context, in *pb.PublishRequest, opts ...grpc.CallOption) (*pb.PublishResponse, error) {
				return nil, expectedErr
			},
		}

		client := &IngressClient{
			conn:   nil,
			client: mockClient,
		}

		msg := &domain.PublishMessage{
			Subject: "test.subject",
			Data:    []byte("test data"),
		}

		_, err := client.Publish(context.Background(), msg)
		if err == nil {
			t.Fatal("expected error")
		}
		// Error should be wrapped
		if !errors.Is(err, expectedErr) {
			t.Errorf("error should wrap original error, got: %v", err)
		}
	})

	t.Run("server error in response", func(t *testing.T) {
		mockClient := &mockIngressServiceClient{
			publishFunc: func(ctx context.Context, in *pb.PublishRequest, opts ...grpc.CallOption) (*pb.PublishResponse, error) {
				return &pb.PublishResponse{
					Sequence:     0,
					ObjectName:   "",
					StatusCode:   500,
					ErrorMessage: "internal server error",
				}, nil
			},
		}

		client := &IngressClient{
			conn:   nil,
			client: mockClient,
		}

		msg := &domain.PublishMessage{
			Subject: "test.subject",
			Data:    []byte("test data"),
		}

		result, err := client.Publish(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.StatusCode != 500 {
			t.Errorf("expected status code 500, got %d", result.StatusCode)
		}
		if result.ErrorMessage != "internal server error" {
			t.Errorf("expected error message 'internal server error', got %s", result.ErrorMessage)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockClient := &mockIngressServiceClient{
			publishFunc: func(ctx context.Context, in *pb.PublishRequest, opts ...grpc.CallOption) (*pb.PublishResponse, error) {
				return nil, context.Canceled
			},
		}

		client := &IngressClient{
			conn:   nil,
			client: mockClient,
		}

		msg := &domain.PublishMessage{
			Subject: "test.subject",
			Data:    []byte("test data"),
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.Publish(ctx, msg)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestIngressClient_Close(t *testing.T) {
	t.Run("close with nil conn", func(t *testing.T) {
		client := &IngressClient{
			conn:   nil,
			client: &mockIngressServiceClient{},
		}

		err := client.Close()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
