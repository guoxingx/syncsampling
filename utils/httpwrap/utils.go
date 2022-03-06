package httpwrap

import (
	"fmt"
	"net/http"
	"strconv"
)

// ParseRequestURLArgsInt64 parse args from url params of request and returns in int64
// the 3rd args is optional as defaultt value
func ParseRequestURLArgsInt64(req *http.Request, key string, defaultt ...int64) int64 {
	res, err := MustParseRequestURLArgsInt64(req, key)
	if err != nil {
		if len(defaultt) == 0 {
			return 0
		}
		return defaultt[0]
	}
	return res
}

// ParseRequestURLArgsBool parse args from url params of request and returns in bool
// the 3rd args is optional as defaultt value
func ParseRequestURLArgsBool(req *http.Request, key string, defaultt ...bool) bool {
	res, err := MustParseRequestURLArgsBool(req, key)
	if err != nil {
		if len(defaultt) == 0 {
			return false
		}
		return defaultt[0]
	}
	return res
}

// ParseRequestURLArgsUint64 parse args from url params of request and returns in uint64
// the 3rd args is optional as defaultt value
func ParseRequestURLArgsUint64(req *http.Request, key string, defaultt ...uint64) uint64 {
	res, err := MustParseRequestURLArgsUint64(req, key)
	if err != nil {
		if len(defaultt) == 0 {
			return 0
		}
		return defaultt[0]
	}
	return res
}

// ParseRequestURLArgsFloat64 parse args from url params of request and returns in float64
// the 3rd args is optional as defaultt value
func ParseRequestURLArgsFloat64(req *http.Request, key string, defaultt ...float64) float64 {
	res, err := MustParseRequestURLArgsFloat64(req, key)
	if err != nil {
		if len(defaultt) == 0 {
			return 0
		}
		return defaultt[0]
	}
	return res
}

// MustParseRequestURLArgsInt64 parse args from url params of request
// and returns in int64 with an error
func MustParseRequestURLArgsInt64(req *http.Request, key string) (int64, error) {
	resStr := req.FormValue(key)
	if resStr != "" {
		return strconv.ParseInt(resStr, 10, 64)
	}
	return 0, fmt.Errorf("\"%s\" not found in url arguments", key)
}

// MustParseRequestURLArgsBool parse args from url params of request
// and returns in bool with an error
func MustParseRequestURLArgsBool(req *http.Request, key string) (bool, error) {
	resStr := req.FormValue(key)
	if resStr != "" {
		return strconv.ParseBool(resStr)
	}
	return false, fmt.Errorf("\"%s\" not found in url arguments", key)
}

// MustParseRequestURLArgsUint64 parse args from url params of request
// and returns in uint64 with an error
func MustParseRequestURLArgsUint64(req *http.Request, key string) (uint64, error) {
	resStr := req.FormValue(key)
	if resStr != "" {
		return strconv.ParseUint(resStr, 10, 64)
	}
	return 0, fmt.Errorf("\"%s\" not found in url arguments", key)
}

// MustParseRequestURLArgsFloat64 parse args from url params of request
// and returns in float64 with an error
func MustParseRequestURLArgsFloat64(req *http.Request, key string) (float64, error) {
	resStr := req.FormValue(key)
	if resStr != "" {
		return strconv.ParseFloat(resStr, 64)
	}
	return 0, fmt.Errorf("\"%s\" not found in url arguments", key)
}
