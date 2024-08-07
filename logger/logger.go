// Package logger provides a customizable logging component for applications.
package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopherd/core/component"
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

	// TimeFormat specifies the time pattern for log entries.
	// Supported values:
	//   - "h": "2006-01-02 15:04:05" in local time
	//   - "H": "2006-01-02 15:04:05.999999" in local time
	//   - "u": "2006-01-02 15:04:05" in UTC
	//   - "U": "2006-01-02 15:04:05.999999" in UTC
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
}

// DefaultOptions returns the default logger options.
// The modifier function can be used to customize the default options.
func DefaultOptions(modifier func(*Options)) Options {
	options := Options{
		Output:       "stderr",
		Level:        slog.LevelInfo,
		JSON:         false,
		TimeFormat:   "H",
		SourceFormat: "S",
		LevelFormat:  "L",
	}
	if modifier != nil {
		modifier(&options)
	}
	return options
}

func init() {
	component.Register(Name, func() component.Component {
		return &loggerComponent{}
	})
}

type loggerComponent struct {
	component.BaseComponent[Options]
	writer io.Writer
	closer io.Closer
}

// Init initializes the logger component.
func (com *loggerComponent) Init(ctx context.Context) error {
	if err := com.initWriter(); err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level:     com.Options().Level,
		AddSource: com.Options().SourceFormat != "",
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if groups != nil {
				return a
			}
			switch a.Key {
			case slog.TimeKey:
				return com.formatTime(a)
			case slog.SourceKey:
				return com.formatSource(a)
			case slog.LevelKey:
				return com.formatLevel(a)
			}
			return a
		},
	}
	var handler slog.Handler
	if com.Options().JSON {
		handler = slog.NewJSONHandler(com.writer, opts)
	} else {
		handler = slog.NewTextHandler(com.writer, opts)
	}
	slog.SetDefault(slog.New(handler))
	return nil
}

// Uninit cleans up resources used by the logger component.
func (com *loggerComponent) Uninit(ctx context.Context) error {
	if com.closer != nil {
		return com.closer.Close()
	}
	return nil
}

func (com *loggerComponent) initWriter() error {
	switch com.Options().Output {
	case "stderr":
		com.writer = os.Stderr
	case "stdout":
		com.writer = os.Stdout
	case "":
		com.writer = io.Discard
	default:
		return errors.New("unsupported output")
	}
	return nil
}

func (com *loggerComponent) formatTime(a slog.Attr) slog.Attr {
	switch p := com.Options().TimeFormat; p {
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

func (com *loggerComponent) formatSource(a slog.Attr) slog.Attr {
	p := com.Options().SourceFormat
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

func (com *loggerComponent) formatLevel(a slog.Attr) slog.Attr {
	switch p := com.Options().LevelFormat; p {
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
		return "2006-01-02 15:04:05.999999"
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
