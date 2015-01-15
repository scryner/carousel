package server

import (
	"carousel/renderer"
	"code.google.com/p/go.tools/playground/socket"
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
	rend        renderer.Renderer

	logger *logg.Logger
}

func NewServer(port int, enableGzip bool, workingPath string, rend renderer.Renderer, staticFiles map[string]StaticContent) *Server {
	logger := logg.GetDefaultLogger("server")

	srv := &Server{
		logger:      logger,
		workingPath: workingPath,
		rend:        rend,
		staticFiles: staticFiles,
	}

	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		defer func() {
			logger.Debugf("%s - \"%s %s\" - \"%s\"", r.Host, r.Method, path, r.UserAgent())
		}()

		switch path {
		case "/":
			srv.handleSlides(w, r)

		case "/refresh":
			srv.handleRefresh(w, r)

		case "/socket":
            wsSrv := socket.NewHandler(r.URL)
			wsSrv.Handler.ServeHTTP(w, r)

		case "/compile":
			srv.handleRedirectToGoPlaygroundAppEngine(w, r)

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
			w.Header().Set("Content-Type", content.Mine+"; charset=utf-8")
			w.Write([]byte(t))
		case []byte:
			w.Header().Set("Content-Type", content.Mine)
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := srv.rend.Render(w)
	if err != nil {
		http.Error(w, fmt.Sprintf("error while rendering: %v", err), http.StatusInternalServerError)
	}
}

func (srv *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	err := srv.rend.Refresh()
	if err != nil {
		http.Error(w, fmt.Sprintf("error while refreshing: %v", err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
