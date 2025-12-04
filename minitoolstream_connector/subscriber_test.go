package minitoolstream_connector

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewSubscriber(t *testing.T) {
	t.Run("empty server address", func(t *testing.T) {
		sub, err := NewSubscriber("", "test-consumer")
		if err == nil {
			t.Fatal("expected error for empty server address")
		}
		if sub != nil {
			t.Error("expected nil subscriber")
		}
	})

	t.Run("empty durable name", func(t *testing.T) {
		// Should use default durable name
		// Can't test actual connection without server
	})

	// Note: Can't test actual connection without running server
}

func TestNewSubscriberBuilder(t *testing.T) {
	t.Run("create builder", func(t *testing.T) {
		builder := NewSubscriberBuilder("localhost:9091")
		if builder == nil {
			t.Fatal("expected non-nil builder")
		}
		if builder.serverAddr != "localhost:9091" {
			t.Errorf("expected server addr 'localhost:9091', got %s", builder.serverAddr)
		}
		if builder.batchSize != 10 {
			t.Errorf("expected default batch size 10, got %d", builder.batchSize)
		}
	})
}

func TestSubscriberBuilder_WithDurableName(t *testing.T) {
	t.Run("set durable name", func(t *testing.T) {
		builder := NewSubscriberBuilder("localhost:9091")
		builder.WithDurableName("my-consumer")

		if builder.durableName != "my-consumer" {
			t.Errorf("expected durable name 'my-consumer', got %s", builder.durableName)
		}
	})

	t.Run("chain calls", func(t *testing.T) {
		builder := NewSubscriberBuilder("localhost:9091").
			WithDurableName("my-consumer")

		if builder.durableName != "my-consumer" {
			t.Errorf("expected durable name 'my-consumer', got %s", builder.durableName)
		}
	})
}

func TestSubscriberBuilder_WithBatchSize(t *testing.T) {
	t.Run("set batch size", func(t *testing.T) {
		builder := NewSubscriberBuilder("localhost:9091")
		builder.WithBatchSize(20)

		if builder.batchSize != 20 {
			t.Errorf("expected batch size 20, got %d", builder.batchSize)
		}
	})

	t.Run("chain calls", func(t *testing.T) {
		builder := NewSubscriberBuilder("localhost:9091").
			WithBatchSize(20)

		if builder.batchSize != 20 {
			t.Errorf("expected batch size 20, got %d", builder.batchSize)
		}
	})
}

func TestSubscriberBuilder_WithDialOptions(t *testing.T) {
	t.Run("set dial options", func(t *testing.T) {
		builder := NewSubscriberBuilder("localhost:9091")
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		builder.WithDialOptions(opts...)

		if len(builder.dialOpts) != 1 {
			t.Errorf("expected 1 dial option, got %d", len(builder.dialOpts))
		}
	})

	t.Run("chain calls", func(t *testing.T) {
		builder := NewSubscriberBuilder("localhost:9091").
			WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials()))

		if len(builder.dialOpts) != 1 {
			t.Errorf("expected 1 dial option, got %d", len(builder.dialOpts))
		}
	})
}

type testSubLogger struct {
	messages []string
}

func (l *testSubLogger) Printf(format string, v ...interface{}) {
	l.messages = append(l.messages, format)
}

func TestSubscriberBuilder_WithLogger(t *testing.T) {
	t.Run("set logger", func(t *testing.T) {
		logger := &testSubLogger{}
		builder := NewSubscriberBuilder("localhost:9091")
		builder.WithLogger(logger)

		if builder.logger != logger {
			t.Error("expected custom logger to be set")
		}
	})

	t.Run("chain calls", func(t *testing.T) {
		logger := &testSubLogger{}
		builder := NewSubscriberBuilder("localhost:9091").
			WithLogger(logger)

		if builder.logger != logger {
			t.Error("expected custom logger to be set")
		}
	})
}

func TestSubscriberBuilder_Build(t *testing.T) {
	t.Run("empty server address", func(t *testing.T) {
		builder := NewSubscriberBuilder("")
		sub, err := builder.Build()

		if err == nil {
			t.Fatal("expected error for empty server address")
		}
		if sub != nil {
			t.Error("expected nil subscriber")
		}
	})

	t.Run("with error set", func(t *testing.T) {
		builder := NewSubscriberBuilder("localhost:9091")
		builder.err = context.Canceled

		sub, err := builder.Build()
		if err == nil {
			t.Fatal("expected error when builder has error")
		}
		if sub != nil {
			t.Error("expected nil subscriber")
		}
	})

	t.Run("empty durable name uses default", func(t *testing.T) {
		builder := NewSubscriberBuilder("localhost:9091")
		builder.durableName = ""

		// Can't build without server, but verify default is set during build attempt
		_, _ = builder.Build()

		// The Build method should set default durable name
		if builder.durableName != "default-subscriber" {
			t.Errorf("expected default durable name, got %s", builder.durableName)
		}
	})

	// Note: Can't test successful build without running server
}

func TestSubscriberBuilder_FullChain(t *testing.T) {
	t.Run("complete builder chain", func(t *testing.T) {
		logger := &testSubLogger{}

		builder := NewSubscriberBuilder("localhost:9091").
			WithDurableName("my-consumer").
			WithBatchSize(20).
			WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials())).
			WithLogger(logger)

		if builder.serverAddr != "localhost:9091" {
			t.Errorf("expected server addr 'localhost:9091', got %s", builder.serverAddr)
		}
		if builder.durableName != "my-consumer" {
			t.Errorf("expected durable name 'my-consumer', got %s", builder.durableName)
		}
		if builder.batchSize != 20 {
			t.Errorf("expected batch size 20, got %d", builder.batchSize)
		}
		if len(builder.dialOpts) != 1 {
			t.Errorf("expected 1 dial option, got %d", len(builder.dialOpts))
		}
		if builder.logger != logger {
			t.Error("expected custom logger to be set")
		}
	})
}
