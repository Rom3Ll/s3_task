package logger

import (
	"os"

	"go.uber.org/zap"
)

var (
	Log *zap.Logger
)

func Init() {
	var err error

	config := zap.NewDevelopmentConfig()
	Log, err = config.Build()

	if err != nil {
		panic("Fatal error logger config building: " + err.Error())
	}

	hostname, _ := os.Hostname()
	Log = Log.With(
		zap.String("service", "api-gateway"),
		zap.String("host", hostname),
	)
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	}
}

func With(fields ...zap.Field) *zap.Logger {
	if Log != nil {
		return Log.With(fields...)
	}
	return nil
}