package logger

import "testing"

import "github.com/sirupsen/logrus"

func TestLogLevel(t *testing.T) {
	tests := map[string]logrus.Level{
		"":      logrus.InfoLevel,
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"error": logrus.ErrorLevel,
		"fatal": logrus.FatalLevel,
	}
	config := &Config{}
	for level, expected := range tests {
		config.Level = level
		err := Setup(config)
		if err != nil {
			t.Fatalf("error setting logging level %v", err)
		}
		if Logger.Level != expected {
			t.Fatalf("invalid logging level. expected %v got %v", expected, Logger.Level)
		}
	}
}
