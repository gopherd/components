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
	"reflect"
	"strings"
	"time"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
	"github.com/gopherd/core/operator"

	httpapi "github.com/gopherd/components/httpserver/http/api"
	loggerapi "github.com/gopherd/components/logger/api"
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

	// HTTPPath specifies the root HTTP path to get/set log level.
	// If empty, the HTTP handler is not registered.
	//
	// - get log level: GET {HTTPPath}/get
	// - set log level: POST {HTTPPath}/set?level={level} where level is one of DEBUG, INFO, WARN, ERROR or a number
	HTTPPath string
}

func (options *Options) setDefaults() {
	operator.SetDefault(&options.Output, "stderr")
	operator.SetDefault(&options.Level, slog.LevelInfo)
	operator.SetDefault(&options.TimeFormat, "H")
	operator.SetDefault(&options.SourceFormat, "S")
	operator.SetDefault(&options.LevelFormat, "L")
}

func init() {
	component.Register(Name, func() component.Component {
		return &loggerComponent{}
	})
}

type loggerComponent struct {
	component.BaseComponentWithRefs[Options, struct {
		HTTPServer      component.OptionalReference[httpapi.Component]
		EventDispatcher component.OptionalReference[event.Dispatcher[reflect.Type]]
	}]
	writer io.Writer
	closer io.Closer
	level  slog.LevelVar
}

// Init initializes the logger component.
func (com *loggerComponent) Init(ctx context.Context) error {
	com.Options().setDefaults()
	if err := com.initWriter(); err != nil {
		return err
	}
	com.level.Set(com.Options().Level)

	opts := &slog.HandlerOptions{
		Level:     &com.level,
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

func (com *loggerComponent) Start(ctx context.Context) error {
	if server := com.Refs().HTTPServer.Component(); server != nil {
		if root := com.Options().HTTPPath; root != "" {
			com.Logger().Info(
				"register HTTP handler",
				"get", path.Join(root, "/get"),
				"set", path.Join(root, "/set"),
			)
			server.HandleFunc([]string{http.MethodGet}, path.Join(root, "/get"), com.handleGetLogLevel)
			server.HandleFunc([]string{http.MethodPost}, path.Join(root, "/set"), com.handleSetLogLevel)
		}
	}
	if dispatcher := com.Refs().EventDispatcher.Component(); dispatcher != nil {
		com.Logger().Info("register event listener")
		dispatcher.AddListener(loggerapi.SetLevelEventListener(com.onSetLevelEvent))
	}
	return nil
}

// Uninit implements the component.Component interface.
func (com *loggerComponent) Uninit(ctx context.Context) error {
	if com.closer != nil {
		return com.closer.Close()
	}
	return nil
}

// handleGetLogLevel handles the HTTP request to get log level.
func (com *loggerComponent) handleGetLogLevel(w http.ResponseWriter, r *http.Request) {
	level := com.level.Level()
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(level.String()))
}

// handleSetLogLevel handles the HTTP request to set log level.
func (com *loggerComponent) handleSetLogLevel(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	if level == "" {
		http.Error(w, "missing level", http.StatusBadRequest)
		return
	}
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		http.Error(w, "invalid level", http.StatusBadRequest)
		return
	}
	com.level.Set(l)
	com.Logger().Log(context.Background(), l, "set log level", "level", l)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(l.String()))
}

// onSetLevelEvent handles the SetLevelEvent event to set log level.
func (com *loggerComponent) onSetLevelEvent(ctx context.Context, e *loggerapi.SetLevelEvent) error {
	com.level.Set(e.Level)
	com.Logger().Log(context.Background(), e.Level, "set log level", "level", e.Level)
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
