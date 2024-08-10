package echohttpserver

import (
	"cmp"
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gopherd/core/component"

	ginapi "github.com/gopherd/components/httpserver/gin/api"
)

// Name is the unique identifier for the httpserver component.
const Name = "github.com/gopherd/components/httpserver/gin"

// Options defines the configuration options for the httpserver component.
type Options struct {
	Addr  string // Addr is the address to listen on.
	Block bool   // Block indicates whether the Start method should block.
}

// Ensure httpserverComponent implements httpserverapi.Component interface.
var _ ginapi.Component = (*ginComponent)(nil)

func init() {
	component.Register(Name, func() component.Component {
		return &ginComponent{}
	})
}

type ginComponent struct {
	component.BaseComponent[Options]
	engine *gin.Engine
	server *http.Server
}

func (com *ginComponent) Init(ctx context.Context) error {
	com.engine = gin.Default()
	return nil
}

func (com *ginComponent) Start(ctx context.Context) error {
	addr := cmp.Or(com.Options().Addr, ":http")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	com.server = &http.Server{
		Addr:    addr,
		Handler: com.engine,
	}
	if com.Options().Block {
		com.Logger().Info("http server started", "addr", addr)
		return com.server.Serve(ln)
	}
	go func() {
		com.Logger().Info("http server started", "addr", addr)
		com.server.Serve(ln)
	}()
	return nil
}

func (com *ginComponent) Shutdown(ctx context.Context) error {
	if com.server == nil {
		return com.server.Shutdown(ctx)
	}
	return nil
}

func (com *ginComponent) Handle(methods []string, path string, h http.Handler) {
	if len(methods) > 0 {
		com.engine.Match(methods, path, gin.WrapH(h))
	} else {
		com.engine.Any(path, gin.WrapH(h))
	}
}

func (com *ginComponent) HandleFunc(methods []string, path string, h http.HandlerFunc) {
	com.Handle(methods, path, h)
}

func (com *ginComponent) Engine() *gin.Engine {
	return com.engine
}
