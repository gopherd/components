package internal

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

	"github.com/gopherd/components/httpserver"
	"github.com/gopherd/components/logger"
)

func init() {
	component.Register(logger.Name, func() component.Component {
		return &LoggerComponent{}
	})
}

type LoggerComponent struct {
	component.BaseComponentWithRefs[logger.Options, struct {
		HTTPServer component.OptionalReference[httpserver.Component]
	}]
	writer io.Writer
	closer io.Closer
	level  slog.LevelVar
}

// Init initializes the logger component.
func (c *LoggerComponent) Init(ctx context.Context) error {
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
func (c *LoggerComponent) Start(ctx context.Context) error {
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
func (c *LoggerComponent) Uninit(ctx context.Context) error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}

// SetLogLevel sets the log level.
func (c *LoggerComponent) SetLogLevel(level slog.Level) {
	c.level.Set(level)
}

// GetLogLevel returns the log level.
func (c *LoggerComponent) GetLogLevel() slog.Level {
	return c.level.Level()
}

// handleGetLogLevel handles the HTTP request to get log level.
func (c *LoggerComponent) handleGetLogLevel(w http.ResponseWriter, r *http.Request) {
	level := c.level.Level()
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(level.String() + "\n"))
}

// handleSetLogLevel handles the HTTP request to set log level.
func (c *LoggerComponent) handleSetLogLevel(w http.ResponseWriter, r *http.Request) {
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

func (c *LoggerComponent) initWriter() error {
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

func (c *LoggerComponent) formatTime(a slog.Attr) slog.Attr {
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

func (c *LoggerComponent) formatSource(a slog.Attr) slog.Attr {
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

func (c *LoggerComponent) formatLevel(a slog.Attr) slog.Attr {
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
