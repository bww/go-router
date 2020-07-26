package router

import (
	"context"
	"io"
	"net/http"
)

type Request http.Request

func NewRequest(method, path string, entity io.Reader) (*Request, error) {
	hreq, err := http.NewRequest(method, path, entity)
	if err != nil {
		return nil, err
	}
	return (*Request)(hreq), nil
}

func (r *Request) Clone(cxt context.Context) *Request {
	return (*Request)((*http.Request)(r).Clone(cxt))
}

func (r *Request) Context() context.Context {
	return (*http.Request)(r).Context()
}

func (r *Request) WithContext(cxt context.Context) *Request {
	return (*Request)((*http.Request)(r).WithContext(cxt))
}
