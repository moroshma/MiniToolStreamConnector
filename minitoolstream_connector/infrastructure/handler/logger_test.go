package handler

import (
	"context"
	"testing"
	"time"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
)

func TestNewLoggerHandler(t *testing.T) {
	t.Run("with all config", func(t *testing.T) {
		logger := &testLogger{}
		config := &LoggerHandlerConfig{
			Prefix: "TEST",
			Logger: logger,
		}

		handler := NewLoggerHandler(config)
		if handler == nil {
			t.Fatal("expected non-nil handler")
		}
		if handler.prefix != "TEST" {
			t.Errorf("expected prefix 'TEST', got %s", handler.prefix)
		}
		if handler.logger != logger {
			t.Error("expected custom logger")
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		config := &LoggerHandlerConfig{
			Prefix: "TEST",
			Logger: nil,
		}

		handler := NewLoggerHandler(config)
		if handler.logger == nil {
			t.Error("expected default logger")
		}
	})

	t.Run("with empty prefix", func(t *testing.T) {
		config := &LoggerHandlerConfig{
			Prefix: "",
			Logger: &testLogger{},
		}

		handler := NewLoggerHandler(config)
		if handler.prefix != "" {
			t.Errorf("expected empty prefix, got %s", handler.prefix)
		}
	})
}

func TestLoggerHandler_Handle(t *testing.T) {
	t.Run("log message with data", func(t *testing.T) {
		logger := &testLogger{}
		config := &LoggerHandlerConfig{
			Prefix: "TEST",
			Logger: logger,
		}

		handler := NewLoggerHandler(config)
		msg := &domain.ReceivedMessage{
			Subject:   "test.subject",
			Sequence:  42,
			Data:      []byte("test data"),
			Headers:   map[string]string{"key": "value"},
			Timestamp: time.Now(),
		}

		err := handler.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify logger was called
		if len(logger.messages) == 0 {
			t.Error("expected logger to be called")
		}
	})

	t.Run("log message without headers", func(t *testing.T) {
		logger := &testLogger{}
		config := &LoggerHandlerConfig{
			Prefix: "TEST",
			Logger: logger,
		}

		handler := NewLoggerHandler(config)
		msg := &domain.ReceivedMessage{
			Subject:   "test.subject",
			Sequence:  1,
			Data:      []byte("test data"),
			Headers:   map[string]string{},
			Timestamp: time.Now(),
		}

		err := handler.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("log text message", func(t *testing.T) {
		logger := &testLogger{}
		config := &LoggerHandlerConfig{
			Prefix: "TEXT",
			Logger: logger,
		}

		handler := NewLoggerHandler(config)
		msg := &domain.ReceivedMessage{
			Subject:  "test.subject",
			Sequence: 1,
			Data:     []byte("short text content"),
			Headers:  map[string]string{"content-type": "text/plain"},
		}

		err := handler.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Should log the text content
		if len(logger.messages) < 2 {
			t.Error("expected at least 2 log messages for text content")
		}
	})

	t.Run("log large text message", func(t *testing.T) {
		logger := &testLogger{}
		config := &LoggerHandlerConfig{
			Prefix: "TEXT",
			Logger: logger,
		}

		handler := NewLoggerHandler(config)
		largeText := make([]byte, 2000) // More than 1000 bytes
		for i := range largeText {
			largeText[i] = 'a'
		}

		msg := &domain.ReceivedMessage{
			Subject:  "test.subject",
			Sequence: 1,
			Data:     largeText,
			Headers:  map[string]string{"content-type": "text/plain"},
		}

		err := handler.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Should not log the text content if too large
		// We just verify it doesn't error
	})

	t.Run("log binary message", func(t *testing.T) {
		logger := &testLogger{}
		config := &LoggerHandlerConfig{
			Prefix: "BIN",
			Logger: logger,
		}

		handler := NewLoggerHandler(config)
		msg := &domain.ReceivedMessage{
			Subject:  "test.subject",
			Sequence: 1,
			Data:     []byte{0x00, 0x01, 0x02, 0x03},
			Headers:  map[string]string{"content-type": "application/octet-stream"},
		}

		err := handler.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Should not log binary content
	})

	t.Run("log empty message", func(t *testing.T) {
		logger := &testLogger{}
		config := &LoggerHandlerConfig{
			Prefix: "EMPTY",
			Logger: logger,
		}

		handler := NewLoggerHandler(config)
		msg := &domain.ReceivedMessage{
			Subject:  "test.subject",
			Sequence: 1,
			Data:     []byte{},
			Headers:  map[string]string{},
		}

		err := handler.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		logger := &testLogger{}
		config := &LoggerHandlerConfig{
			Prefix: "TEST",
			Logger: logger,
		}

		handler := NewLoggerHandler(config)
		msg := &domain.ReceivedMessage{
			Subject:  "test.subject",
			Sequence: 1,
			Data:     []byte("test data"),
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Should still work even with cancelled context
		err := handler.Handle(ctx, msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
