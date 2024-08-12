package tcpserver

import (
	"context"
	"net"
	"time"

	"github.com/gopherd/core/component"
)

// Name is the unique identifier for the tcpserver component.
const Name = "github.com/gopherd/components/tcpserver"

// Options defines the configuration options for the tcpserver component.
type Options struct {
	Network      string
	Addr         string
	KeepAlive    int
	ReadTimeout  int
	WriteTimeout int
}

func DefaultOptions(modifier func(*Options)) Options {
	options := Options{
		Network:      "tcp",
		KeepAlive:    300,
		ReadTimeout:  10,
		WriteTimeout: 10,
	}
	if modifier != nil {
		modifier(&options)
	}
	return options
}

func init() {
	component.Register(Name, func() component.Component {
		return &tcpserverComponent{}
	})
}

type tcpserverComponent struct {
	component.BaseComponent[Options]
	listener net.Listener
}

func (c *tcpserverComponent) Start(ctx context.Context) error {
	// listen and serve
	if err := c.listen(); err != nil {
		return err
	}
	go c.serve()
	return nil
}

func (server *tcpserverComponent) Shutdown(ctx context.Context) error {
	return server.listener.Close()
}

// listen creates a tcp server
func (c *tcpserverComponent) listen() error {
	options := c.Options()
	a, err := net.ResolveTCPAddr("tcp", options.Addr)
	if err == nil {
		c.listener, err = net.ListenTCP("tcp", a)
	}
	if err != nil {
		return err
	}
	if options.KeepAlive > 0 {
		if l, ok := c.listener.(*net.TCPListener); ok {
			c.listener = newTCPKeepAliveListener(l, time.Duration(options.KeepAlive)*time.Second)
		} else {
			c.Logger().Warn("keepalive is not supported", "addr", options.Addr)
		}
	}
	return nil
}

func (server *tcpserverComponent) serve() error {
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				println("accept connection error: " + err.Error() + ", retrying")
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0
		var ip string
		if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
			ip = addr.IP.String()
		}
		server.handle(ip, conn)
	}
}

func (server *tcpserverComponent) handle(ip string, conn net.Conn) {
	// TODO: handle connection
}

// tcpKeepAliveListener wraps TCPListener with a keepalive duration
type tcpKeepAliveListener struct {
	*net.TCPListener
	duration time.Duration
}

// newTCPKeepAliveListener creates a TCPKeepAliveListener
func newTCPKeepAliveListener(ln *net.TCPListener, d time.Duration) *tcpKeepAliveListener {
	return &tcpKeepAliveListener{
		TCPListener: ln,
		duration:    d,
	}
}

// Accept implements net.Listener Accept method
func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	if ln.duration == 0 {
		ln.duration = 3 * time.Minute
	}
	tc.SetKeepAlivePeriod(ln.duration)
	return tc, nil
}
