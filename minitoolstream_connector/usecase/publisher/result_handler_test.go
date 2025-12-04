package publisher

import (
	"context"
	"testing"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
)

func TestNewLoggingResultHandler(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		logger := &testLogger{}
		handler := NewLoggingResultHandler(logger, true)

		if handler == nil {
			t.Fatal("expected non-nil handler")
		}
		if handler.logger != logger {
			t.Error("expected custom logger")
		}
		if !handler.verbose {
			t.Error("expected verbose to be true")
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		handler := NewLoggingResultHandler(nil, false)

		if handler == nil {
			t.Fatal("expected non-nil handler")
		}
		if handler.logger == nil {
			t.Error("expected default logger")
		}
		if handler.verbose {
			t.Error("expected verbose to be false")
		}
	})
}

func TestLoggingResultHandler_Handle(t *testing.T) {
	t.Run("successful result verbose", func(t *testing.T) {
		logger := &testLogger{}
		handler := NewLoggingResultHandler(logger, true)

		result := &domain.PublishResult{
			Sequence:     42,
			ObjectName:   "test-object",
			StatusCode:   0,
			ErrorMessage: "",
		}

		err := handler.Handle(context.Background(), result)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(logger.messages) == 0 {
			t.Error("expected logger to be called")
		}
	})

	t.Run("successful result non-verbose", func(t *testing.T) {
		logger := &testLogger{}
		handler := NewLoggingResultHandler(logger, false)

		result := &domain.PublishResult{
			Sequence:     42,
			ObjectName:   "test-object",
			StatusCode:   0,
			ErrorMessage: "",
		}

		err := handler.Handle(context.Background(), result)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(logger.messages) == 0 {
			t.Error("expected logger to be called")
		}
	})

	t.Run("error result", func(t *testing.T) {
		logger := &testLogger{}
		handler := NewLoggingResultHandler(logger, true)

		result := &domain.PublishResult{
			Sequence:     0,
			ObjectName:   "",
			StatusCode:   500,
			ErrorMessage: "internal server error",
		}

		err := handler.Handle(context.Background(), result)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(logger.messages) == 0 {
			t.Error("expected logger to be called")
		}
	})

	t.Run("nil result", func(t *testing.T) {
		logger := &testLogger{}
		handler := NewLoggingResultHandler(logger, true)

		err := handler.Handle(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil result")
		}
		if err.Error() != "result cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("various status codes", func(t *testing.T) {
		logger := &testLogger{}
		handler := NewLoggingResultHandler(logger, true)

		statusCodes := []int64{0, 200, 400, 404, 500}
		for _, code := range statusCodes {
			result := &domain.PublishResult{
				Sequence:     1,
				ObjectName:   "object",
				StatusCode:   code,
				ErrorMessage: "error",
			}

			err := handler.Handle(context.Background(), result)
			if err != nil {
				t.Fatalf("expected no error for status code %d, got %v", code, err)
			}
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		logger := &testLogger{}
		handler := NewLoggingResultHandler(logger, true)

		result := &domain.PublishResult{
			Sequence:     1,
			ObjectName:   "object",
			StatusCode:   0,
			ErrorMessage: "",
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Should still work even with cancelled context
		err := handler.Handle(ctx, result)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
