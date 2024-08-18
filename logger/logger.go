// Package logger provides a customizable logging component for applications.
package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/op"

	"github.com/gopherd/components/httpserver/http/httpapi"
)

// Name is the unique identifier for the logger component.
const Name = "github.com/gopherd/components/logger"

// Options defines the configuration options for the logger component.
type Options struct {
	// Output specifies where the log is written to.
	// Supported values:
	//   - "stderr": standard error output
	//   - "stdout": standard output
	//   - "": discard logs
	Output string

	// Level sets the minimum log level.
	Level slog.Level

	// JSON determines whether to output logs in JSON format.
	JSON bool

	// TimeFormat specifies the time format for log entries.
	// Supported values:
	//   - "h": "2006-01-02 15:04:05" in local time
	//   - "H": "2006-01-02 15:04:05.000000" in local time
	//   - "u": "2006-01-02 15:04:05" in UTC
	//   - "U": "2006-01-02 15:04:05.000000" in UTC
	//   - "t": RFC3339 format
	//   - "T": RFC3339Nano format
	//   - "s": Unix seconds
	//   - "S": Unix nanoseconds
	//   - "": omit time
	//   - other: custom time format
	TimeFormat string

	// SourceFormat specifies the source code location format.
	// Supported values:
	//   - "s": file:line
	//   - "S": pkg/file:line
	//   - "n": pkg.func/file:line
	//   - "": omit source
	SourceFormat string

	// LevelFormat specifies the log level format.
	// Supported values:
	//   - "l": one letter (D, I, W, E)
	//   - "L": full word (DEBUG, INFO, WARN, ERROR)
	//   - "": omit level
	LevelFormat string

	// HTTPPath specifies the root HTTP path to get/set log level.
	// If empty, the HTTP handler is not registered.
	//
	// - get log level: GET {HTTPPath}/get
	// - set log level: POST {HTTPPath}/set?level={level} where level is one of DEBUG, INFO, WARN, ERROR or a number
	HTTPPath string
}

func (options *Options) OnLoaded() error {
	op.SetOr(&options.Output, "stderr")
	op.SetOr(&options.Level, slog.LevelInfo)
	op.SetOr(&options.TimeFormat, "H")
	op.SetOr(&options.SourceFormat, "S")
	op.SetOr(&options.LevelFormat, "L")
	return nil
}

func init() {
	component.Register(Name, func() component.Component {
		return &loggerComponent{}
	})
}

type loggerComponent struct {
	component.BaseComponentWithRefs[Options, struct {
		HTTPServer component.OptionalReference[httpapi.Component]
	}]
	writer io.Writer
	closer io.Closer
	level  slog.LevelVar
}

// Init initializes the logger component.
func (c *loggerComponent) Init(ctx context.Context) error {
	if err := c.initWriter(); err != nil {
		return err
	}
	c.level.Set(c.Options().Level)

	opts := &slog.HandlerOptions{
		Level:     &c.level,
		AddSource: c.Options().SourceFormat != "",
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if groups != nil {
				return a
			}
			switch a.Key {
			case slog.TimeKey:
				return c.formatTime(a)
			case slog.SourceKey:
				return c.formatSource(a)
			case slog.LevelKey:
				return c.formatLevel(a)
			}
			return a
		},
	}
	var handler slog.Handler
	if c.Options().JSON {
		handler = slog.NewJSONHandler(c.writer, opts)
	} else {
		handler = slog.NewTextHandler(c.writer, opts)
	}
	slog.SetDefault(slog.New(handler))
	return nil
}

// Start implements the component.Component interface.
func (c *loggerComponent) Start(ctx context.Context) error {
	if server := c.Refs().HTTPServer.Component(); server != nil {
		if root := c.Options().HTTPPath; root != "" {
			c.Logger().Info(
				"register HTTP handler",
				"get", path.Join(root, "/get"),
				"set", path.Join(root, "/set"),
			)
			server.HandleFunc([]string{http.MethodGet}, path.Join(root, "/get"), c.handleGetLogLevel)
			server.HandleFunc([]string{http.MethodPost}, path.Join(root, "/set"), c.handleSetLogLevel)
		}
	}
	return nil
}

// Uninit implements the component.Component interface.
func (c *loggerComponent) Uninit(ctx context.Context) error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}

// SetLogLevel sets the log level.
func (c *loggerComponent) SetLogLevel(level slog.Level) {
	c.level.Set(level)
}

// GetLogLevel returns the log level.
func (c *loggerComponent) GetLogLevel() slog.Level {
	return c.level.Level()
}

// handleGetLogLevel handles the HTTP request to get log level.
func (c *loggerComponent) handleGetLogLevel(w http.ResponseWriter, r *http.Request) {
	level := c.level.Level()
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(level.String() + "\n"))
}

// handleSetLogLevel handles the HTTP request to set log level.
func (c *loggerComponent) handleSetLogLevel(w http.ResponseWriter, r *http.Request) {
	level := r.FormValue("level")
	if level == "" {
		http.Error(w, "missing level", http.StatusBadRequest)
		return
	}
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		http.Error(w, "invalid level", http.StatusBadRequest)
		return
	}
	c.level.Set(l)
	c.Logger().Log(context.Background(), l, "set log level", "level", l)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(l.String() + "\n"))
}

func (c *loggerComponent) initWriter() error {
	switch c.Options().Output {
	case "stderr":
		c.writer = os.Stderr
	case "stdout":
		c.writer = os.Stdout
	case "":
		c.writer = io.Discard
	default:
		return errors.New("unsupported output")
	}
	return nil
}

func (c *loggerComponent) formatTime(a slog.Attr) slog.Attr {
	switch p := c.Options().TimeFormat; p {
	case "":
		return slog.Attr{}
	case "s":
		return slog.Int64("time", a.Value.Time().Unix())
	case "S":
		return slog.Int64("time", a.Value.Time().UnixNano())
	case "h", "H", "u", "U", "t", "T":
		t := a.Value.Time()
		if p == "u" || p == "U" {
			t = t.UTC()
		}
		format := getTimeFormat(p)
		return slog.String("time", t.Format(format))
	default:
		return slog.String("time", a.Value.Time().Format(p))
	}
}

func (c *loggerComponent) formatSource(a slog.Attr) slog.Attr {
	p := c.Options().SourceFormat
	if p == "" {
		return slog.Attr{}
	}
	source, ok := a.Value.Any().(*slog.Source)
	if !ok {
		return a
	}
	var b strings.Builder
	switch p {
	case "s":
		b.WriteString(filepath.Base(source.File))
	case "S":
		dir, file := filepath.Split(source.File)
		b.WriteString(filepath.Base(dir))
		b.WriteByte('/')
		b.WriteString(file)
	case "n":
		if lastSlash := strings.LastIndex(source.Function, "/"); lastSlash != -1 {
			b.WriteString(source.Function[lastSlash+1:])
		} else {
			b.WriteString(source.Function)
		}
		b.WriteByte('/')
		b.WriteString(filepath.Base(source.File))
	default:
		return a
	}
	b.WriteByte(':')
	itoa(&b, source.Line)
	return slog.String("source", b.String())
}

func (c *loggerComponent) formatLevel(a slog.Attr) slog.Attr {
	switch p := c.Options().LevelFormat; p {
	case "":
		return slog.Attr{}
	case "l":
		if level := a.Value.String(); level != "" {
			return slog.String("level", level[:1])
		}
	}
	return a
}

func getTimeFormat(p string) string {
	switch p {
	case "h", "u":
		return "2006-01-02 15:04:05"
	case "H", "U":
		return "2006-01-02 15:04:05.000000"
	case "t":
		return time.RFC3339
	case "T":
		return time.RFC3339Nano
	default:
		return p
	}
}

// itoa writes the decimal representation of i to b.
func itoa(b *strings.Builder, i int) {
	if i < 0 {
		b.WriteByte('-')
		i = -i
	}

	if i < 10000 {
		writeSmallInt(b, i)
		return
	}

	if i < 100000 {
		b.WriteByte(digits[i/10000])
		writeSmallInt(b, i%10000)
		return
	}

	var buf [16]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = digits[i%10]
		i /= 10
	}
	b.Write(buf[pos:])
}

func writeSmallInt(b *strings.Builder, i int) {
	if i < 10 {
		b.WriteByte(digits[i])
	} else if i < 100 {
		b.WriteByte(digits[i/10])
		b.WriteByte(digits[i%10])
	} else if i < 1000 {
		b.WriteByte(digits[i/100])
		b.WriteByte(digits[(i/10)%10])
		b.WriteByte(digits[i%10])
	} else {
		b.WriteByte(digits[i/1000])
		b.WriteByte(digits[(i/100)%10])
		b.WriteByte(digits[(i/10)%10])
		b.WriteByte(digits[i%10])
	}
}

const digits = "0123456789"
