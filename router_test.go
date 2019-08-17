package router

import (
	"fmt"
	"net/http"
	"net/url"
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

	m, _ = parsePath("/*").Matches("/a")
	assert.Equal(t, true, m)
	m, _ = parsePath("/*").Matches("/")
	assert.Equal(t, true, m)
	m, _ = parsePath("/a/*/c").Matches("/a/b/c")
	assert.Equal(t, true, m)
	m, _ = parsePath("/a/*").Matches("/a/b")
	assert.Equal(t, true, m)
	m, _ = parsePath("/a/*").Matches("/a/b/c")
	assert.Equal(t, false, m)
	m, _ = parsePath("/a/*").Matches("/a/b/c/d")
	assert.Equal(t, false, m)
	m, _ = parsePath("/a/**").Matches("/a/b/c/d")
	assert.Equal(t, true, m)
	m, _ = parsePath("/a/**").Matches("/a")
	assert.Equal(t, true, m)
	m, _ = parsePath("/a/**").Matches("/")
	assert.Equal(t, false, m)
	m, _ = parsePath("/a/**/c/d").Matches("/a/b/c/d")
	assert.Equal(t, true, m)
	m, _ = parsePath("/**").Matches("/")
	assert.Equal(t, true, m)
	m, _ = parsePath("/**").Matches("/a/b/c/d")
	assert.Equal(t, true, m)

}

func TestRoutes(t *testing.T) {
	var x *Route
	var v map[string]string
	var req *Request
	var err error

	funcA := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetStringEntity("text/plain", "A")
	}
	funcB := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetStringEntity("text/plain", "B")
	}
	funcC := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetStringEntity("text/plain", "C")
	}
	funcD := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetStringEntity("text/plain", "D")
	}
	funcE := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetStringEntity("text/plain", "E")
	}

	r := New()
	r.Add("/a", funcA).Methods("GET")
	r.Add("/a", funcB).Methods("PUT")
	r.Add("/a", funcC)

	r.Add("/b", funcD)
	r.Add("/{var}", funcE)

	s1 := r.Subrouter("/x")
	s1.Add("/a", funcA).Methods("GET")
	s1.Add("/a", funcB).Methods("PUT")

	s2 := s1.Subrouter("/y")
	s2.Add("/a", funcD).Methods("GET")
	s2.Add("/a", funcE).Methods("PUT")

	s3 := r.Subrouter("/z")
	s3.Add("/a", funcA).Methods("GET").Params(url.Values{"foo": {"bar"}})
	s3.Add("/b", funcB).Methods("GET").Params(url.Values{"foo": {"bar"}, "zap": {"pap"}})

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

	// subrouter paths

	req, err = NewRequest("GET", "/x/a", nil)
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

	req, err = NewRequest("PUT", "/x/a", nil)
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

	req, err = NewRequest("GET", "/x/y/a", nil)
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

	req, err = NewRequest("PUT", "/x/y/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = r.Find(req)
		if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
			if assert.NotNil(t, x) {
				r, _ := x.handler(nil, Context{})
				entity, err := r.ReadEntity()
				if assert.Nil(t, err, fmt.Sprint(err)) {
					assert.Equal(t, []byte("E"), entity)
					assert.Equal(t, map[string]string(nil), v)
				}
			}
		}
	}

	// match in subrouter directly

	req, err = NewRequest("GET", "/x/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = s1.Find(req)
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

	req, err = NewRequest("PUT", "/x/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = s1.Find(req)
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

	req, err = NewRequest("GET", "/x/y/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = s2.Find(req)
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

	req, err = NewRequest("PUT", "/x/y/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = s2.Find(req)
		if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
			if assert.NotNil(t, x) {
				r, _ := x.handler(nil, Context{})
				entity, err := r.ReadEntity()
				if assert.Nil(t, err, fmt.Sprint(err)) {
					assert.Equal(t, []byte("E"), entity)
					assert.Equal(t, map[string]string(nil), v)
				}
			}
		}
	}

	// match with parameters

	req, err = NewRequest("GET", "/z/a?foo=nope", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = r.Find(req)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			assert.Nil(t, x)
		}
	}

	req, err = NewRequest("GET", "/z/a?foo=bar", nil)
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

	req, err = NewRequest("GET", "/z/b?foo=bar", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, v, err = r.Find(req)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			assert.Nil(t, x)
		}
	}

	req, err = NewRequest("GET", "/z/b?foo=bar&zap=pap", nil)
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

}
