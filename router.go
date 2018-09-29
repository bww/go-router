package router

import (
  "fmt"
  "strings"
  "net/http"
)

import (
  "github.com/aws/aws-lambda-go/events"
)

// Parameters
type Vars map[string]string

// Request context
type Context struct {
  Vars  Vars
}

// Route handler
type Handler func(events.APIGatewayProxyRequest, Context)(events.APIGatewayProxyResponse, error)

// A path component
type component string

// Does a component match
func (c component) Matches(s string) (bool, string) {
  if string(c) == s {
    return true, ""
  }else if l := len(c); l < 2 {
    return false, ""
  }else if c[0] == '{' && c[l - 1] == '}' {
    return true, strings.TrimSpace(string(c[1:l - 1])) // matches everything
  }else{
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
    }else if e == '{' {
      invar = true
    }else if e == '}' {
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
  for _, e := range p {
    c, s  = splitPath(s)
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
  if s != "" {
    return false, nil
  }
  return true, vars
}

// An individual route
type Route struct {
  handler Handler
  methods []string
  paths   []path
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

// Matches or not
func (r Route) Matches(m, p string) (bool, map[string]string) {
  if len(r.methods) > 0 {
    if !contains(r.methods, m) {
      return false, nil
    }
  }
  for _, e := range r.paths {
    m, vars := e.Matches(p)
    if m {
      return true, vars
    }
  }
  return false, nil
}

// Exec
func (r *Route) Handle(req events.APIGatewayProxyRequest, context Context) (events.APIGatewayProxyResponse, error) {
  return r.handler(req, context)
}

// Dead simple router
type Router struct {
  routes  []*Route
}

// Create a router
func New() *Router {
  return &Router{}
}

// Add a route
func (r *Router) Add(p string, f Handler) *Route {
  v := &Route{f, nil, []path{parsePath(p)}}
  r.routes = append(r.routes, v)
  return v
}

// Find a route for the request, if we have one
func (r Router) Find(req events.APIGatewayProxyRequest) (*Route, Vars, error) {
  for _, e := range r.routes {
    m, vars := e.Matches(req.HTTPMethod, req.Path)
    if m {
      return e, vars, nil
    }
  }
  return nil, nil, nil
}

// Exec
func (r Router) Handle(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  h, vars, err := r.Find(req)
  if err != nil {
    return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: fmt.Sprintf("Could not find route: %v", err) }, nil
  }else if h == nil {
    return events.APIGatewayProxyResponse{StatusCode: http.StatusNotFound, Body: "Not found" }, nil
  }
  if vars == nil {
    vars = make(Vars)
  }
  return h.Handle(req, Context{vars})
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
