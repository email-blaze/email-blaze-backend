package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init(level, mode, output string) error {
	var config zap.Config
	if mode == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return err
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

	// Configure output
	switch output {
	case "console":
		config.OutputPaths = []string{"stdout"}
	case "file":
		config.OutputPaths = []string{"logs/app.log"}
	default:
		return fmt.Errorf("invalid output option: %s", output)
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	log = logger
	return nil
}

func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

func Err(err error) zap.Field {
	return zap.Error(err)
}

func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}

func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func FieldInt(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func FieldString(key string, value string) zap.Field {
	return zap.String(key, value)
}

func FieldBool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}