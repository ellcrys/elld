package logger

import (
	"go.uber.org/zap"
)

// Zap implements Logger
type Zap struct {
	log *zap.SugaredLogger
	cfg *zap.Config
}

// NewZap creates a zap backed logger
func NewZap(dev bool) Logger {

	var log *zap.SugaredLogger

	if !dev {
		cfg := zap.NewProductionConfig()
		logger, _ := cfg.Build()
		defer logger.Sync()
		log = logger.Sugar()
	} else {
		cfg := zap.NewDevelopmentConfig()
		logger, _ := cfg.Build()
		defer logger.Sync()
		log = logger.Sugar()
	}

	l := &Zap{
		log,
		cfg,
	}

	return l
}

// NewZapNoOp creates a logger that logs nothing
func NewZapNoOp() Logger {

	logger := zap.NewNop()
	defer logger.Sync()
	log := logger.Sugar()

	l := &Zap{
		log: log,
	}

	return l
}

// SetToDebug sets the logger to DEBUG level
func (l *Zap) SetToDebug() {
	if l.cfg != nil {
		l.cfg.Level.SetLevel(zap.DebugLevel)
	}
}

// Debug logs a message at level Debug on the standard logger
func (l *Zap) Debug(msg string, keyValues ...interface{}) {
	l.log.Debugw(msg, keyValues...)
}

// Info logs a message at level Info on the standard logger
func (l *Zap) Info(msg string, keyValues ...interface{}) {
	l.log.Infow(msg, keyValues...)
}

// Error logs a message at level Error on the standard logger
func (l *Zap) Error(msg string, keyValues ...interface{}) {
	l.log.Errorw(msg, keyValues...)
}

// Fatal logs a message at level Fatal on the standard logger
func (l *Zap) Fatal(msg string, keyValues ...interface{}) {
	l.log.Errorw(msg, keyValues...)
}
