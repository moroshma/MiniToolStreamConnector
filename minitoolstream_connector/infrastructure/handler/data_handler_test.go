package handler

import (
	"context"
	"testing"
)

type testLogger struct {
	messages []string
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	// Store messages for verification if needed
	l.messages = append(l.messages, format)
}

func TestNewDataHandler(t *testing.T) {
	t.Run("with all config", func(t *testing.T) {
		logger := &testLogger{}
		config := &DataHandlerConfig{
			Subject:     "test.subject",
			Data:        []byte("test data"),
			ContentType: "application/json",
			Headers:     map[string]string{"key": "value"},
			Logger:      logger,
		}

		handler := NewDataHandler(config)
		if handler == nil {
			t.Fatal("expected non-nil handler")
		}
		if handler.subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", handler.subject)
		}
		if string(handler.data) != "test data" {
			t.Errorf("expected data 'test data', got %s", string(handler.data))
		}
		if handler.contentType != "application/json" {
			t.Errorf("expected content type 'application/json', got %s", handler.contentType)
		}
		if handler.logger != logger {
			t.Error("expected custom logger")
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject: "test.subject",
			Data:    []byte("test data"),
			Logger:  nil,
		}

		handler := NewDataHandler(config)
		if handler.logger == nil {
			t.Error("expected default logger")
		}
	})

	t.Run("with nil headers", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject: "test.subject",
			Data:    []byte("test data"),
			Headers: nil,
		}

		handler := NewDataHandler(config)
		if handler.headers == nil {
			t.Error("expected non-nil headers map")
		}
	})

	t.Run("with empty content type", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject:     "test.subject",
			Data:        []byte("test data"),
			ContentType: "",
		}

		handler := NewDataHandler(config)
		if handler.contentType != "application/octet-stream" {
			t.Errorf("expected default content type 'application/octet-stream', got %s", handler.contentType)
		}
	})
}

func TestDataHandler_WithHeaders(t *testing.T) {
	t.Run("add headers", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject: "test.subject",
			Data:    []byte("test data"),
			Headers: map[string]string{"key1": "value1"},
		}

		handler := NewDataHandler(config)
		handler.WithHeaders(map[string]string{
			"key2": "value2",
			"key3": "value3",
		})

		if len(handler.headers) != 3 {
			t.Errorf("expected 3 headers, got %d", len(handler.headers))
		}
		if handler.headers["key1"] != "value1" {
			t.Error("original header should be preserved")
		}
		if handler.headers["key2"] != "value2" {
			t.Error("new header not added")
		}
	})

	t.Run("overwrite header", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject: "test.subject",
			Data:    []byte("test data"),
			Headers: map[string]string{"key": "original"},
		}

		handler := NewDataHandler(config)
		handler.WithHeaders(map[string]string{"key": "updated"})

		if handler.headers["key"] != "updated" {
			t.Errorf("expected header to be updated, got %s", handler.headers["key"])
		}
	})

	t.Run("chain calls", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject: "test.subject",
			Data:    []byte("test data"),
		}

		handler := NewDataHandler(config).
			WithHeaders(map[string]string{"key1": "value1"}).
			WithHeaders(map[string]string{"key2": "value2"})

		if len(handler.headers) != 2 {
			t.Errorf("expected 2 headers, got %d", len(handler.headers))
		}
	})
}

func TestDataHandler_Prepare(t *testing.T) {
	t.Run("successful prepare", func(t *testing.T) {
		logger := &testLogger{}
		config := &DataHandlerConfig{
			Subject:     "test.subject",
			Data:        []byte("test data"),
			ContentType: "application/json",
			Headers:     map[string]string{"custom": "header"},
			Logger:      logger,
		}

		handler := NewDataHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if msg.Subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", msg.Subject)
		}
		if string(msg.Data) != "test data" {
			t.Errorf("expected data 'test data', got %s", string(msg.Data))
		}
		if msg.Headers["content-type"] != "application/json" {
			t.Errorf("expected content-type 'application/json', got %s", msg.Headers["content-type"])
		}
		if msg.Headers["custom"] != "header" {
			t.Error("custom header not included")
		}
		if msg.Headers["timestamp"] == "" {
			t.Error("timestamp header should be set")
		}
	})

	t.Run("prepare with empty data", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject: "test.subject",
			Data:    []byte{},
		}

		handler := NewDataHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(msg.Data) != 0 {
			t.Error("expected empty data")
		}
	})

	t.Run("prepare with nil data", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject: "test.subject",
			Data:    nil,
		}

		handler := NewDataHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if msg.Data != nil {
			t.Error("expected nil data")
		}
	})

	t.Run("headers include timestamp", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject: "test.subject",
			Data:    []byte("test data"),
		}

		handler := NewDataHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, ok := msg.Headers["timestamp"]; !ok {
			t.Error("timestamp header should be present")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		config := &DataHandlerConfig{
			Subject: "test.subject",
			Data:    []byte("test data"),
		}

		handler := NewDataHandler(config)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Prepare should still work even with cancelled context
		// since it doesn't use context internally
		msg, err := handler.Prepare(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if msg == nil {
			t.Error("expected non-nil message")
		}
	})
}
