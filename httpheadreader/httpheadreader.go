package httpheadreader

import (
	"bufio"
	"bytes"
	//l "github.com/ciju/gotunnel/log"
	"net"
	"net/http"
	"regexp"
	"github.com/op/go-logging"
)

var l = logging.MustGetLogger("http header reader")

type HTTPHeadReader struct {
	conn net.Conn

	host string
	buf  []byte
	err  error
	req  *http.Request
}

func NewHTTPHeadReader(c net.Conn) (h *HTTPHeadReader) {
	return &HTTPHeadReader{conn: c}
}

func (c *HTTPHeadReader) parseHeaders() (err error) {
	var buf [http.DefaultMaxHeaderBytes]byte

	n, err := c.conn.Read(buf[0:])
	if err != nil {
		l.Infof("H: error while reading: %v", err)
		return err
	}
	l.Infof("H: bytes: %d", n)
	c.buf = make([]byte, n)
	copy(c.buf, buf[0:n])

	c.req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(c.buf[0:n])))
	if err != nil {
		l.Infof("H: error while parsing header: %v", err)
		return err
	}
	return nil
}

func (c *HTTPHeadReader) regexpHost() string {
	// TODO: make this generic
	reg, err := regexp.Compile(`(\w*\.localtunnel\.net)`)
	if err != nil {
		l.Infof("H: couldn't find host")
		return ""
	}

	if reg.Match(c.buf[0:]) {
		l.Infof("H: found host: %s", reg.FindString(string(c.buf[0:])))
		return reg.FindString(string(c.buf[0:]))
	}
	return ""
}

func (c *HTTPHeadReader) Host() string {
	if c.req != nil {
		return c.req.Host
	}

	err := c.parseHeaders()
	if err != nil {
		l.Infof("H: error:  %s", err)
		return c.regexpHost()
	}

	return c.req.Host
}

func (c *HTTPHeadReader) Read(b []byte) (int, error) {
	// read from internal buffer
	if c.err != nil {
		return 0, c.err
	}
	if len(c.buf) != 0 {
		n := copy(b, c.buf)
		c.buf = c.buf[n:]
		l.Infof("copied: %d - remaining %d ", n, len(c.buf))
		return n, nil
	}
	return c.conn.Read(b)
}
func (c *HTTPHeadReader) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}
func (c *HTTPHeadReader) Close() error {
	return c.conn.Close()
}
