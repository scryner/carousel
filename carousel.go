package main

import (
	"carousel/renderer"
	"carousel/server"
	"carousel/static"
	"flag"
	"fmt"
	"github.com/scryner/logg"
	"net"
	"os"
	"path"
	"time"
)

const (
	APP_NAME = "Carousel"
	VERSION  = "0.1"

	_DEFAULT_PORT      = 3999
	_DEFAULT_LOG_LEVEL = logg.LOG_LEVEL_INFO
)

var (
	port          int
	logFile       string
	enableGzip    bool
	verbose       bool
	launchAtStart bool

	logger *logg.Logger
)

func init() {
	flag.IntVar(&port, "p", _DEFAULT_PORT, "listen port")
	flag.StringVar(&logFile, "log", "stderr", "specify log file (stdout/stderr means standard io)")
	flag.BoolVar(&enableGzip, "z", true, "whether gzip supported or not")
	flag.BoolVar(&launchAtStart, "l", false, "launch local web browser immediately")
	flag.BoolVar(&verbose, "V", false, "logging verbosely")

	flag.Usage = func() {
		fmt.Printf("%s Version %s\n", APP_NAME, VERSION)
		fmt.Printf("Usage: %s [options] filepath\n", os.Args[0])
		fmt.Println("Options are:")

		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	// getting input file path
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputFile := args[0]

	// initializing logger
	defer func() {
		logg.Flush()
		os.Exit(1)
	}()

	var logLevel logg.LogLevel

	if verbose {
		logLevel = logg.LOG_LEVEL_DEBUG
	} else {
		logLevel = _DEFAULT_LOG_LEVEL
	}

	switch logFile {
	case "stdout":
		logg.SetDefaultLogger(os.Stdout, logLevel)
	case "stderr":
		logg.SetDefaultLogger(os.Stderr, logLevel)
	default:
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logg.SetDefaultLogger(os.Stderr, logLevel)
		} else {
			logg.SetDefaultLogger(f, logLevel)
			defer f.Close()
		}
	}

	logger = logg.GetDefaultLogger("main")

	// initializing static file list
	staticFiles := make(map[string]server.StaticContent)
	staticFiles["/static/play.js"] = server.StaticContent{"text/javascript", static.Play_js}
	staticFiles["/static/slides.js"] = server.StaticContent{"text/javascript", static.Slides_js}
	staticFiles["/static/print.css"] = server.StaticContent{"text/css", static.Print_css}
	staticFiles["/static/styles.css"] = server.StaticContent{"text/css", static.Styles_css}

	workingPath := path.Dir(inputFile)

	var rend renderer.Renderer
	rend = renderer.NewFileRenderer(inputFile)

	// initializing server
	srv := server.NewServer(port, enableGzip, workingPath, rend, staticFiles)

	// trying to launch web browser
	if launchAtStart {
		go tryLaunchWebBrowser()
	}

	// starting server
	srv.Start()
}

func tryLaunchWebBrowser() {
	for {
		c, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			c.Close()
			break
		}
		time.Sleep(time.Millisecond * 500)
	}

	launchWebBrowser()
}
