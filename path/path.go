package path

import (
	"strings"
)

const (
	wildOne   = component("*")
	wildMulti = component("**")
)

type Vars map[string]string

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
type Path []component

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
func Parse(s string) Path {
	var p Path
	var c string
	for s != "" {
		c, s = splitPath(s)
		p = append(p, component(c))
	}
	return p
}

// Does a path match
func (p Path) Matches(s string) (bool, Vars) {
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

// Describe this path
func (p Path) String() string {
	b := strings.Builder{}
	for i, e := range p {
		if i > 0 {
			b.WriteRune('/')
		}
		b.WriteString(string(e))
	}
	return b.String()
}
