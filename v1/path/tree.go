package path

import (
	"strings"
)

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

func (t *Tree) New(sep rune) *Tree {
	return &Tree{sep: sep}
}

func (t *Tree) separator() rune {
	if t.sep == 0 {
		return '/'
	} else {
		return t.sep
	}
}

func (t *Tree) Add(p string, v interface{}) {
	t.add(ParseSeparator(p, t.separator()).cmp, v)
}

func (t *Tree) add(p []component, v interface{}) {
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
			f.sub.add(p[1:], v)
		} else {
			f.value = v
		}
	}
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
				}
				if val, x, ok := e.sub.find(r, vars); ok {
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
