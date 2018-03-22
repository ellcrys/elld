package components

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
