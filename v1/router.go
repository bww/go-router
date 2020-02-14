package router

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	pathutil "path"

	"github.com/bww/go-router/v1/path"
)

type Attributes map[string]interface{}

// Request context
type Context struct {
	Vars  path.Vars
	Attrs Attributes
}

// Matching state
type matchState struct {
	Query url.Values
}

// Route handler
type Handler func(*Request, Context) (*Response, error)

// Middleware provides functionality to wrap a handler producing another handler
type Middleware interface {
	Wrap(Handler) Handler
}

// An individual route
type Route struct {
	handler Handler
	methods []string
	paths   []path.Path
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
	p := make([]path.Path, len(s))
	for i, e := range s {
		p[i] = path.Parse(e)
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

// Describe this route
func (r *Route) String() string {
	b := strings.Builder{}
	switch len(r.methods) {
	case 0:
		b.WriteString("*")
	case 1:
		b.WriteString(r.methods[0])
	default:
		b.WriteString("{" + strings.Join(r.methods, ", ") + "}")
	}
	b.WriteString(" ")
	switch len(r.paths) {
	case 0:
		b.WriteString("{}")
	case 1:
		b.WriteString(r.paths[0].String())
	default:
		b.WriteString("{")
		for i, e := range r.paths {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(e.String())
		}
		b.WriteString("}")
	}
	if len(r.params) > 0 {
		b.WriteString(" ?")
		b.WriteString(r.params.Encode())
	}
	return b.String()
}

// Dead simple router
type Router interface {
	Use(m Middleware)
	Add(p string, f Handler) *Route
	Find(r *Request) (*Route, path.Vars, error)
	Handle(r *Request) (*Response, error)
	Subrouter(p string) Router
	Routes() []*Route
}

type router struct {
	routes     []*Route
	middleware []Middleware
}

func New() Router {
	return &router{}
}

// Obtain a copy of all the routes managed by this router
func (r *router) Routes() []*Route {
	routes := make([]*Route, len(r.routes))
	copy(routes, r.routes)
	return routes
}

// Derive a subrouter from this router with the specified path prefix
func (r *router) Subrouter(p string) Router {
	return &subrouter{r, p}
}

// Add middleware which is wraps every route that is added AFTER the
// middeware is defined. Routes added before a middleware will not
// be affected.
//
// Routes are wrapped by middleware in the order the middleware is
// added to the router via this method.
func (r *router) Use(m Middleware) {
	r.middleware = append(r.middleware, m)
}

// Add a route
func (r *router) Add(p string, f Handler) *Route {
	for _, e := range r.middleware {
		f = e.Wrap(f)
	}
	v := &Route{f, nil, []path.Path{path.Parse(p)}, nil, nil}
	r.routes = append(r.routes, v)
	return v
}

// Find a route for the request, if we have one
func (r router) Find(req *Request) (*Route, path.Vars, error) {
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
		return NewResponse(http.StatusInternalServerError).SetString("text/plain", fmt.Sprintf("Could not find route: %v", err))
	} else if h == nil {
		return NewResponse(http.StatusNotFound).SetString("text/plain", "Not found")
	}
	if vars == nil {
		vars = make(path.Vars)
	}
	return h.Handle(req, Context{vars, h.attrs})
}

type subrouter struct {
	parent Router
	prefix string
}

// Subrouter routes are not supported
func (r *subrouter) Routes() []*Route {
	return nil
}

// Derive a subrouter from this router with the specified path prefix
func (r *subrouter) Subrouter(p string) Router {
	return &subrouter{r, p}
}

// Add middleware which is wraps every route that is added AFTER the
// middeware is defined. Routes added before a middleware will not
// be affected.
//
// Routes are wrapped by middleware in the order the middleware is
// added to the router via this method.
func (r *subrouter) Use(m Middleware) {
	r.parent.Use(m)
}

// Add a route
func (r *subrouter) Add(p string, f Handler) *Route {
	return r.parent.Add(pathutil.Join(r.prefix, p), f)
}

// Find a route for the request, if we have one
func (r subrouter) Find(req *Request) (*Route, path.Vars, error) {
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
