package pidfile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gopherd/core/component"
)

// Name is the unique identifier for the pidfile component.
const Name = "github.com/gopherd/components/pidfile"

// Options defines the configuration options for the pidfile component.
type Options struct {
	Filename string // Filename is the path to the pid file.
}

func init() {
	component.Register(Name, func() component.Component {
		return &pidfileComponent{}
	})
}

type pidfileComponent struct {
	component.BaseComponent[Options]
	filename string
}

func (com *pidfileComponent) Init(ctx context.Context) error {
	filename := com.Options().Filename
	if filename == "" {
		return nil
	}
	com.filename = filename
	if err := com.createFile(); err != nil {
		return err
	}
	return nil
}

// createFile creates a new pid file. If the pid file exists and the process is running, an error is returned.
func (com *pidfileComponent) createFile() error {
	dir, _ := filepath.Split(com.filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("PidFile: %v", err)
		}
	}
	if content, err := os.ReadFile(com.filename); err == nil {
		pidStr := strings.TrimSpace(string(content))
		if pidStr == "" {
			return nil
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return nil
		}
		if pidIsExist(pid) {
			return fmt.Errorf("pid file found, ensoure %s is not running", os.Args[0])
		}
	}
	f, err := os.OpenFile(com.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%d", os.Getpid())
	if err == nil {
		err = f.Chmod(0444)
	}
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func (com *pidfileComponent) Uninit(ctx context.Context) error {
	return com.removeFile()
}

// removeFile removes the pid file.
func (com *pidfileComponent) removeFile() error {
	if com.filename != "" {
		if err := os.Chmod(com.filename, 0644); err != nil {
			return err
		}
		return os.Remove(com.filename)
	}
	return nil
}
