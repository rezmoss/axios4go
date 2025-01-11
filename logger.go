package axios4go

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// LogLevel represents different logging level
type LogLevel int

const (
	LevelNone LogLevel = iota
	LevelError
	LevelInfo
	LevelDebug
)

// Logger interface defines methods for logging HTTP requests and responses
type Logger interface {
	LogRequest(*http.Request, LogLevel)
	LogResponse(*http.Response, []byte, time.Duration, LogLevel)
	LogError(error, LogLevel)
	SetLevel(LogLevel)
}

// LogOptions contains configuration for logging
type LogOptions struct {
	Level          LogLevel
	MaxBodyLength  int
	MaskHeaders    []string
	Output         io.Writer
	TimeFormat     string
	IncludeBody    bool
	IncludeHeaders bool
}

// DefaultLogger implements the Logger interface
type DefaultLogger struct {
	options LogOptions
}

// NewDefaultLogger creates a new DefaultLogger with the given options
func NewDefaultLogger(options LogOptions) *DefaultLogger {
	if options.Output == nil {
		options.Output = os.Stdout
	}
	if options.TimeFormat == "" {
		options.TimeFormat = time.RFC3339
	}
	if options.MaxBodyLength == 0 {
		options.MaxBodyLength = 1000
	}
	return &DefaultLogger{options: options}
}

// SetLevel sets the logging level
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.options.Level = level
}

// LogRequest logs HTTP request details
func (l *DefaultLogger) LogRequest(req *http.Request, level LogLevel) {
	if level > l.options.Level {
		return
	}

	var buf strings.Builder
	timestamp := time.Now().Format(l.options.TimeFormat)

	fmt.Fprintf(&buf, "[%s] REQUEST: %s %s\n", timestamp, req.Method, req.URL)

	if l.options.IncludeHeaders {
		buf.WriteString("Headers:\n")
		for key, vals := range req.Header {
			if l.isHeaderMasked(key) {
				fmt.Fprintf(&buf, "  %s: [MASKED]\n", key)
			} else {
				fmt.Fprintf(&buf, "  %s: %s\n", key, strings.Join(vals, ", "))
			}
		}
	}

	if l.options.IncludeBody && req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) > l.options.MaxBodyLength {
				fmt.Fprintf(&buf, "Body: (truncated) %s...\n", body[:l.options.MaxBodyLength])
			} else {
				fmt.Fprintf(&buf, "Body: %s\n", body)
			}
		}
	}

	fmt.Fprintln(l.options.Output, buf.String())
}

// LogResponse logs HTTP response details
func (l *DefaultLogger) LogResponse(resp *http.Response, body []byte, duration time.Duration, level LogLevel) {
	if level > l.options.Level {
		return
	}

	var buf strings.Builder
	timestamp := time.Now().Format(l.options.TimeFormat)

	fmt.Fprintf(&buf, "[%s] RESPONSE: %d %s (%.2fms)\n",
		timestamp, resp.StatusCode, resp.Status, float64(duration.Microseconds())/1000)

	if l.options.IncludeHeaders {
		buf.WriteString("Headers:\n")
		for key, vals := range resp.Header {
			if l.isHeaderMasked(key) {
				fmt.Fprintf(&buf, "  %s: [MASKED]\n", key)
			} else {
				fmt.Fprintf(&buf, "  %s: %s\n", key, strings.Join(vals, ", "))
			}
		}
	}

	if l.options.IncludeBody && body != nil {
		if len(body) > l.options.MaxBodyLength {
			fmt.Fprintf(&buf, "Body: (truncated) %s...\n", body[:l.options.MaxBodyLength])
		} else {
			fmt.Fprintf(&buf, "Body: %s\n", body)
		}
	}

	fmt.Fprintln(l.options.Output, buf.String())
}

// LogError logs error details
func (l *DefaultLogger) LogError(err error, level LogLevel) {
	if level > l.options.Level {
		return
	}

	timestamp := time.Now().Format(l.options.TimeFormat)
	fmt.Fprintf(l.options.Output, "[%s] ERROR: %v\n", timestamp, err)
}

// isHeaderMasked checks if a header should be masked
func (l *DefaultLogger) isHeaderMasked(header string) bool {
	header = strings.ToLower(header)
	for _, masked := range l.options.MaskHeaders {
		if strings.ToLower(masked) == header {
			return true
		}
	}
	return false
}

// Helper function to create a new logger with common defaults
func NewLogger(level LogLevel) Logger {
	return NewDefaultLogger(LogOptions{
		Level:          level,
		MaxBodyLength:  1000,
		MaskHeaders:    []string{"Authorization", "Cookie", "Set-Cookie"},
		Output:         os.Stdout,
		TimeFormat:     time.RFC3339,
		IncludeBody:    true,
		IncludeHeaders: true,
	})
}
