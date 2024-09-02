package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	var err error
	log, err = config.Build()
	if err != nil {
		panic(err)
	}
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