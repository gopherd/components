package gin

import (
	"cmp"
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gopherd/core/component"

	"github.com/gopherd/components/httpserver/gin/ginapi"
)

// Name is the unique identifier for the httpserver component.
const Name = "github.com/gopherd/components/httpserver/gin"

func init() {
	component.Register(Name, func() component.Component {
		return &ginComponent{}
	})
}

// Ensure httpserverComponent implements httpserverapi.Component interface.
var _ ginapi.Component = (*ginComponent)(nil)

// Options defines the configuration options for the httpserver component.
type Options struct {
	Addr  string // Addr is the address to listen on.
	Block bool   // Block indicates whether the Start method should block.
}

type ginComponent struct {
	component.BaseComponent[Options]
	engine *gin.Engine
	server *http.Server
}

func (c *ginComponent) Init(ctx context.Context) error {
	c.engine = gin.Default()
	return nil
}

func (c *ginComponent) Start(ctx context.Context) error {
	addr := cmp.Or(c.Options().Addr, ":http")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	c.server = &http.Server{
		Addr:    addr,
		Handler: c.engine,
	}
	if c.Options().Block {
		c.Logger().Info("http server started", "addr", addr)
		return c.server.Serve(ln)
	}
	go func() {
		c.Logger().Info("http server started", "addr", addr)
		c.server.Serve(ln)
	}()
	return nil
}

func (c *ginComponent) Shutdown(ctx context.Context) error {
	if c.server == nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}

func (c *ginComponent) Handle(methods []string, path string, h http.Handler) {
	if len(methods) > 0 {
		c.engine.Match(methods, path, gin.WrapH(h))
	} else {
		c.engine.Any(path, gin.WrapH(h))
	}
}

func (c *ginComponent) HandleFunc(methods []string, path string, h http.HandlerFunc) {
	c.Handle(methods, path, h)
}

func (c *ginComponent) Engine() *gin.Engine {
	return c.engine
}
