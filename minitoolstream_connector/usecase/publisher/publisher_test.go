package publisher

import (
	"context"
	"errors"
	"testing"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
)

type mockIngressClient struct {
	publishFunc func(ctx context.Context, msg *domain.PublishMessage) (*domain.PublishResult, error)
	closeFunc   func() error
}

func (m *mockIngressClient) Publish(ctx context.Context, msg *domain.PublishMessage) (*domain.PublishResult, error) {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, msg)
	}
	return &domain.PublishResult{
		Sequence:     1,
		ObjectName:   "test-object",
		StatusCode:   0,
		ErrorMessage: "",
	}, nil
}

func (m *mockIngressClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

type testLogger struct {
	messages []string
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	l.messages = append(l.messages, format)
}

func TestNew(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
		}

		pub, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if pub == nil {
			t.Fatal("expected non-nil publisher")
		}
	})

	t.Run("nil config", func(t *testing.T) {
		_, err := New(nil)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
		if err.Error() != "config cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("nil client", func(t *testing.T) {
		config := &Config{
			Client: nil,
		}

		_, err := New(config)
		if err == nil {
			t.Fatal("expected error for nil client")
		}
		if err.Error() != "client cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("with custom logger", func(t *testing.T) {
		logger := &testLogger{}
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
			Logger: logger,
		}

		pub, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if pub.logger != logger {
			t.Error("expected custom logger")
		}
	})

	t.Run("with custom result handler", func(t *testing.T) {
		client := &mockIngressClient{}
		handler := domain.ResultHandlerFunc(func(ctx context.Context, result *domain.PublishResult) error {
			return nil
		})
		config := &Config{
			Client:        client,
			ResultHandler: handler,
		}

		pub, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if pub.resultHandler == nil {
			t.Error("expected result handler to be set")
		}
	})

	t.Run("default logger and handler", func(t *testing.T) {
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
		}

		pub, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if pub.logger == nil {
			t.Error("expected default logger")
		}
		if pub.resultHandler == nil {
			t.Error("expected default result handler")
		}
	})
}

func TestSimplePublisher_RegisterHandler(t *testing.T) {
	t.Run("register single handler", func(t *testing.T) {
		logger := &testLogger{}
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
			Logger: logger,
		}

		pub, _ := New(config)
		preparer := domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
			return &domain.PublishMessage{}, nil
		})

		pub.RegisterHandler(preparer)

		if len(pub.preparers) != 1 {
			t.Errorf("expected 1 preparer, got %d", len(pub.preparers))
		}
	})

	t.Run("register multiple handlers", func(t *testing.T) {
		logger := &testLogger{}
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
			Logger: logger,
		}

		pub, _ := New(config)
		preparer := domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
			return &domain.PublishMessage{}, nil
		})

		pub.RegisterHandler(preparer)
		pub.RegisterHandler(preparer)
		pub.RegisterHandler(preparer)

		if len(pub.preparers) != 3 {
			t.Errorf("expected 3 preparers, got %d", len(pub.preparers))
		}
	})
}

func TestSimplePublisher_RegisterHandlers(t *testing.T) {
	t.Run("register multiple handlers at once", func(t *testing.T) {
		logger := &testLogger{}
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
			Logger: logger,
		}

		pub, _ := New(config)
		preparers := []domain.MessagePreparer{
			domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
				return &domain.PublishMessage{}, nil
			}),
			domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
				return &domain.PublishMessage{}, nil
			}),
		}

		pub.RegisterHandlers(preparers)

		if len(pub.preparers) != 2 {
			t.Errorf("expected 2 preparers, got %d", len(pub.preparers))
		}
	})
}

func TestSimplePublisher_SetResultHandler(t *testing.T) {
	t.Run("set custom result handler", func(t *testing.T) {
		logger := &testLogger{}
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
			Logger: logger,
		}

		pub, _ := New(config)
		handler := domain.ResultHandlerFunc(func(ctx context.Context, result *domain.PublishResult) error {
			return nil
		})

		pub.SetResultHandler(handler)

		// Just verify it doesn't panic
	})
}

func TestSimplePublisher_Publish(t *testing.T) {
	t.Run("successful publish", func(t *testing.T) {
		client := &mockIngressClient{
			publishFunc: func(ctx context.Context, msg *domain.PublishMessage) (*domain.PublishResult, error) {
				return &domain.PublishResult{
					Sequence:     42,
					ObjectName:   "object-42",
					StatusCode:   0,
					ErrorMessage: "",
				}, nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		preparer := domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
			return &domain.PublishMessage{
				Subject: "test.subject",
				Data:    []byte("test data"),
			}, nil
		})

		err := pub.Publish(context.Background(), preparer)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("preparer error", func(t *testing.T) {
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		expectedErr := errors.New("prepare error")
		preparer := domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
			return nil, expectedErr
		})

		err := pub.Publish(context.Background(), preparer)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("nil message from preparer", func(t *testing.T) {
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		preparer := domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
			return nil, nil
		})

		err := pub.Publish(context.Background(), preparer)
		if err == nil {
			t.Fatal("expected error for nil message")
		}
	})

	t.Run("publish error", func(t *testing.T) {
		expectedErr := errors.New("publish error")
		client := &mockIngressClient{
			publishFunc: func(ctx context.Context, msg *domain.PublishMessage) (*domain.PublishResult, error) {
				return nil, expectedErr
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		preparer := domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
			return &domain.PublishMessage{
				Subject: "test.subject",
				Data:    []byte("test data"),
			}, nil
		})

		err := pub.Publish(context.Background(), preparer)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("server error in result", func(t *testing.T) {
		client := &mockIngressClient{
			publishFunc: func(ctx context.Context, msg *domain.PublishMessage) (*domain.PublishResult, error) {
				return &domain.PublishResult{
					Sequence:     0,
					ObjectName:   "",
					StatusCode:   500,
					ErrorMessage: "server error",
				}, nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		preparer := domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
			return &domain.PublishMessage{
				Subject: "test.subject",
				Data:    []byte("test data"),
			}, nil
		})

		err := pub.Publish(context.Background(), preparer)
		if err == nil {
			t.Fatal("expected error for server error")
		}
	})
}

func TestSimplePublisher_PublishAll(t *testing.T) {
	t.Run("successful publish all", func(t *testing.T) {
		publishCount := 0
		client := &mockIngressClient{
			publishFunc: func(ctx context.Context, msg *domain.PublishMessage) (*domain.PublishResult, error) {
				publishCount++
				return &domain.PublishResult{
					Sequence:     uint64(publishCount),
					ObjectName:   "object",
					StatusCode:   0,
					ErrorMessage: "",
				}, nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		preparers := []domain.MessagePreparer{
			domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
				return &domain.PublishMessage{Subject: "test.subject", Data: []byte("data1")}, nil
			}),
			domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
				return &domain.PublishMessage{Subject: "test.subject", Data: []byte("data2")}, nil
			}),
			domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
				return &domain.PublishMessage{Subject: "test.subject", Data: []byte("data3")}, nil
			}),
		}

		err := pub.PublishAll(context.Background(), preparers)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if publishCount != 3 {
			t.Errorf("expected 3 publishes, got %d", publishCount)
		}
	})

	t.Run("no preparers provided", func(t *testing.T) {
		client := &mockIngressClient{}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)

		err := pub.PublishAll(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for no preparers")
		}
	})

	t.Run("use registered preparers", func(t *testing.T) {
		publishCount := 0
		client := &mockIngressClient{
			publishFunc: func(ctx context.Context, msg *domain.PublishMessage) (*domain.PublishResult, error) {
				publishCount++
				return &domain.PublishResult{StatusCode: 0}, nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		pub.RegisterHandler(domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
			return &domain.PublishMessage{Subject: "test", Data: []byte("data")}, nil
		}))
		pub.RegisterHandler(domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
			return &domain.PublishMessage{Subject: "test", Data: []byte("data")}, nil
		}))

		err := pub.PublishAll(context.Background(), nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if publishCount != 2 {
			t.Errorf("expected 2 publishes, got %d", publishCount)
		}
	})

	t.Run("some preparers fail", func(t *testing.T) {
		client := &mockIngressClient{
			publishFunc: func(ctx context.Context, msg *domain.PublishMessage) (*domain.PublishResult, error) {
				return &domain.PublishResult{StatusCode: 0}, nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		preparers := []domain.MessagePreparer{
			domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
				return &domain.PublishMessage{Subject: "test", Data: []byte("data")}, nil
			}),
			domain.MessagePreparerFunc(func(ctx context.Context) (*domain.PublishMessage, error) {
				return nil, errors.New("prepare error")
			}),
		}

		err := pub.PublishAll(context.Background(), preparers)
		if err == nil {
			t.Fatal("expected error when some preparers fail")
		}
	})
}

func TestSimplePublisher_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		closed := false
		client := &mockIngressClient{
			closeFunc: func() error {
				closed = true
				return nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		err := pub.Close()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !closed {
			t.Error("client was not closed")
		}
	})

	t.Run("close error", func(t *testing.T) {
		expectedErr := errors.New("close error")
		client := &mockIngressClient{
			closeFunc: func() error {
				return expectedErr
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		pub, _ := New(config)
		err := pub.Close()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
