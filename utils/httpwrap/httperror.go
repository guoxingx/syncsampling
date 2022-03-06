package httpwrap

import (
	"net/http"
)

// ResponseHTTPError response http status
// content will be ignored if s is an correct http status code;
// otherwise the first argument of content expected as a string
// will be return with the unkown http status code.
func ResponseHTTPError(w http.ResponseWriter, s int, content ...interface{}) {
	w.WriteHeader(s)
	var respStr string
	respStr = http.StatusText(s)
	if respStr == "" {
		if len(content) > 0 {
			respStr, ok := content[0].(string)
			if ok {
				w.Write([]byte(respStr))
			} else {
				w.Write([]byte("unknown http status code"))
			}
		} else {
			w.Write([]byte("unknown http status code"))
		}
	}
	w.Write([]byte(respStr))
}
