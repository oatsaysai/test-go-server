package webserver

import (
	"net"

	"github.com/fasthttp/router"
	"github.com/oklog/run"
	"github.com/valyala/fasthttp/reuseport"

	"github.com/oatsaysai/test-go-server/logger"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	AcceptJSON  = "application/json"
	AcceptRest  = "application/vnd.pgrst.object+json"
	ContentText = "text/plain; charset=utf8"
	ContentRest = "application/vnd.pgrst.object+json; charset=utf-8"
	ContentJSON = "application/json; charset=utf-8"
)

type WebServer struct {
	Config WebConfig
	Addr   string
	Log    *zap.Logger
	ln     net.Listener
	router *router.Router
	debug  bool
}

// NewServer  new fasthttp WebServer
func NewServer(cfg WebConfig) *WebServer {
	log := logger.Console()

	s := &WebServer{
		Config: cfg,
		Addr:   ServerAddr,
		Log:    log,
		router: router.New(),
		debug:  false,
	}
	return s
}

func (ws *WebServer) Close() {
	_ = ws.ln.Close()
}

func (ws *WebServer) Run() (err error) {
	ws.muxRouter()

	// reuse port
	ws.ln, err = reuseport.Listen("tcp4", ws.Addr)
	if err != nil {
		return err
	}
	lg := logger.InitZapLogger(ws.Log)
	s := &fasthttp.Server{
		Handler:            ws.router.Handler,
		Name:               ws.Config.Name,
		ReadBufferSize:     ws.Config.ReadBufferSize,
		MaxConnsPerIP:      ws.Config.MaxConnsPerIP,
		MaxRequestsPerConn: ws.Config.MaxRequestsPerConn,
		MaxRequestBodySize: ws.Config.MaxRequestBodySize,
		Concurrency:        ws.Config.Concurrency,
		Logger:             lg,
	}

	// run fasthttp serv
	var g run.Group
	g.Add(func() error {
		return s.Serve(ws.ln)
	}, func(e error) {
		_ = ws.ln.Close()
	})
	return g.Run()
}
