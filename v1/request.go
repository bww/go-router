package router

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
)

const hdrXForwardedFor = "X-Forwarded-For"

type Request http.Request

func NewRequest(method, path string, entity io.Reader) (*Request, error) {
	hreq, err := http.NewRequest(method, path, entity)
	if err != nil {
		return nil, err
	}
	return (*Request)(hreq), nil
}

func (r *Request) Context() context.Context {
	return (*http.Request)(r).Context()
}

// OriginAddr tries to identify the best address possible representing the
// origination of the request given the information available.
//
// When available it prefers the `X-Forwarded-For` header, which is widely
// used by middleware infrastructure for this purpose. Failing that, the
// http.Request.RemotAddr is used with the port portion of te value removed.
//
// If none of the above is available and empty string is returned.
func (r *Request) OriginAddr() string {
	var addr string
	if h := r.Header.Get(hdrXForwardedFor); h != "" {
		addr = parseForwardedFor(h)
	} else {
		addr = r.RemoteAddr
	}
	if addr == "" {
		return ""
	}
	h, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	} else {
		return h
	}
}

func parseForwardedFor(h string) string {
	p := strings.Split(h, ",")
	if l := len(p); l > 0 {
		return strings.TrimSpace(p[0]) // take the first component; this is the original client
	} else {
		return strings.TrimSpace(h)
	}
}
