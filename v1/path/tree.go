package path

import (
	"errors"
	"strings"
)

var ErrCollision = errors.New("Path collision")

type node struct {
	cmp   component
	value interface{}
	sub   *Tree
}

// A path prefix tree
type Tree struct {
	n   []*node
	sep rune
}

// Create a tree with the specified separator. The zero
// value of a tree uses the default separator: '/'.
func NewTree(sep rune) *Tree {
	return &Tree{sep: sep}
}

func (t *Tree) separator() rune {
	if t.sep == 0 {
		return '/'
	} else {
		return t.sep
	}
}

func (t *Tree) Add(p string, v interface{}) error {
	return t.add(ParseSeparator(p, t.separator()).cmp, v)
}

func (t *Tree) add(p []component, v interface{}) error {
	if l := len(p); l > 0 {
		var f *node

		// attempt to find the first component
		for _, e := range t.n {
			if e.cmp.Equals(p[0]) {
				f = e
				break
			}
		}

		// create the node if we didn't find it
		if f == nil {
			f = &node{cmp: p[0]}
			t.n = append(t.n, f)
		}

		// recurse to it with the remaining components or finalize
		if l > 1 {
			if f.sub == nil {
				f.sub = &Tree{}
			}
			return f.sub.add(p[1:], v)
		} else {
			if f.value != nil {
				return ErrCollision
			}
			f.value = v
		}
	}
	return nil
}

func (t *Tree) Find(s string) (interface{}, Vars, bool) {
	return t.find(s, Vars{})
}

func (t *Tree) find(s string, vars Vars) (interface{}, Vars, bool) {
	c, r := splitPath(s, t.separator(), false)

	// search for a match in this node
	for _, e := range t.n {
		if m, v := e.cmp.Matches(c); m {
			if e.sub == nil {
				if r == "" {
					if v != "" {
						vars[v] = c
					}
					return e.value, vars, true
				}
			} else {
				if r == "" {
					if e.value != nil {
						if v != "" {
							vars[v] = c
						}
						return e.value, vars, true
					} else {
						break
					}
				} else if val, x, ok := e.sub.find(r, vars); ok {
					if v != "" {
						vars[v] = c
					}
					return val, x, ok
				}
			}
		}
	}

	// not found
	return nil, nil, false
}

func (t *Tree) Iter(f func(string, interface{}) bool) {
	t.iter([]component{}, f)
}

func (t *Tree) iter(c []component, f func(p string, v interface{}) bool) bool {
	for _, e := range t.n {
		d := append(c, e.cmp)
		if v := e.value; v != nil {
			if !f(joinCmp(d, t.separator()), v) {
				return false
			}
		}
		if e.sub != nil {
			if !e.sub.iter(d, f) {
				return false
			}
		}
	}
	return true
}

func (t *Tree) Describe() string {
	return t.describe(&strings.Builder{}, 0).String()
}

func (t *Tree) describe(b *strings.Builder, d int) *strings.Builder {
	for _, e := range t.n {
		b.WriteString(strings.Repeat("  ", d))
		b.WriteString("- ")
		if e.cmp == "" {
			b.WriteRune(t.separator())
		} else {
			b.WriteString(string(e.cmp))
		}
		b.WriteString("\n")
		if e.sub != nil {
			e.sub.describe(b, d+1)
		}
	}
	return b
}
