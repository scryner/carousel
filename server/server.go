package server

import (
	"carousel/renderer"
	"fmt"
	"github.com/scryner/logg"
	"net/http"
)

type StaticContent struct {
	Mine    string
	Content interface{}
}

type FilePath string

type Server struct {
	gzipHttpServer
	staticFiles map[string]StaticContent
	workingPath string
	rendFunc    renderer.RenderFunc

	logger *logg.Logger
}

func NewServer(port int, enableGzip bool, workingPath string, rendFunc renderer.RenderFunc, staticFiles map[string]StaticContent) *Server {
	logger := logg.GetDefaultLogger("server")

	srv := &Server{
		logger:      logger,
		workingPath: workingPath,
		rendFunc:    rendFunc,
		staticFiles: staticFiles,
	}

	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			logger.Infof("%s - \"%s %s\" - \"%s\"", r.Host, r.Method, r.URL.Path, r.UserAgent())
		}()

		path := r.URL.Path

		switch path {
		case "/":
			srv.handleSlides(w, r)

		default:
			srv.serveStaticFile(w, r, path)
		}
	}

	httpServer := gzipHttpServer{
		port:       port,
		enableGzip: enableGzip,
		serveHTTP:  serveHTTP,
	}

	srv.gzipHttpServer = httpServer

	return srv
}

func (srv *Server) Start() {
	srv.logger.Infof("Starting server on port %d", srv.port)

	err := srv.start()
	if err != nil {
		srv.logger.Fatalf("Failed to starting server: %v", err)
	}
}

func (srv *Server) serveStaticFile(w http.ResponseWriter, r *http.Request, path string) {
	if content, ok := srv.staticFiles[path]; ok {
		switch t := content.Content.(type) {
		case string:
			w.Header().Add("Content-Type", content.Mine+"; charset=utf-8")
			w.Write([]byte(t))
		case []byte:
			w.Header().Add("Content-Type", content.Mine)
			w.Write(t)
		case FilePath:
			http.ServeFile(w, r, string(t))

		default:
			srv.logger.Errorf("unknown static content: %v", t)
			http.Error(w, "unknown static content", http.StatusInternalServerError)
		}
	} else {
		fpath := fmt.Sprintf("%s%s", srv.workingPath, r.URL.Path)
		http.ServeFile(w, r, fpath)
	}
}

func (srv *Server) handleSlides(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	srv.rendFunc(w)
}
