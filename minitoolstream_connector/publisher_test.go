package minitoolstream_connector

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewPublisher(t *testing.T) {
	t.Run("empty server address", func(t *testing.T) {
		pub, err := NewPublisher("")
		if err == nil {
			t.Fatal("expected error for empty server address")
		}
		if pub != nil {
			t.Error("expected nil publisher")
		}
	})

	// Note: We can't test actual connection without running server
	// These tests verify the API works correctly
}

func TestNewPublisherBuilder(t *testing.T) {
	t.Run("create builder", func(t *testing.T) {
		builder := NewPublisherBuilder("localhost:9090")
		if builder == nil {
			t.Fatal("expected non-nil builder")
		}
		if builder.serverAddr != "localhost:9090" {
			t.Errorf("expected server addr 'localhost:9090', got %s", builder.serverAddr)
		}
	})
}

func TestPublisherBuilder_WithDialOptions(t *testing.T) {
	t.Run("set dial options", func(t *testing.T) {
		builder := NewPublisherBuilder("localhost:9090")
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		builder.WithDialOptions(opts...)

		if len(builder.dialOpts) != 1 {
			t.Errorf("expected 1 dial option, got %d", len(builder.dialOpts))
		}
	})

	t.Run("chain calls", func(t *testing.T) {
		builder := NewPublisherBuilder("localhost:9090").
			WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials()))

		if len(builder.dialOpts) != 1 {
			t.Errorf("expected 1 dial option, got %d", len(builder.dialOpts))
		}
	})
}

func TestPublisherBuilder_WithResultHandler(t *testing.T) {
	t.Run("set result handler", func(t *testing.T) {
		builder := NewPublisherBuilder("localhost:9090")
		handler := ResultHandlerFunc(func(ctx context.Context, result *PublishResult) error {
			return nil
		})

		builder.WithResultHandler(handler)

		if builder.resultHandler == nil {
			t.Error("expected result handler to be set")
		}
	})

	t.Run("chain calls", func(t *testing.T) {
		handler := ResultHandlerFunc(func(ctx context.Context, result *PublishResult) error {
			return nil
		})

		builder := NewPublisherBuilder("localhost:9090").
			WithResultHandler(handler)

		if builder.resultHandler == nil {
			t.Error("expected result handler to be set")
		}
	})
}

func TestPublisherBuilder_Build(t *testing.T) {
	t.Run("empty server address", func(t *testing.T) {
		builder := NewPublisherBuilder("")
		pub, err := builder.Build()

		if err == nil {
			t.Fatal("expected error for empty server address")
		}
		if pub != nil {
			t.Error("expected nil publisher")
		}
	})

	t.Run("with error set", func(t *testing.T) {
		builder := NewPublisherBuilder("localhost:9090")
		builder.err = context.Canceled // Use a known error

		pub, err := builder.Build()
		if err == nil {
			t.Fatal("expected error when builder has error")
		}
		if pub != nil {
			t.Error("expected nil publisher")
		}
	})

	// Note: Can't test successful build without running server
}

func TestPublisherBuilder_FullChain(t *testing.T) {
	t.Run("complete builder chain", func(t *testing.T) {
		handler := ResultHandlerFunc(func(ctx context.Context, result *PublishResult) error {
			return nil
		})

		builder := NewPublisherBuilder("localhost:9090").
			WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials())).
			WithResultHandler(handler)

		if builder.serverAddr != "localhost:9090" {
			t.Errorf("expected server addr 'localhost:9090', got %s", builder.serverAddr)
		}
		if len(builder.dialOpts) != 1 {
			t.Errorf("expected 1 dial option, got %d", len(builder.dialOpts))
		}
		if builder.resultHandler == nil {
			t.Error("expected result handler to be set")
		}
	})
}
