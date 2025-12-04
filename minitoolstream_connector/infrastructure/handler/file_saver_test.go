package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
)

func TestNewFileSaver(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesaver_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("successful creation", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "output")
		logger := &testLogger{}
		config := &FileSaverConfig{
			OutputDir: outputDir,
			Logger:    logger,
		}

		saver, err := NewFileSaver(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if saver == nil {
			t.Fatal("expected non-nil saver")
		}
		if saver.outputDir != outputDir {
			t.Errorf("expected output dir %s, got %s", outputDir, saver.outputDir)
		}

		// Verify directory was created
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			t.Error("output directory was not created")
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "output2")
		config := &FileSaverConfig{
			OutputDir: outputDir,
			Logger:    nil,
		}

		saver, err := NewFileSaver(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if saver.logger == nil {
			t.Error("expected default logger")
		}
	})

	t.Run("existing directory", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "existing")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}

		config := &FileSaverConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		saver, err := NewFileSaver(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if saver == nil {
			t.Fatal("expected non-nil saver")
		}
	})
}

func TestFileSaver_Handle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesaver_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("save message with data", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test1")
		config := &FileSaverConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		saver, err := NewFileSaver(config)
		if err != nil {
			t.Fatalf("failed to create saver: %v", err)
		}

		msg := &domain.ReceivedMessage{
			Subject:   "test.subject",
			Sequence:  42,
			Data:      []byte("test file content"),
			Headers:   map[string]string{"content-type": "text/plain"},
			Timestamp: time.Now(),
		}

		err = saver.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify file was created
		expectedFile := filepath.Join(outputDir, "test.subject_seq_42.txt")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", expectedFile)
		}

		// Verify file content
		content, err := os.ReadFile(expectedFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(content) != "test file content" {
			t.Errorf("expected content 'test file content', got %s", string(content))
		}
	})

	t.Run("save with different content types", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test2")
		config := &FileSaverConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		saver, err := NewFileSaver(config)
		if err != nil {
			t.Fatalf("failed to create saver: %v", err)
		}

		tests := []struct {
			contentType  string
			expectedExt  string
		}{
			{"image/jpeg", ".jpg"},
			{"image/png", ".png"},
			{"application/json", ".json"},
			{"application/pdf", ".pdf"},
			{"text/plain", ".txt"},
			{"application/octet-stream", ".bin"},
		}

		for i, tt := range tests {
			msg := &domain.ReceivedMessage{
				Subject:  "test",
				Sequence: uint64(i),
				Data:     []byte("data"),
				Headers:  map[string]string{"content-type": tt.contentType},
			}

			err = saver.Handle(context.Background(), msg)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			expectedFile := filepath.Join(outputDir, "test_seq_"+string(rune('0'+i))+tt.expectedExt)
			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				t.Errorf("expected file with extension %s was not created", tt.expectedExt)
			}
		}
	})

	t.Run("skip empty data", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test3")
		config := &FileSaverConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		saver, err := NewFileSaver(config)
		if err != nil {
			t.Fatalf("failed to create saver: %v", err)
		}

		msg := &domain.ReceivedMessage{
			Subject:  "test.subject",
			Sequence: 1,
			Data:     []byte{},
			Headers:  map[string]string{},
		}

		err = saver.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// No file should be created
		entries, err := os.ReadDir(outputDir)
		if err != nil {
			t.Fatalf("failed to read directory: %v", err)
		}
		if len(entries) != 0 {
			t.Error("expected no files to be created for empty data")
		}
	})

	t.Run("handle with headers", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test4")
		config := &FileSaverConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		saver, err := NewFileSaver(config)
		if err != nil {
			t.Fatalf("failed to create saver: %v", err)
		}

		msg := &domain.ReceivedMessage{
			Subject:  "test",
			Sequence: 1,
			Data:     []byte("data"),
			Headers: map[string]string{
				"content-type": "text/plain",
				"custom":       "header",
			},
		}

		err = saver.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("no content-type header", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test5")
		config := &FileSaverConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		saver, err := NewFileSaver(config)
		if err != nil {
			t.Fatalf("failed to create saver: %v", err)
		}

		msg := &domain.ReceivedMessage{
			Subject:  "test",
			Sequence: 1,
			Data:     []byte("data"),
			Headers:  map[string]string{},
		}

		err = saver.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// File should be created without extension
		expectedFile := filepath.Join(outputDir, "test_seq_1")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", expectedFile)
		}
	})
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		contentType string
		expectedExt string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"text/plain", ".txt"},
		{"application/json", ".json"},
		{"application/xml", ".xml"},
		{"application/pdf", ".pdf"},
		{"application/octet-stream", ".bin"},
		{"unknown/type", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			ext := getFileExtension(tt.contentType)
			if ext != tt.expectedExt {
				t.Errorf("expected extension %s, got %s", tt.expectedExt, ext)
			}
		})
	}
}
