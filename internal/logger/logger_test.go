package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Redirect log output for testing
	oldOutput := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Restore original output when done
	defer func() {
		os.Stdout = oldOutput
	}()

	// Test cases
	testCases := []struct {
		name     string
		logFunc  func()
		expected LogLevel
		message  string
	}{
		{
			name: "Debug log",
			logFunc: func() {
				Debug("Debug message", nil)
			},
			expected: DEBUG,
			message:  "Debug message",
		},
		{
			name: "Info log",
			logFunc: func() {
				Info("Info message", nil)
			},
			expected: INFO,
			message:  "Info message",
		},
		{
			name: "Warn log",
			logFunc: func() {
				Warn("Warning message", nil)
			},
			expected: WARN,
			message:  "Warning message",
		},
		{
			name: "Error log",
			logFunc: func() {
				Error("Error message", errors.New("test error"), nil)
			},
			expected: ERROR,
			message:  "Error message",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the log function
			tc.logFunc()

			// Flush the writer
			w.Close()

			// Read the output
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Create a new pipe for the next test
			r, w, _ = os.Pipe()
			os.Stdout = w

			// Parse the JSON log entry
			var entry LogEntry
			err := json.Unmarshal(buf.Bytes(), &entry)

			// Verify the log entry
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, entry.Level)
			assert.Equal(t, tc.message, entry.Message)

			// For error logs, verify the error message
			if tc.expected == ERROR {
				assert.Equal(t, "test error", entry.Error)
			}
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	// Redirect log output for testing
	oldOutput := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Restore original output when done
	defer func() {
		os.Stdout = oldOutput
	}()

	// Test with additional fields
	fields := map[string]interface{}{
		"user_id":  uint(123),
		"path":     "/api/test",
		"trace_id": "abc123",
	}

	// Log with fields
	Info("Info with fields", fields)

	// Flush the writer
	w.Close()

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Parse the JSON log entry
	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)

	// Verify the log entry
	assert.NoError(t, err)
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "Info with fields", entry["message"])
	assert.Equal(t, float64(123), entry["user_id"]) // JSON numbers are float64
	assert.Equal(t, "/api/test", entry["path"])
	assert.Equal(t, "abc123", entry["trace_id"])
}
