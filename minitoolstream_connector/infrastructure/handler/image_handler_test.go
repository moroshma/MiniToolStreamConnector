package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewImageHandler(t *testing.T) {
	t.Run("with all config", func(t *testing.T) {
		logger := &testLogger{}
		config := &ImageHandlerConfig{
			Subject:   "test.subject",
			ImagePath: "/path/to/image.png",
			Logger:    logger,
		}

		handler := NewImageHandler(config)
		if handler == nil {
			t.Fatal("expected non-nil handler")
		}
		if handler.subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", handler.subject)
		}
		if handler.imagePath != "/path/to/image.png" {
			t.Errorf("expected image path '/path/to/image.png', got %s", handler.imagePath)
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		config := &ImageHandlerConfig{
			Subject:   "test.subject",
			ImagePath: "/path/to/image.png",
			Logger:    nil,
		}

		handler := NewImageHandler(config)
		if handler.logger == nil {
			t.Error("expected default logger")
		}
	})
}

func TestImageHandler_Prepare(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "imagehandler_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("successful prepare - png", func(t *testing.T) {
		// Create a fake PNG file (just test data)
		testFile := filepath.Join(tmpDir, "test.png")
		testData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
		if err := os.WriteFile(testFile, testData, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		config := &ImageHandlerConfig{
			Subject:   "test.subject",
			ImagePath: testFile,
			Logger:    &testLogger{},
		}

		handler := NewImageHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if msg.Subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", msg.Subject)
		}
		if len(msg.Data) != len(testData) {
			t.Errorf("expected data length %d, got %d", len(testData), len(msg.Data))
		}
		if msg.Headers["content-type"] != "image/png" {
			t.Errorf("expected content-type 'image/png', got %s", msg.Headers["content-type"])
		}
		if msg.Headers["filename"] != "test.png" {
			t.Errorf("expected filename 'test.png', got %s", msg.Headers["filename"])
		}
		if msg.Headers["timestamp"] == "" {
			t.Error("timestamp header should be set")
		}
	})

	t.Run("successful prepare - jpeg", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.jpg")
		testData := []byte{0xFF, 0xD8, 0xFF} // JPEG header
		if err := os.WriteFile(testFile, testData, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		config := &ImageHandlerConfig{
			Subject:   "test.subject",
			ImagePath: testFile,
			Logger:    &testLogger{},
		}

		handler := NewImageHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if msg.Headers["content-type"] != "image/jpeg" {
			t.Errorf("expected content-type 'image/jpeg', got %s", msg.Headers["content-type"])
		}
	})

	t.Run("file not found", func(t *testing.T) {
		config := &ImageHandlerConfig{
			Subject:   "test.subject",
			ImagePath: "/nonexistent/image.png",
			Logger:    &testLogger{},
		}

		handler := NewImageHandler(config)
		_, err := handler.Prepare(context.Background())
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})

	t.Run("empty image file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty.png")
		if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		config := &ImageHandlerConfig{
			Subject:   "test.subject",
			ImagePath: testFile,
			Logger:    &testLogger{},
		}

		handler := NewImageHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(msg.Data) != 0 {
			t.Error("expected empty data")
		}
	})
}

func TestDetectImageContentType(t *testing.T) {
	tests := []struct {
		imagePath  string
		expectedCT string
	}{
		{"/path/to/image.png", "image/png"},
		{"/path/to/image.jpg", "image/jpeg"},
		{"/path/to/image.jpeg", "image/jpeg"},
		{"/path/to/image.gif", "image/gif"},
		{"/path/to/image.webp", "image/webp"},
		{"/path/to/image.bmp", "image/bmp"},
		{"/path/to/image.svg", "image/svg+xml"},
		{"/path/to/image.unknown", "image/jpeg"}, // default
		{"/path/to/image", "image/jpeg"},          // default
	}

	for _, tt := range tests {
		t.Run(tt.imagePath, func(t *testing.T) {
			ct := detectImageContentType(tt.imagePath)
			if ct != tt.expectedCT {
				t.Errorf("expected content type %s, got %s", tt.expectedCT, ct)
			}
		})
	}
}
