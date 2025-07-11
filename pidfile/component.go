// Code generated by "next 0.2.8"; DO NOT EDIT.

package pidfile

import "github.com/gopherd/core/op"

var _ = op.SetDefault[any]

// Name represents the pidfile component name.
const Name = "github.com/gopherd/components/pidfile";

type Options struct {
	// Filename is the path to the pid file.
	Filename string
}

func (x *Options) OnLoaded() {
}
