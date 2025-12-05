package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init(level, format string) {
	Log = logrus.New()

	// Set output
	Log.SetOutput(os.Stdout)

	// Set log level
	switch level {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	// Set format
	if format == "json" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
}

// Info logs an info message
func Info(msg string, fields map[string]interface{}) {
	if fields != nil {
		Log.WithFields(logrus.Fields(fields)).Info(msg)
	} else {
		Log.Info(msg)
	}
}

// Error logs an error message
func Error(msg string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	Log.WithFields(logrus.Fields(fields)).Error(msg)
}

// Warn logs a warning message
func Warn(msg string, fields map[string]interface{}) {
	if fields != nil {
		Log.WithFields(logrus.Fields(fields)).Warn(msg)
	} else {
		Log.Warn(msg)
	}
}

// Debug logs a debug message
func Debug(msg string, fields map[string]interface{}) {
	if fields != nil {
		Log.WithFields(logrus.Fields(fields)).Debug(msg)
	} else {
		Log.Debug(msg)
	}
}

// Audit logs an audit event (always at info level with audit=true field)
func Audit(event string, userID string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["audit"] = true
	fields["event"] = event
	fields["user_id"] = userID
	Log.WithFields(logrus.Fields(fields)).Info("AUDIT")
}
