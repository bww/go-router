package router

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/bww/go-router/v2/entity"
)

type Response struct {
	Status int
	Header http.Header
	Entity io.ReadCloser
}

func NewResponse(status int) *Response {
	return &Response{
		Status: status,
		Header: make(http.Header),
	}
}

func (r *Response) SetHeader(k, v string) *Response {
	r.Header.Set(k, v)
	return r
}

func (r *Response) SetEntity(e entity.Entity) (*Response, error) {
	data := e.Data()
	var closer io.ReadCloser
	if c, ok := data.(io.ReadCloser); ok {
		closer = c
	} else {
		closer = ioutil.NopCloser(data)
	}
	r.Entity = closer
	r.Header.Set("Content-Type", e.Type())
	return r, nil
}

func (r *Response) SetBytes(t string, d []byte) (*Response, error) {
	e, err := entity.NewBytes(t, d)
	if err != nil {
		return nil, err
	}
	return r.SetEntity(e)
}

func (r *Response) SetString(t, d string) (*Response, error) {
	e, err := entity.NewString(t, d)
	if err != nil {
		return nil, err
	}
	return r.SetEntity(e)
}

func (r *Response) SetJSON(d interface{}) (*Response, error) {
	e, err := entity.NewJSON(d)
	if err != nil {
		return nil, err
	}
	return r.SetEntity(e)
}

func (r *Response) ReadEntity() ([]byte, error) {
	if r.Entity == nil {
		return []byte{}, nil
	} else {
		return ioutil.ReadAll(r.Entity)
	}
}
