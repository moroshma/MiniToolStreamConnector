package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/domain"
)

func TestNewImageProcessor(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "imageprocessor_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("successful creation", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "output")
		logger := &testLogger{}
		config := &ImageProcessorConfig{
			OutputDir: outputDir,
			Logger:    logger,
		}

		processor, err := NewImageProcessor(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if processor == nil {
			t.Fatal("expected non-nil processor")
		}
		if processor.outputDir != outputDir {
			t.Errorf("expected output dir %s, got %s", outputDir, processor.outputDir)
		}

		// Verify directory was created
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			t.Error("output directory was not created")
		}
	})

	t.Run("with nil logger", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "output2")
		config := &ImageProcessorConfig{
			OutputDir: outputDir,
			Logger:    nil,
		}

		processor, err := NewImageProcessor(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if processor.logger == nil {
			t.Error("expected default logger")
		}
	})
}

func TestImageProcessor_Handle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "imageprocessor_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("save image with data", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test1")
		config := &ImageProcessorConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		processor, err := NewImageProcessor(config)
		if err != nil {
			t.Fatalf("failed to create processor: %v", err)
		}

		msg := &domain.ReceivedMessage{
			Subject:   "images",
			Sequence:  1,
			Data:      []byte{0x89, 0x50, 0x4E, 0x47}, // PNG header
			Headers:   map[string]string{"content-type": "image/png"},
			Timestamp: time.Now(),
		}

		err = processor.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify file was created
		expectedFile := filepath.Join(outputDir, "images_seq_1.png")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", expectedFile)
		}

		// Verify file content
		content, err := os.ReadFile(expectedFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if len(content) != 4 {
			t.Errorf("expected 4 bytes, got %d", len(content))
		}
	})

	t.Run("save with original filename", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test2")
		config := &ImageProcessorConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		processor, err := NewImageProcessor(config)
		if err != nil {
			t.Fatalf("failed to create processor: %v", err)
		}

		msg := &domain.ReceivedMessage{
			Subject:  "images",
			Sequence: 42,
			Data:     []byte{0xFF, 0xD8, 0xFF}, // JPEG header
			Headers: map[string]string{
				"content-type": "image/jpeg",
				"filename":     "original.jpg",
			},
			Timestamp: time.Now(),
		}

		err = processor.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify file was created with original filename
		expectedFile := filepath.Join(outputDir, "images_seq_42_original.jpg")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", expectedFile)
		}
	})

	t.Run("save with different image types", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test3")
		config := &ImageProcessorConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		processor, err := NewImageProcessor(config)
		if err != nil {
			t.Fatalf("failed to create processor: %v", err)
		}

		tests := []struct {
			contentType string
			expectedExt string
		}{
			{"image/jpeg", ".jpg"},
			{"image/png", ".png"},
			{"image/gif", ".gif"},
			{"image/webp", ".webp"},
			{"image/bmp", ".bmp"},
			{"image/svg+xml", ".svg"},
		}

		for i, tt := range tests {
			msg := &domain.ReceivedMessage{
				Subject:  "images",
				Sequence: uint64(i),
				Data:     []byte("fake image data"),
				Headers:  map[string]string{"content-type": tt.contentType},
			}

			err = processor.Handle(context.Background(), msg)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			expectedFile := filepath.Join(outputDir, "images_seq_"+string(rune('0'+i))+tt.expectedExt)
			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				t.Errorf("expected file with extension %s was not created", tt.expectedExt)
			}
		}
	})

	t.Run("skip empty data", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test4")
		config := &ImageProcessorConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		processor, err := NewImageProcessor(config)
		if err != nil {
			t.Fatalf("failed to create processor: %v", err)
		}

		msg := &domain.ReceivedMessage{
			Subject:  "images",
			Sequence: 1,
			Data:     []byte{},
			Headers:  map[string]string{},
		}

		err = processor.Handle(context.Background(), msg)
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

	t.Run("no content-type header", func(t *testing.T) {
		outputDir := filepath.Join(tmpDir, "test5")
		config := &ImageProcessorConfig{
			OutputDir: outputDir,
			Logger:    &testLogger{},
		}

		processor, err := NewImageProcessor(config)
		if err != nil {
			t.Fatalf("failed to create processor: %v", err)
		}

		msg := &domain.ReceivedMessage{
			Subject:  "images",
			Sequence: 1,
			Data:     []byte("image data"),
			Headers:  map[string]string{},
		}

		err = processor.Handle(context.Background(), msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// File should be created without extension
		expectedFile := filepath.Join(outputDir, "images_seq_1")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", expectedFile)
		}
	})
}

func TestGetImageExtension(t *testing.T) {
	tests := []struct {
		contentType string
		expectedExt string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"image/bmp", ".bmp"},
		{"image/svg+xml", ".svg"},
		{"unknown/type", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			ext := getImageExtension(tt.contentType)
			if ext != tt.expectedExt {
				t.Errorf("expected extension %s, got %s", tt.expectedExt, ext)
			}
		})
	}
}
