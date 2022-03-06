package httpwrap

import ()

// RestResponse is the comman Response object for http request
type RestResponse struct {
	Code  int         `json:"code"`
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
}

// NewRestSuccess return a json marshalled bytes of Response object
func NewRestSuccess(data interface{}) *RestResponse { return &RestResponse{0, "", data} }

// NewRestError return a json marshalled bytes of Response object in error
func NewRestError(code int, msg string) *RestResponse { return &RestResponse{code, msg, ""} }
