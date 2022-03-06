package httpwrap

import (
	"net/http"
	"testing"
)

func TestHttpServer(t *testing.T) {
	http.HandleFunc("/test", Handler(hello).Base)
	http.ListenAndServe("127.0.0.1:8080", nil)
}

func hello(w http.ResponseWriter, req *http.Request) (int, interface{}) {
	return 200, NewRestSuccess("hello")
}
