package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileHandler(t *testing.T) {
	t.Run("with all config", func(t *testing.T) {
		logger := &testLogger{}
		config := &FileHandlerConfig{
			Subject:     "test.subject",
			FilePath:    "/path/to/file.txt",
			ContentType: "text/plain",
			Logger:      logger,
		}

		handler := NewFileHandler(config)
		if handler == nil {
			t.Fatal("expected non-nil handler")
		}
		if handler.subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", handler.subject)
		}
		if handler.filePath != "/path/to/file.txt" {
			t.Errorf("expected file path '/path/to/file.txt', got %s", handler.filePath)
		}
		if handler.contentType != "text/plain" {
			t.Errorf("expected content type 'text/plain', got %s", handler.contentType)
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		config := &FileHandlerConfig{
			Subject:  "test.subject",
			FilePath: "/path/to/file.txt",
			Logger:   nil,
		}

		handler := NewFileHandler(config)
		if handler.logger == nil {
			t.Error("expected default logger")
		}
	})
}

func TestFileHandler_Prepare(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "filehandler_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("successful prepare", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tmpDir, "test.txt")
		testData := []byte("test file content")
		if err := os.WriteFile(testFile, testData, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		config := &FileHandlerConfig{
			Subject:  "test.subject",
			FilePath: testFile,
			Logger:   &testLogger{},
		}

		handler := NewFileHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if msg.Subject != "test.subject" {
			t.Errorf("expected subject 'test.subject', got %s", msg.Subject)
		}
		if string(msg.Data) != "test file content" {
			t.Errorf("expected data 'test file content', got %s", string(msg.Data))
		}
		if msg.Headers["filename"] != "test.txt" {
			t.Errorf("expected filename 'test.txt', got %s", msg.Headers["filename"])
		}
		if msg.Headers["timestamp"] == "" {
			t.Error("timestamp header should be set")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		config := &FileHandlerConfig{
			Subject:  "test.subject",
			FilePath: "/nonexistent/file.txt",
			Logger:   &testLogger{},
		}

		handler := NewFileHandler(config)
		_, err := handler.Prepare(context.Background())
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})

	t.Run("auto detect content type - json", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.json")
		testData := []byte(`{"key": "value"}`)
		if err := os.WriteFile(testFile, testData, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		config := &FileHandlerConfig{
			Subject:  "test.subject",
			FilePath: testFile,
			Logger:   &testLogger{},
		}

		handler := NewFileHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if msg.Headers["content-type"] != "application/json" {
			t.Errorf("expected content-type 'application/json', got %s", msg.Headers["content-type"])
		}
	})

	t.Run("explicit content type overrides detection", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.json")
		testData := []byte(`{"key": "value"}`)
		if err := os.WriteFile(testFile, testData, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		config := &FileHandlerConfig{
			Subject:     "test.subject",
			FilePath:    testFile,
			ContentType: "text/plain",
			Logger:      &testLogger{},
		}

		handler := NewFileHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if msg.Headers["content-type"] != "text/plain" {
			t.Errorf("expected content-type 'text/plain', got %s", msg.Headers["content-type"])
		}
	})

	t.Run("empty file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		config := &FileHandlerConfig{
			Subject:  "test.subject",
			FilePath: testFile,
			Logger:   &testLogger{},
		}

		handler := NewFileHandler(config)
		msg, err := handler.Prepare(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(msg.Data) != 0 {
			t.Error("expected empty data")
		}
	})
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		filePath    string
		expectedCT  string
	}{
		{"/path/to/file.json", "application/json"},
		{"/path/to/file.xml", "application/xml"},
		{"/path/to/file.txt", "text/plain"},
		{"/path/to/file.html", "text/html"},
		{"/path/to/file.pdf", "application/pdf"},
		{"/path/to/file.zip", "application/zip"},
		{"/path/to/file.unknown", "application/octet-stream"},
		{"/path/to/file", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			ct := detectContentType(tt.filePath)
			if ct != tt.expectedCT {
				t.Errorf("expected content type %s, got %s", tt.expectedCT, ct)
			}
		})
	}
}
