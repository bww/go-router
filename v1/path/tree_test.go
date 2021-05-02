package path

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	tree := &Tree{}
	tree.Add("/", "/")
	tree.Add("/a", "/a")
	tree.Add("/a/b", "/a/b")
	tree.Add("/a/b/c", "/a/b/c")
	tree.Add("/a/b/d", "/a/b/d")
	tree.Add("/a/{var}/e", "/a/{var}/e")
	tree.Add("/a/b/e", "/a/b/e")
	tree.Add("/a/{var}", "/a/{var}")
	tree.Add("/a/{var}/c", "/a/{var}/c")
	tree.Add("/a/*/*/d", "/a/*/*/d")
	tree.Add("/a/b/c/d/e/f/g", "/a/b/c/d/e/f/g")

	fmt.Println(tree.Describe())

	tests := []struct {
		Path   string
		Expect bool
		Value  interface{}
		Vars   Vars
	}{
		{
			"/a/b", true, "/a/b", Vars{},
		},
		{
			"/a/b/c", true, "/a/b/c", Vars{},
		},
		{
			"/a/X", true, "/a/{var}", Vars{"var": "X"},
		},
		{
			"/a/Y/c", true, "/a/{var}/c", Vars{"var": "Y"},
		},
		{
			"/a/Y/d", false, nil, nil,
		},
		{
			"/a/Y/X", false, nil, nil,
		},
		{
			"/a/Y/e", true, "/a/{var}/e", Vars{"var": "Y"},
		},
		{
			"/a/b/c/d", true, "/a/*/*/d", Vars{},
		},
		{
			"/a/b/c/d/e/f", false, nil, nil,
		},
	}
	for i, e := range tests {
		v, x, ok := tree.Find(e.Path)
		assert.Equal(t, e.Expect, ok, fmt.Sprintf("#%d: %s", i, e.Path))
		assert.Equal(t, e.Vars, x, fmt.Sprintf("#%d: %s", i, e.Path))
		assert.Equal(t, e.Value, v, fmt.Sprintf("#%d: %s", i, e.Path))
	}

}
