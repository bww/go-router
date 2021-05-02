package path

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	paths := []struct {
		Path  string
		Error error
	}{
		{"/", nil},
		{"/a", nil},
		{"/a/b", nil},
		{"/a/b/c", nil},
		{"/a/b/d", nil},
		{"/a/{var}/e", nil},
		{"/a/b/e", nil},
		{"/a/{var}", nil},
		{"/a/{var}/c", nil},
		{"/a/*/*/d", nil},
		{"/a/b/c/d/e/f/g", nil},

		{"/a/b", ErrCollision},
		{"/a/b/c", ErrCollision},
		{"/a/{var}/c", ErrCollision},
		{"/a/*/*/d", ErrCollision},
		{"/a/b/c/d/e/f/g", ErrCollision},
	}

	tree := &Tree{}
	for _, e := range paths {
		err := tree.Add(e.Path, e.Path)
		assert.Equal(t, e.Error, err)
	}

	fmt.Println(tree.Describe())

	tests := []struct {
		Path   string
		Expect bool
		Value  interface{}
		Vars   Vars
	}{
		{
			"/", true, "/", Vars{},
		},
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
