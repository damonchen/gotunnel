package tcprouter

import (
	"fmt"

	"net"
	"regexp"
	"github.com/op/go-logging"
	"math/rand"
)

const (
	chars        = "abcdefghiklmnopqrstuvwxyz"
	subDomainLen = 1
)

var l = logging.MustGetLogger("router")

// client request for a string, if already taken, get a new one. else
// use the one asked by client.
func newRandString() string {
	//return "proxy.network.idcos.net"
	var str [subDomainLen]byte
	// rand.Seed(time.Now().Unix())
	for i := 0; i < subDomainLen; i++ {
		num := rand.Intn(len(chars))
		str[i] = chars[num]
	}
	return string(str[:])
}

type TCPRouter struct {
	pool    *PortPool
	proxies map[string]*Proxy
}

func NewTCPRouter(poolStart, poolEnd int) *TCPRouter {
	return &TCPRouter{
		NewPortPool(poolStart, poolEnd),
		make(map[string]*Proxy),
	}
}

func IdForHost(host string) (string, bool) {
	h, _, err := net.SplitHostPort(host)
	if h == "" { // assumes host:port or host as parameter
		h = host
	}

	l.Debugf("id for host: %s", host)

	reg, err := regexp.Compile(`^([A-Za-z]*)`)
	if err != nil {
		l.Errorf("regex compile error %s", err)
		return "", false
	}

	if reg.Match([]byte(host)) {
		l.Infof("Router: %s id for host %s", reg.FindString(h), host)
		return reg.FindString(host), true
	}
	return "", false
}

func (r *TCPRouter) setupClientCommChan(id string) *ProxyClient {
	port, ok := r.pool.GetAvailable()
	if !ok {
		l.Fatal("Couldn't get a port for client communication")
	}

	l.Infof("client conn chan port %v", port)

	proxy, err := NewProxyClient(port)
	if err != nil {
		l.Fatalf("Couldn't setup client communication channel: %s", err)
	}

	return proxy
}

func (r *TCPRouter) Register(adminCtl net.Conn, suggestedId string) (proxy *Proxy) {
	// check if its suggestedId is already registered.
	proxyClient := r.setupClientCommChan(suggestedId)

	id := suggestedId
	for _, ok := r.proxies[id]; ok || id == ""; _, ok = r.proxies[id] {
		id = newRandString()
	}

	l.Infof("Router: registering with (%s)", id)
	r.proxies[id] = &Proxy{Proxy: proxyClient, Admin: adminCtl, id: id}

	return r.proxies[id]
}

func (r *TCPRouter) Unregister(p *Proxy) {
	delete(r.proxies, p.id)
}

func (r *TCPRouter) String() string {
	return fmt.Sprintf("Router: %v", r.proxies)
}

// given a connection, figures out the sub domain and gives respective
// proxy.
func (r *TCPRouter) GetProxy(host string) (*Proxy, bool) {
	id, ok := IdForHost(host)
	if !ok {
		l.Infof("Router: Couldn't find the sub domain for the request %s", host)
		return nil, false
	}

	l.Infof("Router: for id: %s, router: %s", id, r.String())
	if proxy, ok := r.proxies[id]; ok {
		l.Info("Router: found proxy")
		return proxy, true
	}
	return nil, false
}
