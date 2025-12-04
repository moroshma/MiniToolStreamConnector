package usecase

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
)

type mockEgressClient struct {
	subscribeFunc      func(ctx context.Context, config *domain.SubscriptionConfig) (domain.NotificationStream, error)
	fetchFunc          func(ctx context.Context, config *domain.SubscriptionConfig) (domain.MessageStream, error)
	getLastSequenceFunc func(ctx context.Context, subject string) (uint64, error)
	closeFunc          func() error
}

func (m *mockEgressClient) Subscribe(ctx context.Context, config *domain.SubscriptionConfig) (domain.NotificationStream, error) {
	if m.subscribeFunc != nil {
		return m.subscribeFunc(ctx, config)
	}
	return &mockNotificationStream{}, nil
}

func (m *mockEgressClient) Fetch(ctx context.Context, config *domain.SubscriptionConfig) (domain.MessageStream, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, config)
	}
	return &mockMessageStream{}, nil
}

func (m *mockEgressClient) GetLastSequence(ctx context.Context, subject string) (uint64, error) {
	if m.getLastSequenceFunc != nil {
		return m.getLastSequenceFunc(ctx, subject)
	}
	return 0, nil
}

func (m *mockEgressClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

type mockNotificationStream struct {
	notifications []*domain.Notification
	index         int
	recvFunc      func() (*domain.Notification, error)
}

func (m *mockNotificationStream) Recv() (*domain.Notification, error) {
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

type mockMessageStream struct {
	messages []*domain.ReceivedMessage
	index    int
	recvFunc func() (*domain.ReceivedMessage, error)
}

func (m *mockMessageStream) Recv() (*domain.ReceivedMessage, error) {
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

type testLogger struct {
	messages []string
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	l.messages = append(l.messages, format)
}

func TestNew(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		client := &mockEgressClient{}
		config := &Config{
			Client:      client,
			DurableName: "test-consumer",
			BatchSize:   10,
		}

		sub, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if sub == nil {
			t.Fatal("expected non-nil subscriber")
		}
		if sub.durableName != "test-consumer" {
			t.Errorf("expected durable name 'test-consumer', got %s", sub.durableName)
		}
		if sub.batchSize != 10 {
			t.Errorf("expected batch size 10, got %d", sub.batchSize)
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
		client := &mockEgressClient{}
		config := &Config{
			Client: client,
			Logger: logger,
		}

		sub, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if sub.logger != logger {
			t.Error("expected custom logger")
		}
	})

	t.Run("default batch size", func(t *testing.T) {
		client := &mockEgressClient{}
		config := &Config{
			Client:    client,
			BatchSize: 0,
		}

		sub, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if sub.batchSize != 10 {
			t.Errorf("expected default batch size 10, got %d", sub.batchSize)
		}
	})

	t.Run("negative batch size", func(t *testing.T) {
		client := &mockEgressClient{}
		config := &Config{
			Client:    client,
			BatchSize: -5,
		}

		sub, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if sub.batchSize != 10 {
			t.Errorf("expected default batch size 10, got %d", sub.batchSize)
		}
	})
}

func TestMultiSubject_RegisterHandler(t *testing.T) {
	t.Run("register single handler", func(t *testing.T) {
		client := &mockEgressClient{}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)
		handler := domain.MessageHandlerFunc(func(ctx context.Context, msg *domain.ReceivedMessage) error {
			return nil
		})

		sub.RegisterHandler("test.subject", handler)

		if len(sub.handlers) != 1 {
			t.Errorf("expected 1 handler, got %d", len(sub.handlers))
		}
	})

	t.Run("register multiple handlers", func(t *testing.T) {
		client := &mockEgressClient{}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)
		handler := domain.MessageHandlerFunc(func(ctx context.Context, msg *domain.ReceivedMessage) error {
			return nil
		})

		sub.RegisterHandler("test.subject1", handler)
		sub.RegisterHandler("test.subject2", handler)
		sub.RegisterHandler("test.subject3", handler)

		if len(sub.handlers) != 3 {
			t.Errorf("expected 3 handlers, got %d", len(sub.handlers))
		}
	})
}

func TestMultiSubject_RegisterHandlers(t *testing.T) {
	t.Run("register multiple handlers at once", func(t *testing.T) {
		client := &mockEgressClient{}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)
		handler := domain.MessageHandlerFunc(func(ctx context.Context, msg *domain.ReceivedMessage) error {
			return nil
		})

		handlers := map[string]domain.MessageHandler{
			"test.subject1": handler,
			"test.subject2": handler,
		}

		sub.RegisterHandlers(handlers)

		if len(sub.handlers) != 2 {
			t.Errorf("expected 2 handlers, got %d", len(sub.handlers))
		}
	})
}

func TestMultiSubject_Start(t *testing.T) {
	t.Run("no handlers registered", func(t *testing.T) {
		client := &mockEgressClient{}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)

		err := sub.Start()
		if err == nil {
			t.Fatal("expected error when no handlers registered")
		}
		if err.Error() != "no handlers registered" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("successful start", func(t *testing.T) {
		client := &mockEgressClient{
			subscribeFunc: func(ctx context.Context, config *domain.SubscriptionConfig) (domain.NotificationStream, error) {
				return &mockNotificationStream{
					notifications: []*domain.Notification{
						{Subject: "test.subject", Sequence: 1},
					},
				}, nil
			},
			fetchFunc: func(ctx context.Context, config *domain.SubscriptionConfig) (domain.MessageStream, error) {
				return &mockMessageStream{
					messages: []*domain.ReceivedMessage{
						{Subject: "test.subject", Sequence: 1, Data: []byte("test")},
					},
				}, nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)
		messageReceived := false
		handler := domain.MessageHandlerFunc(func(ctx context.Context, msg *domain.ReceivedMessage) error {
			messageReceived = true
			return nil
		})

		sub.RegisterHandler("test.subject", handler)

		err := sub.Start()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Give it some time to process
		time.Sleep(100 * time.Millisecond)

		sub.Stop()

		if messageReceived {
			// Message processing happened
		}
	})
}

func TestMultiSubject_Stop(t *testing.T) {
	t.Run("stop subscriber", func(t *testing.T) {
		closed := false
		client := &mockEgressClient{
			subscribeFunc: func(ctx context.Context, config *domain.SubscriptionConfig) (domain.NotificationStream, error) {
				return &mockNotificationStream{}, nil
			},
			closeFunc: func() error {
				closed = true
				return nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)
		handler := domain.MessageHandlerFunc(func(ctx context.Context, msg *domain.ReceivedMessage) error {
			return nil
		})

		sub.RegisterHandler("test.subject", handler)
		sub.Start()

		time.Sleep(50 * time.Millisecond)
		sub.Stop()

		if !closed {
			t.Error("client was not closed")
		}
	})
}

func TestMultiSubject_ProcessNotification(t *testing.T) {
	t.Run("process notification successfully", func(t *testing.T) {
		messageHandled := false
		client := &mockEgressClient{
			fetchFunc: func(ctx context.Context, config *domain.SubscriptionConfig) (domain.MessageStream, error) {
				return &mockMessageStream{
					messages: []*domain.ReceivedMessage{
						{Subject: "test.subject", Sequence: 1, Data: []byte("test")},
					},
				}, nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)
		handler := domain.MessageHandlerFunc(func(ctx context.Context, msg *domain.ReceivedMessage) error {
			messageHandled = true
			return nil
		})

		notification := &domain.Notification{
			Subject:  "test.subject",
			Sequence: 1,
		}

		err := sub.processNotification("test.subject", notification, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !messageHandled {
			t.Error("message was not handled")
		}
	})

	t.Run("handler returns error", func(t *testing.T) {
		client := &mockEgressClient{
			fetchFunc: func(ctx context.Context, config *domain.SubscriptionConfig) (domain.MessageStream, error) {
				return &mockMessageStream{
					messages: []*domain.ReceivedMessage{
						{Subject: "test.subject", Sequence: 1, Data: []byte("test")},
					},
				}, nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)
		handler := domain.MessageHandlerFunc(func(ctx context.Context, msg *domain.ReceivedMessage) error {
			return errors.New("handler error")
		})

		notification := &domain.Notification{
			Subject:  "test.subject",
			Sequence: 1,
		}

		// Should not return error even if handler fails
		err := sub.processNotification("test.subject", notification, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("fetch error", func(t *testing.T) {
		expectedErr := errors.New("fetch error")
		client := &mockEgressClient{
			fetchFunc: func(ctx context.Context, config *domain.SubscriptionConfig) (domain.MessageStream, error) {
				return nil, expectedErr
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)
		handler := domain.MessageHandlerFunc(func(ctx context.Context, msg *domain.ReceivedMessage) error {
			return nil
		})

		notification := &domain.Notification{
			Subject:  "test.subject",
			Sequence: 1,
		}

		err := sub.processNotification("test.subject", notification, handler)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMultiSubject_Wait(t *testing.T) {
	t.Run("wait for completion", func(t *testing.T) {
		client := &mockEgressClient{
			subscribeFunc: func(ctx context.Context, config *domain.SubscriptionConfig) (domain.NotificationStream, error) {
				return &mockNotificationStream{}, nil
			},
		}
		config := &Config{
			Client: client,
			Logger: &testLogger{},
		}

		sub, _ := New(config)
		handler := domain.MessageHandlerFunc(func(ctx context.Context, msg *domain.ReceivedMessage) error {
			return nil
		})

		sub.RegisterHandler("test.subject", handler)
		sub.Start()

		// Stop in a goroutine
		go func() {
			time.Sleep(50 * time.Millisecond)
			sub.Stop()
		}()

		// Wait should block until Stop is called
		sub.Wait()
	})
}
