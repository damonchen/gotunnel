package rwtunnel

import (
	"io"
	"github.com/op/go-logging"
)

var l = logging.MustGetLogger("tunnel")

func copyFromTo(a, b io.ReadWriteCloser) {
	defer func() {
		a.Close()
	}()
	io.Copy(a, b)
}

type RWTunnel struct {
	src, dst io.ReadWriteCloser
}

func (p *RWTunnel) Proxy() {
	// go copypaste(p.src, p.dst, false, "f")
	// go copypaste(p.dst, p.src, true, "b")
	go copyFromTo(p.src, p.dst)
	go copyFromTo(p.dst, p.src)
}

func NewRWTunnel(src, dst io.ReadWriteCloser) (p *RWTunnel) {
	b := &RWTunnel{src: src, dst: dst}
	b.Proxy()
	return b
}

// // only the actual host/port client should be able to close a
// // connection. ex: Keep-Alive and websockets.

func copypaste(in, out io.ReadWriteCloser, close_in bool, msg string) {
	var buf [512]byte

	defer func() {
		if close_in {
			l.Info("eof closing connection")
			in.Close()
			out.Close()
		}
	}()

	for {
		n, err := in.Read(buf[0:])
		// on readerror, only bail if no other choice.
		if err == io.EOF {
			l.Infof("msg: %v", msg)
			// fmt.Print(msg)
			// time.Sleep(1e9)
			l.Infof("eof %v", msg)
			return
		}
		l.Infof("-- read %v", n)
		if err != nil {
			l.Infof("something wrong while copying in ot out : %v", msg)
			l.Infof("error: %v", err)
			return
		}
		// if n < 1 {
		// 	fmt.Println("nothign to read")
		// 	return
		// }

		l.Infof("-- wrote msg bytes %v", n)

		_, err = out.Write(buf[0:n])
		if err != nil {
			l.Info("something wrong while copying out to in ")
			// l.Fatal("something wrong while copying out to in", err)
			return
		}
	}
}
