package grpc

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
	pb "github.com/moroshma/MiniToolStreamConnector/model"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock EgressServiceClient for testing
type mockEgressServiceClient struct {
	pb.EgressServiceClient
	subscribeFunc      func(ctx context.Context, in *pb.SubscribeRequest, opts ...grpc.CallOption) (pb.EgressService_SubscribeClient, error)
	fetchFunc          func(ctx context.Context, in *pb.FetchRequest, opts ...grpc.CallOption) (pb.EgressService_FetchClient, error)
	getLastSequenceFunc func(ctx context.Context, in *pb.GetLastSequenceRequest, opts ...grpc.CallOption) (*pb.GetLastSequenceResponse, error)
}

func (m *mockEgressServiceClient) Subscribe(ctx context.Context, in *pb.SubscribeRequest, opts ...grpc.CallOption) (pb.EgressService_SubscribeClient, error) {
	if m.subscribeFunc != nil {
		return m.subscribeFunc(ctx, in, opts...)
	}
	return &mockSubscribeClient{}, nil
}

func (m *mockEgressServiceClient) Fetch(ctx context.Context, in *pb.FetchRequest, opts ...grpc.CallOption) (pb.EgressService_FetchClient, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, in, opts...)
	}
	return &mockFetchClient{}, nil
}

func (m *mockEgressServiceClient) GetLastSequence(ctx context.Context, in *pb.GetLastSequenceRequest, opts ...grpc.CallOption) (*pb.GetLastSequenceResponse, error) {
	if m.getLastSequenceFunc != nil {
		return m.getLastSequenceFunc(ctx, in, opts...)
	}
	return &pb.GetLastSequenceResponse{LastSequence: 0}, nil
}

// Mock subscribe stream
type mockSubscribeClient struct {
	grpc.ClientStream
	notifications []*pb.Notification
	index         int
	recvFunc      func() (*pb.Notification, error)
}

func (m *mockSubscribeClient) Recv() (*pb.Notification, error) {
	if m.recvFunc != nil {
		return m.recvFunc()
	}
	if m.index >= len(m.notifications) {
		return nil, io.EOF
	}
	notif := m.notifications[m.index]
	m.index++
	return notif, nil
}

// Mock fetch stream
type mockFetchClient struct {
	grpc.ClientStream
	messages []*pb.Message
	index    int
	recvFunc func() (*pb.Message, error)
}

func (m *mockFetchClient) Recv() (*pb.Message, error) {
	if m.recvFunc != nil {
		return m.recvFunc()
	}
	if m.index >= len(m.messages) {
		return nil, io.EOF
	}
	msg := m.messages[m.index]
	m.index++
	return msg, nil
}

func TestNewEgressClient(t *testing.T) {
	t.Run("empty server address", func(t *testing.T) {
		client, err := NewEgressClient("")
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
		opts := []grpc.DialOption{
			grpc.WithBlock(),
		}
		client, _ := NewEgressClient("invalid:9999", opts...)
		if client != nil {
			client.Close()
		}
		// We just verify the function accepts options without panicking
	})
}

func TestEgressClient_Subscribe(t *testing.T) {
	t.Run("successful subscribe", func(t *testing.T) {
		mockClient := &mockEgressServiceClient{
			subscribeFunc: func(ctx context.Context, in *pb.SubscribeRequest, opts ...grpc.CallOption) (pb.EgressService_SubscribeClient, error) {
				if in.Subject != "test.subject" {
					t.Errorf("expected subject 'test.subject', got %s", in.Subject)
				}
				if in.DurableName != "test-consumer" {
					t.Errorf("expected durable name 'test-consumer', got %s", in.DurableName)
				}
				return &mockSubscribeClient{
					notifications: []*pb.Notification{
						{Subject: "test.subject", Sequence: 1},
						{Subject: "test.subject", Sequence: 2},
					},
				}, nil
			},
		}

		client := &EgressClient{
			conn:   nil,
			client: mockClient,
		}

		config := &domain.SubscriptionConfig{
			Subject:     "test.subject",
			DurableName: "test-consumer",
		}

		stream, err := client.Subscribe(context.Background(), config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Read first notification
		notif, err := stream.Recv()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if notif.Subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", notif.Subject)
		}
		if notif.Sequence != 1 {
			t.Errorf("expected sequence 1, got %d", notif.Sequence)
		}

		// Read second notification
		notif, err = stream.Recv()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if notif.Sequence != 2 {
			t.Errorf("expected sequence 2, got %d", notif.Sequence)
		}

		// Should get EOF
		_, err = stream.Recv()
		if err != io.EOF {
			t.Errorf("expected EOF, got %v", err)
		}
	})

	t.Run("nil config", func(t *testing.T) {
		client := &EgressClient{
			conn:   nil,
			client: &mockEgressServiceClient{},
		}

		_, err := client.Subscribe(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
		if err.Error() != "subscription config cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("empty subject", func(t *testing.T) {
		client := &EgressClient{
			conn:   nil,
			client: &mockEgressServiceClient{},
		}

		config := &domain.SubscriptionConfig{
			Subject:     "",
			DurableName: "test-consumer",
		}

		_, err := client.Subscribe(context.Background(), config)
		if err == nil {
			t.Fatal("expected error for empty subject")
		}
		if err.Error() != "subject cannot be empty" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("with start sequence", func(t *testing.T) {
		startSeq := uint64(100)
		mockClient := &mockEgressServiceClient{
			subscribeFunc: func(ctx context.Context, in *pb.SubscribeRequest, opts ...grpc.CallOption) (pb.EgressService_SubscribeClient, error) {
				if in.StartSequence == nil {
					t.Error("expected start sequence to be set")
				} else if *in.StartSequence != 100 {
					t.Errorf("expected start sequence 100, got %d", *in.StartSequence)
				}
				return &mockSubscribeClient{}, nil
			},
		}

		client := &EgressClient{
			conn:   nil,
			client: mockClient,
		}

		config := &domain.SubscriptionConfig{
			Subject:       "test.subject",
			DurableName:   "test-consumer",
			StartSequence: &startSeq,
		}

		_, err := client.Subscribe(context.Background(), config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("grpc error", func(t *testing.T) {
		expectedErr := errors.New("grpc error")
		mockClient := &mockEgressServiceClient{
			subscribeFunc: func(ctx context.Context, in *pb.SubscribeRequest, opts ...grpc.CallOption) (pb.EgressService_SubscribeClient, error) {
				return nil, expectedErr
			},
		}

		client := &EgressClient{
			conn:   nil,
			client: mockClient,
		}

		config := &domain.SubscriptionConfig{
			Subject:     "test.subject",
			DurableName: "test-consumer",
		}

		_, err := client.Subscribe(context.Background(), config)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestEgressClient_Fetch(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		now := time.Now()
		mockClient := &mockEgressServiceClient{
			fetchFunc: func(ctx context.Context, in *pb.FetchRequest, opts ...grpc.CallOption) (pb.EgressService_FetchClient, error) {
				if in.Subject != "test.subject" {
					t.Errorf("expected subject 'test.subject', got %s", in.Subject)
				}
				if in.BatchSize != 10 {
					t.Errorf("expected batch size 10, got %d", in.BatchSize)
				}
				return &mockFetchClient{
					messages: []*pb.Message{
						{
							Subject:   "test.subject",
							Sequence:  1,
							Data:      []byte("data1"),
							Headers:   map[string]string{"key": "value1"},
							Timestamp: timestamppb.New(now),
						},
						{
							Subject:   "test.subject",
							Sequence:  2,
							Data:      []byte("data2"),
							Headers:   map[string]string{"key": "value2"},
							Timestamp: timestamppb.New(now),
						},
					},
				}, nil
			},
		}

		client := &EgressClient{
			conn:   nil,
			client: mockClient,
		}

		config := &domain.SubscriptionConfig{
			Subject:     "test.subject",
			DurableName: "test-consumer",
			BatchSize:   10,
		}

		stream, err := client.Fetch(context.Background(), config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Read first message
		msg, err := stream.Recv()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if msg.Sequence != 1 {
			t.Errorf("expected sequence 1, got %d", msg.Sequence)
		}
		if string(msg.Data) != "data1" {
			t.Errorf("expected data 'data1', got %s", string(msg.Data))
		}

		// Read second message
		msg, err = stream.Recv()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if msg.Sequence != 2 {
			t.Errorf("expected sequence 2, got %d", msg.Sequence)
		}

		// Should get EOF
		_, err = stream.Recv()
		if err != io.EOF {
			t.Errorf("expected EOF, got %v", err)
		}
	})

	t.Run("nil config", func(t *testing.T) {
		client := &EgressClient{
			conn:   nil,
			client: &mockEgressServiceClient{},
		}

		_, err := client.Fetch(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})

	t.Run("empty subject", func(t *testing.T) {
		client := &EgressClient{
			conn:   nil,
			client: &mockEgressServiceClient{},
		}

		config := &domain.SubscriptionConfig{
			Subject:   "",
			BatchSize: 10,
		}

		_, err := client.Fetch(context.Background(), config)
		if err == nil {
			t.Fatal("expected error for empty subject")
		}
	})
}

func TestEgressClient_GetLastSequence(t *testing.T) {
	t.Run("successful get last sequence", func(t *testing.T) {
		mockClient := &mockEgressServiceClient{
			getLastSequenceFunc: func(ctx context.Context, in *pb.GetLastSequenceRequest, opts ...grpc.CallOption) (*pb.GetLastSequenceResponse, error) {
				if in.Subject != "test.subject" {
					t.Errorf("expected subject 'test.subject', got %s", in.Subject)
				}
				return &pb.GetLastSequenceResponse{LastSequence: 42}, nil
			},
		}

		client := &EgressClient{
			conn:   nil,
			client: mockClient,
		}

		seq, err := client.GetLastSequence(context.Background(), "test.subject")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if seq != 42 {
			t.Errorf("expected sequence 42, got %d", seq)
		}
	})

	t.Run("empty subject", func(t *testing.T) {
		client := &EgressClient{
			conn:   nil,
			client: &mockEgressServiceClient{},
		}

		_, err := client.GetLastSequence(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty subject")
		}
		if err.Error() != "subject cannot be empty" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("grpc error", func(t *testing.T) {
		expectedErr := errors.New("grpc error")
		mockClient := &mockEgressServiceClient{
			getLastSequenceFunc: func(ctx context.Context, in *pb.GetLastSequenceRequest, opts ...grpc.CallOption) (*pb.GetLastSequenceResponse, error) {
				return nil, expectedErr
			},
		}

		client := &EgressClient{
			conn:   nil,
			client: mockClient,
		}

		_, err := client.GetLastSequence(context.Background(), "test.subject")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestEgressClient_Close(t *testing.T) {
	t.Run("close with nil conn", func(t *testing.T) {
		client := &EgressClient{
			conn:   nil,
			client: &mockEgressServiceClient{},
		}

		err := client.Close()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
