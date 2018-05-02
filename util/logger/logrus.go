package logger

import (
	"fmt"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
)

// Logrus implements Logger
type Logrus struct {
	log *logrus.Logger
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

// SetToDebug sets the logger to DEBUG level
func (l *Logrus) SetToDebug() {
	l.log.SetLevel(logrus.DebugLevel)
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

// Debug logs a message at level Debug on the standard logger
func (l *Logrus) Debug(msg string, keyValues ...interface{}) {
	if err := isValidKeyValues(keyValues); err != nil {
		panic(err)
	}

	l.log.WithFields(l.toFields(keyValues)).Debug(msg)
}

// Info logs a message at level Info on the standard logger
func (l *Logrus) Info(msg string, keyValues ...interface{}) {
	if err := isValidKeyValues(keyValues); err != nil {
		panic(err)
	}

	l.log.WithFields(l.toFields(keyValues)).Info(msg)
}

// Error logs a message at level Error on the standard logger
func (l *Logrus) Error(msg string, keyValues ...interface{}) {
	if err := isValidKeyValues(keyValues); err != nil {
		panic(err)
	}

	l.log.WithFields(l.toFields(keyValues)).Error(msg)
}

// Fatal logs a message at level Fatal on the standard logger
func (l *Logrus) Fatal(msg string, keyValues ...interface{}) {
	if err := isValidKeyValues(keyValues); err != nil {
		panic(err)
	}

	l.log.WithFields(l.toFields(keyValues)).Fatal(msg)
}
