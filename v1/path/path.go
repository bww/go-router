package path

import (
	"strings"
)

const (
	wildOne    = component("*")
	wildMulti  = component("**")
	defaultSep = '/'
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
type Path struct {
	cmp []component
	sep rune
}

// Split a path into (first component, remainder)
func splitPath(s string, sep rune, vars bool) (string, string) {
	var invar bool
	for i, e := range s {
		if e == sep && !invar {
			return s[:i], s[i+1:]
		} else if vars && e == '{' {
			invar = true
		} else if vars && e == '}' {
			invar = false
		}
	}
	return s, ""
}

// Parse a path using the default separator '/'
func Parse(s string) Path {
	return ParseSeparator(s, defaultSep)
}

// Parse a path using the specified separator
func ParseSeparator(s string, sep rune) Path {
	var p []component
	var c string
	for s != "" {
		c, s = splitPath(s, sep, true)
		p = append(p, component(c))
	}
	return Path{
		cmp: p,
		sep: sep,
	}
}

// Does a path match
func (p Path) Matches(s string) (bool, Vars) {
	var vars map[string]string
	var c string
	var e component
	for _, e = range p.cmp {
		c, s = splitPath(s, p.sep, false)
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
	for i, e := range p.cmp {
		if i > 0 {
			b.WriteRune(p.sep)
		}
		b.WriteString(string(e))
	}
	return b.String()
}
