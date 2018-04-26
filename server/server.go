package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

import (
	"github.com/ciju/gotunnel/httpheadreader"
	//l "github.com/ciju/gotunnel/log"
	proto "github.com/ciju/gotunnel/protocol"
	"github.com/ciju/gotunnel/tcprouter"
)

// for isAlive
import (
	"io"
	"time"
	"github.com/op/go-logging"
)

var l = logging.MustGetLogger("server")

// https://groups.google.com/d/topic/golang-nuts/e8sUeulwD3c/discussion
func isAlive(c net.Conn) (ret bool) {
	ret = false

	defer func() {
		if r := recover(); r != nil {
			l.Infof("isAlive: Recovering from f: %s", r)
			ret = false
		}
	}()

	one := make([]byte, 10)
	c.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, err := c.Read(one)
	// l.Infof()("isAlive: read %v - %v", len(one), string(one))
	if err == io.EOF {
		l.Infof("isAlive: %s detected closed LAN connection", c)
		c.Close()
		c = nil
		return
	}
	if n == 0 {
		l.Info("isAlive: read 0 bytes. Probably client out of reach")
		return
	}

	c.SetReadDeadline(time.Time{})
	ret = true
	return
}

func setupClient(externalAddr, port string, adminCtrl net.Conn) {
	host := proto.ReceiveSubRequest(adminCtrl)

	l.Infof("Client: asked for %s, (%s) ", connStr(adminCtrl), host)

	proxy := router.Register(adminCtrl, host)

	requestURL, backendURL := proxy.FrontHost(externalAddr, port), proxy.BackendHost(externalAddr)
	l.Infof("Client: --- sending %v to %v", requestURL, backendURL)

	proto.SendProxyInfo(adminCtrl, requestURL, backendURL)

	for {
		time.Sleep(2 * time.Second)
		if !isAlive(adminCtrl) {
			router.Unregister(proxy)
			break
		}
	}
	l.Infof("Client: closing backend connection")
}

func forwardRequest(conn net.Conn) {
	l.Infof("Request: %s", connStr(conn))
	httpConn := httpheadreader.NewHTTPHeadReader(conn)

	l.Infof("Request: external addr: %s host:  %s", *externalAddr, httpConn.Host())

	if httpConn.Host() == *externalAddr || httpConn.Host() == "www." + *externalAddr {
		conn.Write([]byte(defaultMsg))
		conn.Close()
		return
	}

	// if host is '.m.'+*externalAddr then fwd it to a redis server
	// or something.

	p, ok := router.GetProxy(httpConn.Host())
	if !ok {
		l.Infof("Request: couldn't find proxy for %s", httpConn.Host())
		conn.Write([]byte(fmt.Sprintf("Couldn't fine proxy for <%s>", httpConn.Host())))
		conn.Close()
		return
	}

	l.Info("create new proxy request")
	proto.SendConnRequest(p.Admin)
	p.Proxy.Forward(httpConn)
}

var router = tcprouter.NewTCPRouter(35000, 36000)
var defaultMsg = `
<html>

<body>
    <style>
        body {
            background-color: lightGray;
        }

        h1 {
            margin: 0 auto;
            width: 600px;
            padding: 100px;
            text-align: center;
        }

        a {
            color: #4e4e4e;
            text-decoration: none;
        }

        h3 {
            float: right;
            margin-right: 100px
        }

    </style>
    <h1>
        <a href="http://github.com/ciju/gotunnel">github.com/ciju/gotunnel</a>
    </h1>
    <h3>Sponsored by
        <a href="http://activesphere.com">ActiveSphere</a>
    </h3>
</body>

</html>
`

var (
	port = flag.String("p", "32000", "Access the tunnel sites on this port.")
	// haproxy (or any other supporting WebSocket) can fwd the *80 traffic to the port above.
	externalAddr  = flag.String("a", "localtunnel.net", "the address to be used by the users")
	backProxyAddr = flag.String("x", "0.0.0.0:34000", "Port for clients to connect to")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if *port == "" || *backProxyAddr == "" || *externalAddr == "" {
		flag.Usage()
		os.Exit(1)
	}

	// new clients
	go func() {
		l.Infof("will listen back proxy address: %s", *backProxyAddr)
		backProxy, err := net.Listen("tcp", *backProxyAddr)
		if err != nil {
			l.Fatal("Client: Couldn't start server to connect clients", err)
		}

		for {
			adminCtl, err := backProxy.Accept()
			if err != nil {
				l.Fatal("Client: Problem accepting new client", err)
			}
			l.Infof("accept new conn from back proxy")
			go setupClient(*externalAddr, *port, adminCtl)
		}

	}()

	// new request
	server, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", *port))
	if server == nil {
		l.Fatal("Request: cannot listen: %v", err)
	}
	l.Infof("Listening at: %s", *port)

	for {
		conn, err := server.Accept()
		if err != nil {
			l.Fatal("Request: failed to accept new request: ", err)
		}
		l.Infof("accept new conn")
		go forwardRequest(conn)
	}
}

func connStr(conn net.Conn) string {
	return string(conn.LocalAddr().String()) + " <-> " + string(conn.RemoteAddr().String())
}
