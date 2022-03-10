package webserver

import (
	"net/http"

	"syncsampling/logs"
	"syncsampling/utils/httpwrap"

	"go.uber.org/zap"
)

var (
	logger   *zap.SugaredLogger
	host     string
	actionCh chan int
)

func StartServer() error {
	logger = logs.GetLogger()
	host = "localhost:3000"

	httpwrap.SetCors("http://localhost:8080")
	// httpwrap.SetCors("http://127.0.0.1:8080")

	http.HandleFunc("/api/action", httpwrap.Handler(handleAction).Base)
	http.HandleFunc("/api/image", httpwrap.Handler(handleImage).Base)
	http.HandleFunc("/api/images", httpwrap.Handler(handleImages).Base)
	http.HandleFunc("/api/images/contrast", httpwrap.Handler(handleContrast).Base)

	http.Handle("/", http.FileServer(http.Dir("../statics")))

	return http.ListenAndServe(host, nil)
}

func InitActionCh(ch chan int) {
	actionCh = ch
}
