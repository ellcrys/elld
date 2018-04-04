package modules

import (
	"go.uber.org/zap"
)

// NewLogger creates a logger
func NewLogger(moduleName string) *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	log := logger.Sugar().Named(moduleName)
	return log
}

// NewNopLogger creates a logger that does nothing. Useful for test environment
func NewNopLogger() *zap.SugaredLogger {
	logger := zap.NewNop()
	defer logger.Sync()
	log := logger.Sugar()
	return log
}
