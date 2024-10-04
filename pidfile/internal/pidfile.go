package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gopherd/components/pidfile"
	"github.com/gopherd/core/component"
)

func init() {
	component.Register(pidfile.Name, func() component.Component {
		return &PIDFileComponent{}
	})
}

type PIDFileComponent struct {
	component.BaseComponent[pidfile.Options]
	filename string
}

func (c *PIDFileComponent) Init(ctx context.Context) error {
	filename := c.Options().Filename
	if filename == "" {
		return nil
	}
	c.filename = filename
	if err := c.createFile(); err != nil {
		return err
	}
	return nil
}

// createFile creates a new pid file. If the pid file exists and the process is running, an error is returned.
func (c *PIDFileComponent) createFile() error {
	dir, _ := filepath.Split(c.filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("PidFile: %v", err)
		}
	}
	if content, err := os.ReadFile(c.filename); err == nil {
		pidStr := strings.TrimSpace(string(content))
		if pidStr == "" {
			return nil
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return nil
		}
		if pidExist(pid) {
			return fmt.Errorf("pid file found, ensoure %s is not running", os.Args[0])
		}
	}
	f, err := os.OpenFile(c.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

func (c *PIDFileComponent) Uninit(ctx context.Context) error {
	return c.removeFile()
}

// removeFile removes the pid file.
func (c *PIDFileComponent) removeFile() error {
	if c.filename != "" {
		if err := os.Chmod(c.filename, 0644); err != nil {
			return err
		}
		return os.Remove(c.filename)
	}
	return nil
}
