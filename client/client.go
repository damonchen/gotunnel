package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

import (
	"github.com/ciju/gotunnel/gtclient"
	//l "github.com/ciju/gotunnel/log"
	hs "github.com/ciju/gotunnel/simplehttpserver"
	"github.com/ciju/vercheck"
	"github.com/op/go-logging"
)

var version string
var l = logging.MustGetLogger("client")

var (
	port         = flag.String("p", "", "port")
	subDomain    = flag.String("sub", "", "request subDomain to serve on")
	remote       = flag.String("r", "gotunnel.in:34000", "the remote go tunnel server host/ip:port")
	skipVerCheck = flag.Bool("sc", false, "Skip version check")
	fileServer   = flag.Bool("fs", false, "Server files in the current directory. Use -p to specify the port.")
	serveDir     = flag.String("d", "", "The directory to serve. To be used with -fs.")
	showVersion  = flag.Bool("v", false, "Show version and exit")
)

// var version string

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if *showVersion {
		fmt.Println("Version - ", version)
		return
	}

	if *port == "" || *remote == "" {
		flag.Usage()
		os.Exit(1)
	}

	if !*skipVerCheck {
		if vercheck.HasMinorUpdate(
			"https://raw.github.com/ciju/gotunnel/master/VERSION",
			version,
		) {
			l.Info("\nNew version of go tunnel is available. Please update your code and run again. Or start with option -sc to continue with this version.\n")
			os.Exit(0)
		}
	}

	if *fileServer {
		dir := ""
		// Simple file server.
		if *port == "" {
			fmt.Fprintf(os.Stderr, "-fs needs -p (port) option")
			flag.Usage()
			os.Exit(1)
		}
		if *serveDir == "" {
			dir, _ = os.Getwd()
		} else {
			if path.IsAbs(*serveDir) {
				dir = path.Clean(*serveDir)
			} else {
				wd, _ := os.Getwd()
				dir = path.Clean(path.Join(wd, *serveDir))
			}
		}
		go hs.NewSimpleHTTPServer(*port, dir)
	}

	serverInfo := make(chan string)

	go func() {
		servedAt := <-serverInfo
		fmt.Printf("Your site should be available at: \033[1;34m%s\033[0m\n", servedAt)
	}()

	if !gtclient.SetupClient(*port, *remote, *subDomain, serverInfo) {
		flag.Usage()
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
