package router

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaths(t *testing.T) {
	var m bool
	var v map[string]string

	m, _ = parsePath("/").Matches("/")
	assert.Equal(t, true, m)
	m, _ = parsePath("/a").Matches("/a")
	assert.Equal(t, true, m)
	m, _ = parsePath("/a/b").Matches("/a/b")
	assert.Equal(t, true, m)

	m, v = parsePath("/a/{var}").Matches("/a/b")
	assert.Equal(t, true, m)
	assert.Equal(t, map[string]string{"var": "b"}, v)
	m, v = parsePath("/a/{var}/c").Matches("/a/b/c")
	assert.Equal(t, true, m)
	assert.Equal(t, map[string]string{"var": "b"}, v)
	m, v = parsePath("/a/{v/r}/c").Matches("/a/b/c")
	assert.Equal(t, true, m)
	assert.Equal(t, map[string]string{"v/r": "b"}, v)
	m, v = parsePath("/a/{var1}/c/{var2}").Matches("/a/b/c/d")
	assert.Equal(t, true, m)
	assert.Equal(t, map[string]string{"var1": "b", "var2": "d"}, v)

	m, _ = parsePath("/").Matches("/a")
	assert.Equal(t, false, m)
	m, _ = parsePath("/a").Matches("/")
	assert.Equal(t, false, m)
	m, _ = parsePath("/a/b").Matches("/a/c")
	assert.Equal(t, false, m)
	m, _ = parsePath("/a/{var}").Matches("/a/b/c")
	assert.Equal(t, false, m)
	m, _ = parsePath("/a/c/{var}").Matches("/a/b/c")
	assert.Equal(t, false, m)
	m, _ = parsePath("/a/{var1}/{var2}").Matches("/x/b/c")
	assert.Equal(t, false, m)
	m, _ = parsePath("/a/{var1}/{var2}").Matches("/a/b/c/d")
	assert.Equal(t, false, m)

}

func TestRoutes(t *testing.T) {
	var x *Route
	var v map[string]string
	var req *Request
	var err error

	funcA := func(*Request, Context) (*Response, error) {
		return NewStringResponse(http.StatusOK, "A")
	}
	funcB := func(*Request, Context) (*Response, error) {
		return NewStringResponse(http.StatusOK, "B")
	}
	funcC := func(*Request, Context) (*Response, error) {
		return NewStringResponse(http.StatusOK, "C")
	}
	funcD := func(*Request, Context) (*Response, error) {
		return NewStringResponse(http.StatusOK, "D")
	}
	funcE := func(*Request, Context) (*Response, error) {
		return NewStringResponse(http.StatusOK, "E")
	}

	r := &Router{}
	r.Add("/a", funcA).Methods("GET")
	r.Add("/a", funcB).Methods("PUT")
	r.Add("/a", funcC)

	r.Add("/b", funcD)
	r.Add("/{var}", funcE)

	req, err = NewRequest("GET", "/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = r.Find(req)
		if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
			if assert.NotNil(t, x) {
				r, _ := x.handler(nil, Context{})
				entity, err := r.ReadEntity()
				if assert.Nil(t, err, fmt.Sprint(err)) {
					assert.Equal(t, []byte("A"), entity)
					assert.Equal(t, map[string]string(nil), v)
				}
			}
		}
	}

	req, err = NewRequest("PUT", "/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = r.Find(req)
		if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
			if assert.NotNil(t, x) {
				r, _ := x.handler(nil, Context{})
				entity, err := r.ReadEntity()
				if assert.Nil(t, err, fmt.Sprint(err)) {
					assert.Equal(t, []byte("B"), entity)
					assert.Equal(t, map[string]string(nil), v)
				}
			}
		}
	}

	req, err = NewRequest("ANYTHING", "/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = r.Find(req)
		if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
			if assert.NotNil(t, x) {
				r, _ := x.handler(nil, Context{})
				entity, err := r.ReadEntity()
				if assert.Nil(t, err, fmt.Sprint(err)) {
					assert.Equal(t, []byte("C"), entity)
					assert.Equal(t, map[string]string(nil), v)
				}
			}
		}
	}

	req, err = NewRequest("GET", "/b", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = r.Find(req)
		if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
			if assert.NotNil(t, x) {
				r, _ := x.handler(nil, Context{})
				entity, err := r.ReadEntity()
				if assert.Nil(t, err, fmt.Sprint(err)) {
					assert.Equal(t, []byte("D"), entity)
					assert.Equal(t, map[string]string(nil), v)
				}
			}
		}
	}

	req, err = NewRequest("GET", "/c", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = r.Find(req)
		if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
			if assert.NotNil(t, x) {
				r, _ := x.handler(nil, Context{})
				entity, err := r.ReadEntity()
				if assert.Nil(t, err, fmt.Sprint(err)) {
					assert.Equal(t, []byte("E"), entity)
					assert.Equal(t, map[string]string{"var": "c"}, v)
				}
			}
		}
	}

}
