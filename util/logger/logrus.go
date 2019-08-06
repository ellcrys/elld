package logger

import (
	"fmt"
	"io/ioutil"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// Logrus implements Logger
type Logrus struct {
	log        *logrus.Logger
	filePath   string
	ns         string
	debugLevel bool
}

// NewLogrus creates a logrus backed logger
func NewLogrus() Logger {
	l := &Logrus{
		log: logrus.New(),
	}

	l.log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	}
	l.log.SetLevel(logrus.InfoLevel)

	return l
}

// NewLogrusWithFileRotation creates a logger
// with file backend and file rotation enabled.
// Two log file are created:
// - filePath.out stores DEBUG and INFO
// - filePath.err stores ERROR
func NewLogrusWithFileRotation(filePath string) Logger {

	l := &Logrus{
		log:      logrus.New(),
		filePath: filePath,
	}

	configureFileRotation(l)

	return l
}

func configureFileRotation(l *Logrus) {
	l.log.Formatter = &logrus.TextFormatter{ForceColors: true}
	l.log.SetLevel(logrus.InfoLevel)

	writer, _ := rotatelogs.New(
		l.filePath+".out.%Y%m%d%H%M",
		rotatelogs.WithLinkName(l.filePath),
		rotatelogs.WithMaxAge(time.Duration(86400)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(604800)*time.Second),
	)

	writerErr, _ := rotatelogs.New(
		l.filePath+".err.%Y%m%d%H%M",
		rotatelogs.WithLinkName(l.filePath),
		rotatelogs.WithMaxAge(time.Duration(86400)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(604800)*time.Second),
	)

	l.log.Hooks.Add(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.InfoLevel:  writer,
			logrus.DebugLevel: writer,
			logrus.ErrorLevel: writerErr,
		},
		&logrus.JSONFormatter{},
	))
}

// NewLogrusNoOp creates a logrus backed logger that logs nothing
func NewLogrusNoOp() Logger {
	l := &Logrus{
		log: logrus.New(),
	}

	l.log.Formatter = &logrus.JSONFormatter{}
	l.log.SetLevel(logrus.PanicLevel)
	l.log.Out = ioutil.Discard
	return l
}

func isValidKeyValues(kv []interface{}) error {
	if len(kv)%2 != 0 {
		return fmt.Errorf("key %v has no value", kv[len(kv)-1])
	}
	return nil
}

// Module derives a new logger from l with a namespace.
// If the current logger has file rotation configured,
// the new logger will also be support file rotation.
// If the current logger is set to debug, then the new
// logger is also set to debug.
func (l *Logrus) Module(ns string) Logger {
	newLog := &Logrus{
		log:      logrus.New(),
		filePath: l.filePath,
		ns:       ns,
	}
	if newLog.filePath != "" {
		configureFileRotation(newLog)
	}
	if l.debugLevel {
		newLog.log.SetLevel(logrus.DebugLevel)
	}
	return newLog
}

// SetToDebug sets the logger to DEBUG level
func (l *Logrus) SetToDebug() {
	l.debugLevel = true
	l.log.SetLevel(logrus.DebugLevel)
}

// SetToInfo sets the logger to INFO level
func (l *Logrus) SetToInfo() {
	l.log.SetLevel(logrus.InfoLevel)
}

func (l *Logrus) toFields(kv []interface{}) (f logrus.Fields) {
	f = logrus.Fields{}
	for i := 0; i < len(kv); i++ {
		if (i + 1) < len(kv) {
			if _v, ok := kv[i].(string); ok {
				f[_v] = kv[i+1]
				i++
			} else {
				panic(fmt.Errorf("string key expected, got %v ", kv))
			}
		}
	}
	return
}

func (l *Logrus) getNamespace() string {
	if l.ns != "" {
		return fmt.Sprintf("[%s]: ", l.ns)
	}
	return ""
}

// Debug logs a message at level Debug on the standard logger
func (l *Logrus) Debug(msg string, keyValues ...interface{}) {
	if err := isValidKeyValues(keyValues); err != nil {
		panic(err)
	}
	l.log.WithFields(l.toFields(keyValues)).Debug(l.getNamespace() + msg)
}

// Info logs a message at level Info on the standard logger
func (l *Logrus) Info(msg string, keyValues ...interface{}) {
	if err := isValidKeyValues(keyValues); err != nil {
		panic(err)
	}

	l.log.WithFields(l.toFields(keyValues)).Info(l.getNamespace() + msg)
}

// Error logs a message at level Error on the standard logger
func (l *Logrus) Error(msg string, keyValues ...interface{}) {
	if err := isValidKeyValues(keyValues); err != nil {
		panic(err)
	}

	l.log.WithFields(l.toFields(keyValues)).Error(l.getNamespace() + msg)
}

// Fatal logs a message at level Fatal on the standard logger
func (l *Logrus) Fatal(msg string, keyValues ...interface{}) {
	if err := isValidKeyValues(keyValues); err != nil {
		panic(err)
	}

	l.log.WithFields(l.toFields(keyValues)).Fatal(l.getNamespace() + msg)
}

// Warn logs a message at level Warn on the standard logger
func (l *Logrus) Warn(msg string, keyValues ...interface{}) {
	if err := isValidKeyValues(keyValues); err != nil {
		panic(err)
	}

	l.log.WithFields(l.toFields(keyValues)).Warn(l.getNamespace() + msg)
}
