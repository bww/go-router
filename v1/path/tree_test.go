package path

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	tree := &Tree{}
	tree.Add("/")
	tree.Add("/a")
	tree.Add("/a/b")
	tree.Add("/a/b/c")
	tree.Add("/a/b/d")
	tree.Add("/a/b/e")
	tree.Add("/a/{var}")
	tree.Add("/a/{var}/c")

	fmt.Println(tree.Describe())

	tests := []struct {
		Path   string
		Expect bool
		Vars   Vars
	}{
		{
			"/a/b", true, Vars{},
		},
		{
			"/a/X", true, Vars{"var": "X"},
		},
		{
			"/a/Y/c", true, Vars{"var": "Y"},
		},
		{
			"/a/Y/d", false, nil,
		},
	}
	for _, e := range tests {
		_, v, ok := tree.Find(e.Path)
		assert.Equal(t, e.Expect, ok)
		assert.Equal(t, e.Vars, v)
	}

}
