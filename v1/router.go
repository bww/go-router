package router

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"

	pathutil "path"

	"github.com/bww/go-router/v1/path"
)

// Request attributes
type Attributes map[string]interface{}

func (a Attributes) Copy() Attributes {
	c := make(Attributes)
	for k, v := range a {
		c[k] = v
	}
	return c
}

// Request context
type Context struct {
	Vars  path.Vars
	Attrs Attributes
	Path  string
}

// Matching state
type matchState struct {
	Query url.Values
}

// Route handler
type Handler func(*Request, Context) (*Response, error)

// Candidate route matcher
type Matcher func(*Request, Route) bool

// Middleware provides functionality to wrap a handler producing another handler
type Middleware interface {
	Wrap(Handler) Handler
}

// Middleware function wrapper
type MiddlewareFunc func(Handler) Handler

func (m MiddlewareFunc) Wrap(h Handler) Handler {
	return m(h)
}

// A matched route
type Match struct {
	Method string
	Path   string
	Params url.Values
	Vars   path.Vars
}

// An individual route
type Route struct {
	handler Handler
	methods map[string]struct{}
	paths   []path.Path
	params  url.Values
	attrs   Attributes
	matcher Matcher
}

// Set methods
func (r *Route) Methods(m ...string) *Route {
	if r.methods == nil {
		r.methods = make(map[string]struct{})
	}
	for _, e := range m {
		r.methods[strings.ToLower(e)] = struct{}{}
	}
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

// Match via a user-provided function
func (r *Route) Func(m Matcher) *Route {
	r.matcher = m
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

// Matches the provided request or not; returns the details of
// the match if successful, otherwise nil.
func (r Route) Matches(req *Request, state *matchState) *Match {
	if r.methods != nil { // if no methods specified, all methods match
		if _, ok := r.methods[strings.ToLower(req.Method)]; !ok {
			return nil
		}
	}

	var (
		match bool
		tmpl  string
		vars  map[string]string
	)
	for _, e := range r.paths {
		match, vars = e.Matches(req.URL.Path)
		if match {
			tmpl = e.String()
			break
		}
	}
	if !match {
		return nil
	}

	if len(r.params) > 0 {
		if state.Query == nil {
			state.Query = req.URL.Query()
		}
		for k, v := range r.params {
			c, ok := state.Query[k]
			if !ok {
				return nil
			}
			if !reflect.DeepEqual(v, c) {
				return nil
			}
		}
	}

	if r.matcher != nil {
		if !r.matcher(req, r) {
			return nil
		}
	}

	return &Match{
		Method: req.Method,
		Path:   tmpl,
		Params: r.params,
		Vars:   vars,
	}
}

// Handle the request
func (r *Route) Handle(req *Request, cxt Context) (*Response, error) {
	return r.handler(req, cxt)
}

// Obtain a context for this route and the provided match
func (r *Route) Context(match *Match) Context {
	var vars path.Vars
	if match.Vars != nil {
		vars = match.Vars
	} else {
		vars = make(path.Vars)
	}
	return Context{
		Vars:  vars,
		Attrs: r.attrs.Copy(),
		Path:  match.Path,
	}
}

// Describe this route
func (r *Route) String() string {
	b := strings.Builder{}
	b.WriteString(entryList(r.methods))
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
	Find(r *Request) (*Route, *Match, error)
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
	v := &Route{f, nil, []path.Path{path.Parse(p)}, nil, nil, nil}
	r.routes = append(r.routes, v)
	return v
}

// Find a route for the request, if we have one
func (r router) Find(req *Request) (*Route, *Match, error) {
	state := &matchState{}
	for _, e := range r.routes {
		match := e.Matches(req, state)
		if match != nil {
			return e, match, nil
		}
	}
	return nil, nil, nil
}

// Handle the request
func (r router) Handle(req *Request) (*Response, error) {
	route, match, err := r.Find(req)
	if err != nil {
		return NewResponse(http.StatusInternalServerError).SetString("text/plain", fmt.Sprintf("Could not find route: %v", err))
	} else if route == nil {
		return NewResponse(http.StatusNotFound).SetString("text/plain", "Not found")
	}
	var vars path.Vars
	if match.Vars != nil {
		vars = match.Vars
	} else {
		vars = make(path.Vars)
	}
	return route.Handle(
		(*Request)((*http.Request)(req).WithContext(NewMatchContext(req.Context(), match))),
		Context{
			Vars:  vars,
			Attrs: route.attrs.Copy(),
			Path:  match.Path,
		},
	)
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

// Add middleware which wraps every route AFTER the middleware is added
func (r *subrouter) Use(m Middleware) {
	r.parent.Use(m)
}

// Add a route
func (r *subrouter) Add(p string, f Handler) *Route {
	return r.parent.Add(pathutil.Join(r.prefix, p), f)
}

// Find a route for the request, if we have one
func (r subrouter) Find(req *Request) (*Route, *Match, error) {
	return r.parent.Find(req)
}

// Handle the request
func (r subrouter) Handle(req *Request) (*Response, error) {
	return r.parent.Handle(req)
}

// List of set entries
func entryList(m map[string]struct{}) string {
	if len(m) == 0 {
		return "*"
	}
	n := make([]string, 0, len(m))
	for k, _ := range m {
		n = append(n, k)
	}
	switch len(n) {
	case 1:
		return n[0]
	default:
		sort.Strings(n)
		return "{" + strings.Join(n, ", ") + "}"
	}
}
