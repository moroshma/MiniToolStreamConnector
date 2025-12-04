package domain

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

func TestMessagePreparerFunc(t *testing.T) {
	t.Run("successful prepare", func(t *testing.T) {
		expectedMsg := &PublishMessage{
			Subject: "test.subject",
			Data:    []byte("test data"),
			Headers: map[string]string{"key": "value"},
		}

		preparer := MessagePreparerFunc(func(ctx context.Context) (*PublishMessage, error) {
			return expectedMsg, nil
		})

		msg, err := preparer.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if msg != expectedMsg {
			t.Errorf("expected message %v, got %v", expectedMsg, msg)
		}
	})

	t.Run("prepare with error", func(t *testing.T) {
		expectedErr := errors.New("prepare error")
		preparer := MessagePreparerFunc(func(ctx context.Context) (*PublishMessage, error) {
			return nil, expectedErr
		})

		msg, err := preparer.Prepare(context.Background())
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if msg != nil {
			t.Errorf("expected nil message, got %v", msg)
		}
	})

	t.Run("prepare with context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		preparer := MessagePreparerFunc(func(ctx context.Context) (*PublishMessage, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return &PublishMessage{}, nil
		})

		_, err := preparer.Prepare(ctx)
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

func TestResultHandlerFunc(t *testing.T) {
	t.Run("successful handle", func(t *testing.T) {
		called := false
		result := &PublishResult{
			Sequence:     1,
			ObjectName:   "test-object",
			StatusCode:   0,
			ErrorMessage: "",
		}

		handler := ResultHandlerFunc(func(ctx context.Context, r *PublishResult) error {
			called = true
			if r != result {
				t.Errorf("expected result %v, got %v", result, r)
			}
			return nil
		})

		err := handler.Handle(context.Background(), result)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !called {
			t.Error("handler was not called")
		}
	})

	t.Run("handle with error", func(t *testing.T) {
		expectedErr := errors.New("handle error")
		handler := ResultHandlerFunc(func(ctx context.Context, r *PublishResult) error {
			return expectedErr
		})

		err := handler.Handle(context.Background(), &PublishResult{})
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestMessageHandlerFunc(t *testing.T) {
	t.Run("successful handle", func(t *testing.T) {
		called := false
		msg := &ReceivedMessage{
			Subject:   "test.subject",
			Sequence:  1,
			Data:      []byte("test data"),
			Headers:   map[string]string{"key": "value"},
			Timestamp: time.Now(),
		}

		handler := MessageHandlerFunc(func(ctx context.Context, m *ReceivedMessage) error {
			called = true
			if m != msg {
				t.Errorf("expected message %v, got %v", msg, m)
			}
			return nil
		})

		err := handler.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !called {
			t.Error("handler was not called")
		}
	})

	t.Run("handle with error", func(t *testing.T) {
		expectedErr := errors.New("handle error")
		handler := MessageHandlerFunc(func(ctx context.Context, m *ReceivedMessage) error {
			return expectedErr
		})

		err := handler.Handle(context.Background(), &ReceivedMessage{})
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestIsEOF(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "EOF error",
			err:      io.EOF,
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEOF(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPublishMessage(t *testing.T) {
	t.Run("create publish message", func(t *testing.T) {
		msg := &PublishMessage{
			Subject: "test.subject",
			Data:    []byte("test data"),
			Headers: map[string]string{
				"content-type": "application/json",
				"key":          "value",
			},
		}

		if msg.Subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", msg.Subject)
		}
		if string(msg.Data) != "test data" {
			t.Errorf("expected data 'test data', got %s", string(msg.Data))
		}
		if len(msg.Headers) != 2 {
			t.Errorf("expected 2 headers, got %d", len(msg.Headers))
		}
	})
}

func TestReceivedMessage(t *testing.T) {
	t.Run("create received message", func(t *testing.T) {
		now := time.Now()
		msg := &ReceivedMessage{
			Subject:   "test.subject",
			Sequence:  42,
			Data:      []byte("test data"),
			Headers:   map[string]string{"key": "value"},
			Timestamp: now,
		}

		if msg.Subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", msg.Subject)
		}
		if msg.Sequence != 42 {
			t.Errorf("expected sequence 42, got %d", msg.Sequence)
		}
		if string(msg.Data) != "test data" {
			t.Errorf("expected data 'test data', got %s", string(msg.Data))
		}
		if !msg.Timestamp.Equal(now) {
			t.Errorf("expected timestamp %v, got %v", now, msg.Timestamp)
		}
	})
}

func TestPublishResult(t *testing.T) {
	t.Run("successful result", func(t *testing.T) {
		result := &PublishResult{
			Sequence:     100,
			ObjectName:   "object-123",
			StatusCode:   0,
			ErrorMessage: "",
		}

		if result.Sequence != 100 {
			t.Errorf("expected sequence 100, got %d", result.Sequence)
		}
		if result.ObjectName != "object-123" {
			t.Errorf("expected object name 'object-123', got %s", result.ObjectName)
		}
		if result.StatusCode != 0 {
			t.Errorf("expected status code 0, got %d", result.StatusCode)
		}
	})

	t.Run("error result", func(t *testing.T) {
		result := &PublishResult{
			Sequence:     0,
			ObjectName:   "",
			StatusCode:   500,
			ErrorMessage: "internal error",
		}

		if result.StatusCode != 500 {
			t.Errorf("expected status code 500, got %d", result.StatusCode)
		}
		if result.ErrorMessage != "internal error" {
			t.Errorf("expected error message 'internal error', got %s", result.ErrorMessage)
		}
	})
}

func TestNotification(t *testing.T) {
	t.Run("create notification", func(t *testing.T) {
		notif := &Notification{
			Subject:  "test.subject",
			Sequence: 123,
		}

		if notif.Subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", notif.Subject)
		}
		if notif.Sequence != 123 {
			t.Errorf("expected sequence 123, got %d", notif.Sequence)
		}
	})
}

func TestSubscriptionConfig(t *testing.T) {
	t.Run("config without start sequence", func(t *testing.T) {
		config := &SubscriptionConfig{
			Subject:       "test.subject",
			DurableName:   "test-consumer",
			StartSequence: nil,
			BatchSize:     10,
		}

		if config.Subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", config.Subject)
		}
		if config.DurableName != "test-consumer" {
			t.Errorf("expected durable name 'test-consumer', got %s", config.DurableName)
		}
		if config.StartSequence != nil {
			t.Errorf("expected nil start sequence, got %v", config.StartSequence)
		}
		if config.BatchSize != 10 {
			t.Errorf("expected batch size 10, got %d", config.BatchSize)
		}
	})

	t.Run("config with start sequence", func(t *testing.T) {
		startSeq := uint64(100)
		config := &SubscriptionConfig{
			Subject:       "test.subject",
			DurableName:   "test-consumer",
			StartSequence: &startSeq,
			BatchSize:     20,
		}

		if config.StartSequence == nil {
			t.Fatal("expected non-nil start sequence")
		}
		if *config.StartSequence != 100 {
			t.Errorf("expected start sequence 100, got %d", *config.StartSequence)
		}
	})
}
