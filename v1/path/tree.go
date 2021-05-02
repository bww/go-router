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

func (t *Tree) Add(p string) {
	t.add(ParseSeparator(p, t.separator()).cmp)
}

func (t *Tree) add(p []component) {
	if len(p) > 0 {
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
			f = &node{cmp: p[0], sub: &Tree{}}
			t.n = append(t.n, f)
		}

		// recurse to it with the remaining components
		f.sub.add(p[1:])
	}
}

func (t *Tree) Find(s string) (Path, Vars, bool) {
	return t.find(s, []component{}, Vars{})
}

func (t *Tree) find(s string, p []component, vars Vars) (Path, Vars, bool) {
	c, r := splitPath(s, t.separator(), false)

	// search for a match in this node
	var n *node
	for _, e := range t.n {
		if m, v := e.cmp.Matches(c); m {
			if v != "" {
				vars[v] = c
			}
			n = e
			break
		}
	}
	if n == nil {
		return Path{}, nil, false
	}

	// append to our path and continue
	p = append(p, n.cmp)
	if r != "" {
		return n.sub.find(r, p, vars)
	} else {
		return Path{cmp: p, sep: t.separator()}, vars, true
	}
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
		e.sub.describe(b, d+1)
	}
	return b
}
