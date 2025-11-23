package minitoolstream_connector

import (
	"github.com/moroshma/MiniToolStreamConnector/minitoolstream_connector/infrastructure/handler"
)

// Re-export handler types and functions for convenience

// Publisher handlers
var (
	NewDataHandler  = handler.NewDataHandler
	NewFileHandler  = handler.NewFileHandler
	NewImageHandler = handler.NewImageHandler
)

// Subscriber handlers
var (
	NewFileSaver      = handler.NewFileSaver
	NewImageProcessor = handler.NewImageProcessor
	NewLoggerHandler  = handler.NewLoggerHandler
)

// Handler configs
type (
	DataHandlerConfig     = handler.DataHandlerConfig
	FileHandlerConfig     = handler.FileHandlerConfig
	ImageHandlerConfig    = handler.ImageHandlerConfig
	FileSaverConfig       = handler.FileSaverConfig
	ImageProcessorConfig  = handler.ImageProcessorConfig
	LoggerHandlerConfig   = handler.LoggerHandlerConfig
)
