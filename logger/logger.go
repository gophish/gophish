package logger

import (
	"io"
	"os"

	"github.com/gophish/gophish/config"
	"github.com/sirupsen/logrus"
)

// Logger is the main logger that is abstracted in this package.
// It is exported here for use with gorm.
var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.Formatter = &logrus.TextFormatter{DisableColors: true}
}

// Setup configures the logger based on options in the config.json.
func Setup() error {
	Logger.SetLevel(logrus.InfoLevel)
	// Set up logging to a file if specified in the config
	logFile := config.Conf.LogFile
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		mw := io.MultiWriter(os.Stderr, f)
		Logger.Out = mw
	}
	return nil
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

func Writer() *io.PipeWriter {
	return Logger.Writer()
}
