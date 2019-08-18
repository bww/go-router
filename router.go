package router

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	paths "path"
)

const (
	wildOne   = component("*")
	wildMulti = component("**")
)

type Vars map[string]string
type Attributes map[string]interface{}

// Request context
type Context struct {
	Vars  Vars
	Attrs Attributes
}

// Matching state
type matchState struct {
	Query url.Values
}

// Route handler
type Handler func(*Request, Context) (*Response, error)

// A path component
type component string

// Does a component match
func (c component) Matches(s string) (bool, string) {
	if c == wildOne || c == wildMulti {
		return true, "" // matches everything, captures nothing
	} else if string(c) == s {
		return true, ""
	} else if l := len(c); l < 2 {
		return false, ""
	} else if c[0] == '{' && c[l-1] == '}' {
		return true, strings.TrimSpace(string(c[1 : l-1])) // matches everything
	} else {
		return false, ""

	}
}

// A matching path
type path []component

// Split a path into (first component, remainder)
func splitPath(s string) (string, string) {
	var invar bool
	for i, e := range s {
		if e == '/' && !invar {
			return s[:i], s[i+1:]
		} else if e == '{' {
			invar = true
		} else if e == '}' {
			invar = false
		}
	}
	return s, ""
}

// Parse a path
func parsePath(s string) path {
	p := make(path, 0)
	var c string
	for s != "" {
		c, s = splitPath(s)
		p = append(p, component(c))
	}
	return p
}

// Does a path match
func (p path) Matches(s string) (bool, Vars) {
	var vars map[string]string
	var c string
	var e component
	for _, e = range p {
		c, s = splitPath(s)
		m, n := e.Matches(c)
		if !m {
			return false, nil
		}
		if n != "" {
			if vars == nil {
				vars = make(map[string]string)
			}
			vars[n] = c
		}
	}
	if s != "" && e != wildMulti {
		return false, nil
	}
	return true, vars
}

// An individual route
type Route struct {
	handler Handler
	methods []string
	paths   []path
	params  url.Values
	attrs   Attributes
}

// Set methods
func (r *Route) Methods(m ...string) *Route {
	r.methods = m
	return r
}

// Add additional paths
func (r *Route) Paths(s ...string) *Route {
	p := make([]path, len(s))
	for i, e := range s {
		p[i] = parsePath(e)
	}
	r.paths = append(r.paths, p...)
	return r
}

// Match a single parameter
func (r *Route) Param(k, v string) *Route {
	if r.params == nil {
		r.params = make(url.Values)
	}
	r.params.Set(k, v)
	return r
}

// Match a set of parameters
func (r *Route) Params(p url.Values) *Route {
	if r.params == nil {
		r.params = make(url.Values)
	}
	for k, v := range p {
		r.params[k] = v
	}
	return r
}

// Set an attribute
func (r *Route) Attr(k string, v interface{}) *Route {
	if r.attrs == nil {
		r.attrs = make(Attributes)
	}
	r.attrs[k] = v
	return r
}

// Matches or not
func (r Route) Matches(req *Request, state *matchState) (bool, map[string]string) {
	if len(r.methods) > 0 {
		if !contains(r.methods, req.Method) {
			return false, nil
		}
	}

	var match bool
	var vars map[string]string
	for _, e := range r.paths {
		match, vars = e.Matches(req.URL.Path)
		if match {
			break
		}
	}
	if !match {
		return false, nil
	}

	if len(r.params) > 0 {
		if state.Query == nil {
			state.Query = req.URL.Query()
		}
		for k, v := range r.params {
			c, ok := state.Query[k]
			if !ok {
				return false, nil
			}
			if !reflect.DeepEqual(v, c) {
				return false, nil
			}
		}
	}

	return match, vars
}

// Handle the request
func (r *Route) Handle(req *Request, context Context) (*Response, error) {
	return r.handler(req, context)
}

// Dead simple router
type Router interface {
	Use(f Handler)
	Add(p string, f Handler) *Route
	Find(r *Request) (*Route, Vars, error)
	Handle(r *Request) (*Response, error)
	Subrouter(p string) Router
}

type router struct {
	routes     []*Route
	middleware []Handler
}

func New() Router {
	return &router{}
}

// Derive a subrouter from this router with the specified path prefix
func (r *router) Subrouter(p string) Router {
	return &subrouter{r, p}
}

// Add middleware which is executed for every route
func (r *router) Use(f Handler) {
	r.middleware = append(r.middleware, f)
}

// Add a route
func (r *router) Add(p string, f Handler) *Route {
	v := &Route{f, nil, []path{parsePath(p)}, nil, nil}
	r.routes = append(r.routes, v)
	return v
}

// Find a route for the request, if we have one
func (r router) Find(req *Request) (*Route, Vars, error) {
	state := &matchState{}
	for _, e := range r.routes {
		m, vars := e.Matches(req, state)
		if m {
			return e, vars, nil
		}
	}
	return nil, nil, nil
}

// Handle the request
func (r router) Handle(req *Request) (*Response, error) {
	h, vars, err := r.Find(req)
	if err != nil {
		return NewResponse(http.StatusInternalServerError).SetStringEntity("text/plain", fmt.Sprintf("Could not find route: %v", err))
	} else if h == nil {
		return NewResponse(http.StatusNotFound).SetStringEntity("text/plain", "Not found")
	}
	if vars == nil {
		vars = make(Vars)
	}

	cxt := Context{vars, h.attrs}
	for _, e := range r.middleware {
		rsp, err := e(req, cxt)
		if err != nil {
			return nil, err
		}
		if rsp != nil {
			return rsp, nil
		}
	}

	return h.Handle(req, cxt)
}

type subrouter struct {
	parent Router
	prefix string
}

// Derive a subrouter from this router with the specified path prefix
func (r *subrouter) Subrouter(p string) Router {
	return &subrouter{r, p}
}

// Add middleware which is executed for every route
func (r *subrouter) Use(f Handler) {
	r.parent.Use(f)
}

// Add a route
func (r *subrouter) Add(p string, f Handler) *Route {
	return r.parent.Add(paths.Join(r.prefix, p), f)
}

// Find a route for the request, if we have one
func (r subrouter) Find(req *Request) (*Route, Vars, error) {
	return r.parent.Find(req)
}

// Handle the request
func (r subrouter) Handle(req *Request) (*Response, error) {
	return r.parent.Handle(req)
}

// Is a string in the set
func contains(a []string, s string) bool {
	for _, e := range a {
		if strings.EqualFold(e, s) {
			return true
		}
	}
	return false
}
