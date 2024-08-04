package logger

import (
	"encoding/json"

	"github.com/gopherd/core/log"
	"github.com/gopherd/core/operator"
)

// Name represents the name of the component.
const Name = "github.com/gopherd/components/logger"

// Options represents the options of the component.
type Options struct {
	// Name is the name of the log provider which is registered by log.Register.
	Name string `json:"name"`
	// Options used to create the log provider.
	Options json.RawMessage `json:"options"`
}

// DefaultOptions returns the default options.
func DefaultOptions() *Options {
	return &Options{
		Name: "stderr",
		Options: operator.First(json.Marshal(log.StdOptions{
			AddSource: true,
		})),
	}
}
