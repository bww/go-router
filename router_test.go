package router

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/bww/go-router/v2/path"
	"github.com/bww/go-util/v1/errors"

	"github.com/stretchr/testify/assert"
)

var randomChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func tracef(f string, a ...interface{}) {
	_trace(fmt.Sprintf(f, a...))
}

func trace(msgs ...interface{}) {
	_trace(fmt.Sprint(msgs...))
}

func _trace(msg string) {
	pc := make([]uintptr, 10) // at least 1 entry needed
	n := runtime.Callers(3, pc)
	fs := runtime.CallersFrames(pc[:n])
	var f runtime.Frame
	more := true
	for i := 0; more; i++ {
		f, more = fs.Next()
		if i == 0 {
			fmt.Printf("%s:%d %s: %s\n", f.File, f.Line, f.Function, msg)
		} else {
			fmt.Printf("    #%d %s:%d %s\n", i, f.File, f.Line, f.Function)
		}
	}
}

func randomString(n int) string {
	v := make([]byte, n)
	for i := 0; i < n; i++ {
		v[i] = randomChars[seededRand.Intn(len(randomChars))]
	}
	return string(v)
}

func checkRoute(t *testing.T, r Router, req *Request, path string, capture path.Vars, expect []byte, xerr error) {
	x, match, err := r.Find(req)
	if xerr != nil {
		assert.Equal(t, xerr, err)
	} else if assert.NoError(t, err) {
		if assert.NotNil(t, x) {
			r, _ := x.handler((*Request)(req), Context{})
			entity, err := r.ReadEntity()
			if assert.NoError(t, err) {
				assert.Equal(t, expect, entity)
				assert.Equal(t, path, match.Path)
				assert.Equal(t, capture, match.Vars)
			}
		}
	}
}

func handleRoute(t *testing.T, r Router, req *Request, status int, expect []byte, xerr error) {
	rsp, err := r.Handle(req)
	if xerr != nil {
		assert.Equal(t, xerr, err)
	} else if assert.NoError(t, err) && assert.NotNil(t, rsp) {
		assert.Equal(t, status, rsp.Status)
		assert.Equal(t, expect, errors.Must(io.ReadAll(rsp.Entity)))
	}
}

func TestRoutes(t *testing.T) {
	var req *Request
	var err error

	funcA := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetString("text/plain", "A")
	}
	funcB := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetString("text/plain", "B")
	}
	funcC := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetString("text/plain", "C")
	}
	funcD := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetString("text/plain", "D")
	}
	funcE := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetString("text/plain", "E")
	}
	funcF := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetString("text/plain", "F")
	}
	funcG := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetString("text/plain", "G")
	}

	r := New()
	r.Add("/a", funcA).Methods("GET")
	r.Add("/a", funcB).Methods("PUT")
	r.Add("/a", funcC)

	r.Add("/b", funcF).Match(func(req *Request, route *Route) bool {
		return req.Header.Get("Check-Header") == "F"
	})
	r.Add("/b", funcG).Match(func(req *Request, route *Route) bool {
		return req.Header.Get("Check-Header") == "G"
	})
	r.Add("/b", funcD)
	r.Add("/{var}", funcE)

	s1 := r.Subrouter("/x")
	s1.Add("/a", funcA).Methods("GET", "POST")
	s1.Add("/a", funcB).Methods("PUT")

	s2 := s1.Subrouter("/y")
	s2.Add("/a", funcD).Methods("GET")
	s2.Add("/a", funcE).Methods("PUT")
	s2.Add("/a/*/c", funcA).Methods("GET")
	s2.Add("/a/b/**", funcB).Methods("GET")

	s3 := r.Subrouter("/z")
	s3.Add("/a", funcA).Methods("GET").Param("foo", "bar")
	s3.Add("/b", funcB).Methods("GET").Param("foo", "bar").Param("zap", "pap")
	s3.Add("/b", funcC).Methods("GET").Params(url.Values{"foo": {"bar", "car"}, "zap": {"pap"}})

	for _, e := range r.Routes() {
		fmt.Println("> ", e.Describe(true))
	}

	// simple matching

	req, err = NewRequest("GET", "/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/a", nil, []byte("A"), nil)
	}
	req, err = NewRequest("PUT", "/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/a", nil, []byte("B"), nil)
	}
	req, err = NewRequest("ANYTHING", "/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/a", nil, []byte("C"), nil)
	}

	req, err = NewRequest("GET", "/b", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/b", nil, []byte("D"), nil)
	}
	req, err = NewRequest("GET", "/c", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/{var}", map[string]string{"var": "c"}, []byte("E"), nil)
	}

	// subrouter paths

	req, err = NewRequest("GET", "/x/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/a", nil, []byte("A"), nil)
	}
	req, err = NewRequest("POST", "/x/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/a", nil, []byte("A"), nil)
	}
	req, err = NewRequest("PUT", "/x/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/a", nil, []byte("B"), nil)
	}

	req, err = NewRequest("GET", "/x/y/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/y/a", nil, []byte("D"), nil)
	}
	req, err = NewRequest("PUT", "/x/y/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/y/a", nil, []byte("E"), nil)
	}
	req, err = NewRequest("GET", "/x/y/a/foo/c", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/y/a/*/c", nil, []byte("A"), nil)
	}
	req, err = NewRequest("GET", "/x/y/a/b/c/d", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/y/a/b/**", nil, []byte("B"), nil)
	}

	// match in subrouter directly

	req, err = NewRequest("GET", "/x/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/a", nil, []byte("A"), nil)
	}
	req, err = NewRequest("PUT", "/x/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/a", nil, []byte("B"), nil)
	}

	req, err = NewRequest("GET", "/x/y/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/y/a", nil, []byte("D"), nil)
	}
	req, err = NewRequest("PUT", "/x/y/a", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/x/y/a", nil, []byte("E"), nil)
	}

	// match with parameters

	req, err = NewRequest("GET", "/z/a?foo=nope", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, _, err := r.Find(req)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			assert.Nil(t, x)
		}
	}
	req, err = NewRequest("GET", "/z/a?foo=bar", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/z/a", nil, []byte("A"), nil)
	}
	req, err = NewRequest("GET", "/z/b?foo=bar", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		x, _, err := r.Find(req)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			assert.Nil(t, x)
		}
	}
	req, err = NewRequest("GET", "/z/b?foo=bar&zap=pap", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/z/b", nil, []byte("B"), nil)
	}
	req, err = NewRequest("GET", "/z/b?foo=bar&foo=car&zap=pap", nil)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/z/b", nil, []byte("C"), nil)
	}

	// match with function

	req, err = NewRequest("GET", "/b", nil)
	req.Header.Set("Check-Header", "F") // matches F instead of D because of this header
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/b", nil, []byte("F"), nil)
	}
	req, err = NewRequest("GET", "/b", nil)
	req.Header.Set("Check-Header", "G") // matches G instead of D or F because of this header
	if assert.Nil(t, err, fmt.Sprint(err)) {
		checkRoute(t, r, req, "/b", nil, []byte("G"), nil)
	}

}

func BenchmarkRoutes(b *testing.B) {

	funcA := func(*Request, Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetString("text/plain", "A")
	}

	r := New()
	r.Add("/a", funcA).Methods("GET")

	s1 := r.Subrouter("/x")
	s1.Add("/a", funcA).Methods("GET")

	s2 := s1.Subrouter("/y")
	s2.Add("/a", funcA).Methods("GET")

	s3 := r.Subrouter("/z")
	s3.Add("/a", funcA).Methods("GET").Param("foo", "bar")
	s3.Add("/b", funcA).Methods("GET").Param("foo", "bar").Param("zap", "pap")
	s3.Add("/b", funcA).Methods("GET").Params(url.Values{"foo": {"bar", "car"}, "zap": {"pap"}})

	for n := 0; n < b.N; n++ {
		req, err := NewRequest("GET", "/z/b?foo=bar&foo=car&zap=pap", nil)
		if err != nil {
			panic(err)
		}
		x, _, err := r.Find(req)
		if err != nil {
			panic(err)
		}
		if x == nil {
			panic(fmt.Errorf("Could not route: %v", req))
		}
	}

}

func TestRouteAttrs(t *testing.T) {
	var req *Request
	var err error

	funcA := func(_ *Request, cxt Context) (*Response, error) {
		key, val := randomString(16), randomString(16)
		cxt.Attrs[key] = val
		fmt.Println(">>>", cxt.Attrs)
		assert.Equal(t, Attributes{"key": "val", key: val}, cxt.Attrs)
		return NewResponse(http.StatusOK).SetString("text/plain", "A")
	}

	r := New()
	r.Add("/a", funcA).Methods("GET").Attr("key", "val")

	for _, e := range r.Routes() {
		fmt.Println("> ", e)
	}

	for i := 0; i < 3; i++ {
		req, err = NewRequest("GET", "/a", nil)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			_, err := r.Handle(req)
			assert.Nil(t, err, fmt.Sprint(err))
		}
	}

}

func TestMiddleware(t *testing.T) {
	var req *Request
	var err error

	handlerA := func(req *Request, cxt Context) (*Response, error) {
		return NewResponse(http.StatusOK).SetString("text/plain", cxt.Path)
	}

	middleB := func(h Handler) Handler {
		return func(req *Request, cxt Context) (*Response, error) {
			match := MatchFromContext(req.Context())
			if assert.NotNil(t, match) {
				assert.Equal(t, match.Path, cxt.Path)
			}
			rsp, err := h(req, cxt)
			return rsp, err
		}
	}

	middleC := func(h Handler) Handler {
		return func(req *Request, cxt Context) (*Response, error) {
			rsp, err := h(req, cxt)
			if err != nil {
				return nil, err
			}
			return NewResponse(http.StatusOK).SetString("text/plain", fmt.Sprintf("%s: %v",
				string(errors.Must(io.ReadAll(rsp.Entity))),
				cxt.Attrs,
			))
		}
	}

	r := New()
	r.Use(MiddleFunc(middleB))
	r.Add("/a", handlerA).Methods("GET").Attr("key", "val")
	r.Add("/b", handlerA).Methods("GET").Attr("key", "val").Use(MiddleFunc(middleC))

	req, err = NewRequest("GET", "/a", nil)
	if assert.NoError(t, err) {
		handleRoute(t, r, req, http.StatusOK, []byte(req.URL.Path), nil)
	}
	req, err = NewRequest("GET", "/b", nil)
	if assert.NoError(t, err) {
		handleRoute(t, r, req, http.StatusOK, []byte(fmt.Sprintf("%s: key=val", req.URL.Path)), nil)
	}
}
