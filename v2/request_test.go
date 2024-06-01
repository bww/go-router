package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustNewRequest(method, path string, hdrs map[string]string, remote string) *Request {
	req, err := NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	for k, v := range hdrs {
		req.Header.Set(k, v)
	}
	req.RemoteAddr = remote
	return req
}

func TestOriginAddr(t *testing.T) {
	tests := []struct {
		Req    *Request
		Expect string
	}{
		{
			Req:    mustNewRequest("GET", "/", map[string]string{hdrXForwardedFor: "addr1"}, "remote:19876"),
			Expect: "addr1",
		},
		{
			Req:    mustNewRequest("GET", "/", map[string]string{hdrXForwardedFor: "addr1, addr2, addr3"}, "remote:19876"),
			Expect: "addr1",
		},
		{
			Req:    mustNewRequest("GET", "/", map[string]string{}, "remote:19876"),
			Expect: "remote",
		},
		{
			Req:    mustNewRequest("GET", "/", map[string]string{}, "remote"),
			Expect: "remote",
		},
		{
			Req:    mustNewRequest("GET", "/", map[string]string{}, ""),
			Expect: "",
		},
	}
	for _, e := range tests {
		assert.Equal(t, e.Req.OriginAddr(), e.Expect)
	}
}
