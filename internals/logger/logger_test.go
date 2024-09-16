package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoggerFunctions(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a custom logger for testing
	testEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	testCore := zapcore.NewCore(testEncoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	testLogger := zap.New(testCore)

	// Replace the global logger with our test logger
	originalLogger := log
	log = testLogger
	defer func() { log = originalLogger }()

	tests := []struct {
		name     string
		logFunc  func()
		expected map[string]interface{}
	}{
		{
			name: "Info",
			logFunc: func() {
				Info("test info message", Field("key", "value"))
			},
			expected: map[string]interface{}{
				"level": "info",
				"msg":   "test info message",
				"key":   "value",
			},
		},
		{
			name: "Error",
			logFunc: func() {
				Error("test error message", FieldInt("code", 500))
			},
			expected: map[string]interface{}{
				"level": "error",
				"msg":   "test error message",
				"code":  float64(500), // JSON numbers are floats
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			var result map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &result)
			assert.NoError(t, err)

			for k, v := range tt.expected {
				assert.Equal(t, v, result[k])
			}
		})
	}
}

func TestErrField(t *testing.T) {
	testErr := assert.AnError
	field := Err(testErr)

	assert.Equal(t, zap.Error(testErr), field)
}

func TestFatalDoesNotPanic(t *testing.T) {
	// This test ensures that calling Fatal doesn't panic
	// Note: In a real scenario, Fatal would terminate the program
	// Here we're just making sure it doesn't panic when called
	assert.NotPanics(t, func() {
		Fatal("test fatal message")
	})
}
