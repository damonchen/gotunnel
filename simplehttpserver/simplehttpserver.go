package simplehttpserver

import (
	"fmt"
	"net/http"
	"github.com/op/go-logging"
)


var l = logging.MustGetLogger("http server")

func NewSimpleHTTPServer(port string, dir string) {

	if dir == "" {
		l.Fatal("No directory given, to serve")
	}

	http.Handle("/", http.FileServer(http.Dir(dir)))

	fmt.Println("Serving", dir, "at port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		l.Fatalf("error %v", err)
	}
}
