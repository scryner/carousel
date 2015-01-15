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
	VERSION  = "0.2"

	_DEFAULT_PORT      = 3999
	_DEFAULT_LOG_LEVEL = logg.LOG_LEVEL_INFO
)

var (
	port          int
	logFile       string
	enableGzip    bool
	verbose       bool
	launchAtStart bool

	playEnabled      bool
	remotePlayground bool

	logger *logg.Logger
)

func init() {
	flag.IntVar(&port, "p", _DEFAULT_PORT, "listen port")
	flag.StringVar(&logFile, "log", "stderr", "specify log file (stdout/stderr means standard io)")
	flag.BoolVar(&enableGzip, "z", true, "whether gzip supported or not")
	flag.BoolVar(&launchAtStart, "l", false, "launch local web browser immediately")
	flag.BoolVar(&verbose, "V", false, "logging verbosely")
	flag.BoolVar(&playEnabled, "P", false, "enable go playground")
	flag.BoolVar(&remotePlayground, "R", false, "go playground via Go official site")

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
	staticFiles["/static/slides.js"] = server.StaticContent{"text/javascript", static.Slides_js}
	staticFiles["/static/print.css"] = server.StaticContent{"text/css", static.Print_css}
	staticFiles["/static/styles.css"] = server.StaticContent{"text/css", static.Styles_css}

	if playEnabled {
		logger.Infof("Go playground enabled")

		if remotePlayground {
			logger.Infof("\t: to Go official playground by HTTP")
			staticFiles["/static/play.js"] = server.StaticContent{"text/javascript", static.Play_js + "\ninitPlayground(new HTTPTransport());\n"}
		} else {
			logger.Infof("\t: to local playground by WebSocket")
			staticFiles["/static/play.js"] = server.StaticContent{"text/javascript", static.Play_js + "\ninitPlayground(new SocketTransport());\n"}
		}
	} else {
		logger.Infof("Go playground disabled")
	}

	workingPath := path.Dir(inputFile)

	var rend renderer.Renderer
	rend = renderer.NewFileRenderer(inputFile, playEnabled)

	// initializing server
	srv := server.NewServer(port, enableGzip, workingPath, rend, staticFiles)

	// trying to launch web browser
	if launchAtStart {
		go tryLaunchWebBrowser()
	}

    // print logo
    fmt.Printf(asciiLogo, VERSION);

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

// Actually below codes are not needed any more
func getSocketAddr() string {
	_, localIp, err := getHostnameAndLocalIpAddress()
	if err != nil {
		return fmt.Sprintf("ws://localhost:%d/socket", port)
	}

	return fmt.Sprintf("ws://%s:%d/socket", localIp, port)
}

func getHostnameAndLocalIpAddress() (hostname, localIp string, err error) {
	hostname, err = os.Hostname()
	if err != nil {
		return
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return
	}

	for _, ip := range ips {
		ip4 := ip.To4()
		if ip4 != nil {
			localIp = ip.String()
			break
		}
	}

	return
}
