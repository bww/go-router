package router

import (
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
