package tcprouter

import (
	"io"
	"net"
)

// can be a Stringer interface.
type Proxy struct {
	id    string
	Proxy *ProxyClient
	Admin io.ReadWriteCloser
}

func (p *Proxy) Id() string {
	return p.id
}

func (p *Proxy) BackendHost(addr string) string {
	return net.JoinHostPort(addr, p.Proxy.Port())
}
func (p *Proxy) FrontHost(addr, port string) string {
	return p.id + "." + addr // assumes id exists
}
func (p *Proxy) Port() string {
	return p.Proxy.Port()
}
