package httpwrap

import (
	"encoding/json"
	"net/http"
	"strings"
)

var (
	corsAllow = "*"
)

// SetCors for some shit cors issue
func SetCors(s string) { corsAllow = s }

// Handler wraps http.HandleFunc
// s is the http status code,
// the way v processed will depend on the method called of Handler.
type Handler func(w http.ResponseWriter, req *http.Request) (s int, v interface{})

// Base json mashalled v of handler returned as response
func (h Handler) Base(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Access-Control-Allow-Origin", corsAllow)
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type,x-requested-with")
	w.Header().Set("Content-Type", "application/json; charset=utf-8; text/plain")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// w.Header().Add("Access-Control-Allow-Headers", "x-requested-with,content-type")

	s, v := h(w, req)
	// http status code error
	if s != http.StatusOK {
		ResponseHTTPError(w, s, v)
		return
	}

	// http status 200
	data, err := json.Marshal(v)
	if err != nil {
		// failed to json marshal v, return http 500
		ResponseHTTPError(w, 500)
	}
	w.Write(data)
	// w.Header().Add("Access-Control-Allow-Origin", "*")
	return
}

// JSON require header content-type application/json
func (h Handler) JSON(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		w.WriteHeader(400)
		w.Write([]byte("400: json request required\n"))
		return
	}
	h.Base(w, req)
}
