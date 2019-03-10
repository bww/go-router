package router

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
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

type Response struct {
	Status int
	Header http.Header
	Entity io.ReadCloser
}

func NewResponse(status int, entity io.Reader) (*Response, error) {
	var closer io.ReadCloser
	if c, ok := entity.(io.ReadCloser); ok {
		closer = c
	} else {
		closer = ioutil.NopCloser(entity)
	}
	return &Response{
		Status: status,
		Header: make(http.Header),
		Entity: closer,
	}, nil
}

func NewBytesResponse(status int, entity []byte) (*Response, error) {
	return NewResponse(status, bytes.NewReader(entity))
}

func NewStringResponse(status int, entity string) (*Response, error) {
	return NewResponse(status, bytes.NewReader([]byte(entity)))
}

func NewJSONResponse(status int, entity interface{}) (*Response, error) {
	data, err := json.Marshal(entity)
	if err != nil {
		return nil, err
	}

	rsp, err := NewBytesResponse(status, data)
	if err != nil {
		return nil, err
	}

	rsp.Header.Set("Content-Type", "application/json")
	return rsp, nil
}

func (r *Response) ReadEntity() ([]byte, error) {
	if r.Entity == nil {
		return []byte{}, nil
	} else {
		return ioutil.ReadAll(r.Entity)
	}
}
