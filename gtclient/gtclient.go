package gtclient

import (
	"net"
	"os"
	"time"
)

import (
	//l "github.com/ciju/gotunnel/log"
	proto "github.com/ciju/gotunnel/protocol"
	"github.com/ciju/gotunnel/rwtunnel"
	"github.com/op/go-logging"
)

var l = logging.MustGetLogger("client")

func setupHeartbeat(c net.Conn) {

	for {
		time.Sleep(1 * time.Second)
		c.SetWriteDeadline(time.Now().Add(3 * time.Second))

		_, err := c.Write([]byte("ping"))
		if err != nil {
			l.Infof("Couldn't connect to server. Check your network connection, and run client again.")
			os.Exit(1)
		}
	}
}

// connect to server:
// - send the requested sub domain to server.
// - server replies back with a port to setup command channel on.
// - it also replies with the server address that users can access the site on.
func setupCommandChannel(addr, sub string, req, quit chan bool, conn, serverInfo chan string) {
	backProxy, err := net.Dial("tcp", addr)
	if err != nil {
		l.Infof("CMD: Couldn't connect to %s, err: %s", addr, err)
		quit <- true
		return
	}
	defer backProxy.Close()

	proto.SendSubRequest(backProxy, sub)
	l.Infof("send sub domain %s to the remote server", sub)

	// the port to connect on
	servedAt, connTo, _ := proto.ReceiveProxyInfo(backProxy)
	l.Infof("proxy info %s, conn to %s", servedAt, connTo)

	conn <- connTo
	serverInfo <- servedAt

	go setupHeartbeat(backProxy)

	for {
		req <- proto.ReceiveConnRequest(backProxy)
	}
}

func SetupClient(port, remote, subDomain string, serverInfo chan string) bool {
	localServer := net.JoinHostPort("127.0.0.1", port)

	// if !ensureServer(localServer) {
	// 	return false
	// }

	req, quit, conn := make(chan bool), make(chan bool), make(chan string)

	// fmt.Printf("Setting go tunnel server %s with local server on %s\n\n", remote, port)

	go setupCommandChannel(remote, subDomain, req, quit, conn, serverInfo)

	remoteProxy := <-conn

	for {
		select {
		case <-req:
			// fmt.Printf("New link b/w %s and %s\n", remoteProxy, localServer)
			l.Infof("get req, then connect to remote proxy %s", remoteProxy)
			rp, err := net.Dial("tcp", remoteProxy)
			if err != nil {
				l.Infof("Couldn't connect to remote client proxy: %s", err)
				return false
			}
			l.Infof("then connect to local server %v", localServer)
			lp, err := net.Dial("tcp", localServer)
			if err != nil {
				l.Infof("Couldn't connect to local server: %s", err)
				return false
			}

			go rwtunnel.NewRWTunnel(rp, lp)
		case <-quit:
			return true
		}
	}
	return true
}
