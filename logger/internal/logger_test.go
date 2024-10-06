package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/op"
	"github.com/gopherd/core/types"

	"github.com/gopherd/components/logger"
)

type mockEntity struct{}

func (mockEntity) GetComponent(uuid string) component.Component {
	return nil
}

func (mockEntity) Logger() *slog.Logger {
	return slog.Default()
}

func mustNew(t *testing.T, name string, options logger.Options) component.Component {
	t.Helper()
	comp, err := component.Create(name)
	if err != nil {
		t.Fatalf("Failed to create component %q: %v", name, err)
		return nil
	}
	if err := comp.Setup(mockEntity{}, &component.Config{
		Name:    name,
		Options: types.NewRawObject(op.MustResult(json.Marshal(options))),
	}, false); err != nil {
		t.Fatalf("Failed to setup component %q: %v", name, err)
	}

	return comp
}

func TestLoggerComponentInit(t *testing.T) {
	tests := []struct {
		name          string
		options       logger.Options
		expectedError bool
	}{
		{
			name: "Valid stderr output",
			options: logger.Options{
				Output: "stderr",
				Level:  slog.LevelInfo,
			},
			expectedError: false,
		},
		{
			name: "Valid stdout output",
			options: logger.Options{
				Output: "stdout",
				Level:  slog.LevelDebug,
			},
			expectedError: false,
		},
		{
			name: "Valid discard output",
			options: logger.Options{
				Output: "",
				Level:  slog.LevelWarn,
			},
			expectedError: false,
		},
		{
			name: "Invalid output",
			options: logger.Options{
				Output: "invalid",
				Level:  slog.LevelError,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := mustNew(t, logger.Name, tt.options)

			err := comp.Init(context.Background())

			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				err = comp.Uninit(context.Background())
				if err != nil {
					t.Errorf("Unexpected error during Uninit: %v", err)
				}
			}
		})
	}
}

func TestLoggerOutput(t *testing.T) {
	tests := []struct {
		name     string
		options  logger.Options
		logFunc  func(*testing.T, *slog.Logger)
		expected string
	}{
		{
			name: "Info log",
			options: logger.Options{
				Output:       "stdout",
				Level:        slog.LevelInfo,
				JSON:         false,
				TimeFormat:   "",
				SourceFormat: "",
				LevelFormat:  "L",
			},
			logFunc: func(t *testing.T, logger *slog.Logger) {
				logger.Info("test message")
			},
			expected: `level=INFO msg="test message"`,
		},
		{
			name: "Debug log with JSON",
			options: logger.Options{
				Output:       "stdout",
				Level:        slog.LevelDebug,
				JSON:         true,
				TimeFormat:   "s",
				SourceFormat: "s",
				LevelFormat:  "l",
			},
			logFunc: func(t *testing.T, logger *slog.Logger) {
				logger.Debug("debug message", "key", "value")
			},
			expected: `{"time":`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := mustNew(t, logger.Name, tt.options)

			// Redirect stdout to a buffer
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := comp.Init(context.Background())
			if err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}

			tt.logFunc(t, slog.Default())

			// Restore stdout
			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, but got: %s", tt.expected, output)
			}

			err = comp.Uninit(context.Background())
			if err != nil {
				t.Errorf("Unexpected error during Uninit: %v", err)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name       string
		timeFormat string
		expected   string
	}{
		{"Empty", "", ""},
		{"Unix seconds", "s", "time="},
		{"Unix nanoseconds", "S", "time="},
		{"RFC3339", "t", "time="},
		{"RFC3339Nano", "T", "time="},
		{"UTC", "u", "time="},
		{"UTCNano", "U", "time="},
		{"Custom", "2006-01-02", "time="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := mustNew(t, logger.Name, logger.Options{
				Output:     "stdout",
				TimeFormat: tt.timeFormat,
			})

			// Redirect stdout to a buffer
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := comp.Init(context.Background())
			if err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}

			slog.Info("test")

			// Restore stdout
			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)

			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, but got: %s", tt.expected, output)
			}

			err = comp.Uninit(context.Background())
			if err != nil {
				t.Errorf("Unexpected error during Uninit: %v", err)
			}
		})
	}
}

func TestFormatSource(t *testing.T) {
	tests := []struct {
		name         string
		sourceFormat string
		expected     string
	}{
		{"Empty", "", ""},
		{"File and line", "s", "source=logger_test.go:"},
		{"Package, file, and line", "S", "source=internal/logger_test.go:"},
		{"Full", "n", "source=internal.TestFormatSource/logger_test.go:"},
		{"Invalid", "invalid", "/internal/logger_test.go:"},
	}

	for _, tt := range tests {
		comp := mustNew(t, logger.Name, logger.Options{
			Output:       "stdout",
			SourceFormat: tt.sourceFormat,
		})

		// Redirect stdout to a buffer
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := comp.Init(context.Background())
		if err != nil {
			t.Fatalf("Failed to initialize logger: %v", err)
		}

		slog.Info("test")

		// Restore stdout
		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)

		output := buf.String()
		if tt.expected == "" {
			if strings.Contains(output, "source") {
				t.Errorf("Expected no source in output, but got: %s", output)
			}
		} else if !strings.Contains(output, tt.expected) {
			t.Errorf("Expected output to contain %q, but got: %s", tt.expected, output)
		}

		err = comp.Uninit(context.Background())
		if err != nil {
			t.Errorf("Unexpected error during Uninit: %v", err)
		}
	}
}

func TestFormatLevel(t *testing.T) {
	tests := []struct {
		name        string
		levelFormat string
		level       slog.Level
		expected    string
	}{
		{"Empty", "", slog.LevelInfo, ""},
		{"One letter", "l", slog.LevelWarn, "level=W"},
		{"Full word", "L", slog.LevelError, "level=ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := mustNew(t, logger.Name, logger.Options{
				Output:      "stdout",
				LevelFormat: tt.levelFormat,
			})

			// Redirect stdout to a buffer
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := comp.Init(context.Background())
			if err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}

			slog.Log(context.Background(), tt.level, "test")

			// Restore stdout
			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)

			output := buf.String()
			if tt.expected == "" {
				if strings.Contains(output, "level") {
					t.Errorf("Expected no level in output, but got: %s", output)
				}
			} else if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, but got: %s", tt.expected, output)
			}

			err = comp.Uninit(context.Background())
			if err != nil {
				t.Errorf("Unexpected error during Uninit: %v", err)
			}
		})
	}
}

func TestGetTimeFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Local short", "h", "2006-01-02 15:04:05"},
		{"Local long", "H", "2006-01-02 15:04:05.000000"},
		{"UTC short", "u", "2006-01-02 15:04:05"},
		{"UTC long", "U", "2006-01-02 15:04:05.000000"},
		{"RFC3339", "t", time.RFC3339},
		{"RFC3339Nano", "T", time.RFC3339Nano},
		{"Custom", "2006", "2006"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTimeFormat(tt.input)
			if result != tt.expected {
				t.Errorf("getTimeFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"Zero", 0, "0"},
		{"One digit", 2, "2"},
		{"Positive small", 42, "42"},
		{"Three digits", 136, "136"},
		{"Four digits", 1236, "1236"},
		{"Five digits", 12346, "12346"},
		{"Positive large", 123456, "123456"},
		{"Negative", -789, "-789"},
		{"Max int32", 2147483647, "2147483647"},
		{"Min int32", -2147483648, "-2147483648"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			itoa(&b, tt.input)
			if b.String() != tt.expected {
				t.Errorf("itoa(%d) = %q, want %q", tt.input, b.String(), tt.expected)
			}
		})
	}
}

func TestWriteSmallInt(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"Single digit", 5, "5"},
		{"Two digits", 42, "42"},
		{"Three digits", 789, "789"},
		{"Four digits", 1234, "1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeSmallInt(&b, tt.input)
			if b.String() != tt.expected {
				t.Errorf("writeSmallInt(%d) = %q, want %q", tt.input, b.String(), tt.expected)
			}
		})
	}
}
