package path

import (
	"errors"
	"strings"
)

var ErrCollision = errors.New("Path collision")

type node[T any] struct {
	cmp   component
	value T
	isset bool
	sub   *Tree[T]
}

// A path prefix tree
type Tree[T any] struct {
	n   []*node[T]
	sep rune
}

// Create a tree with the specified separator. The zero
// value of a tree uses the default separator: '/'.
func NewTree[T any](sep rune) *Tree[T] {
	return &Tree[T]{sep: sep}
}

func (t *Tree[T]) separator() rune {
	if t.sep == 0 {
		return '/'
	} else {
		return t.sep
	}
}

func (t *Tree[T]) Add(p string, v T) error {
	return t.add(ParseSeparator(p, t.separator()).cmp, v)
}

func (t *Tree[T]) add(p []component, v T) error {
	if l := len(p); l > 0 {
		var f *node[T]

		// attempt to find the first component
		for _, e := range t.n {
			if e.cmp.Equals(p[0]) {
				f = e
				break
			}
		}

		// create the node if we didn't find it
		if f == nil {
			f = &node[T]{cmp: p[0]}
			t.n = append(t.n, f)
		}

		// recurse to it with the remaining components or finalize
		if l > 1 {
			if f.sub == nil {
				f.sub = &Tree[T]{sep: t.sep}
			}
			return f.sub.add(p[1:], v)
		} else {
			if f.isset {
				return ErrCollision
			}
			f.value = v
			f.isset = true
		}
	}
	return nil
}

func (t *Tree[T]) Find(s string) (interface{}, Vars, bool) {
	return t.find(s, Vars{})
}

func (t *Tree[T]) find(s string, vars Vars) (interface{}, Vars, bool) {
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
					if e.isset {
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

func (t *Tree[T]) Iter(f func(string, T) bool) {
	t.iter([]component{}, f)
}

func (t *Tree[T]) iter(c []component, f func(p string, v T) bool) bool {
	for _, e := range t.n {
		d := append(c, e.cmp)
		if e.isset {
			if !f(joinCmp(d, t.separator()), e.value) {
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

func (t *Tree[T]) Describe() string {
	return t.describe(&strings.Builder{}, 0).String()
}

func (t *Tree[T]) describe(b *strings.Builder, d int) *strings.Builder {
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
