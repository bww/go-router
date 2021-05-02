package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaths(t *testing.T) {
	tests := []struct {
		Path   string
		Match  string
		Expect bool
		Vars   Vars
	}{
		{
			"/", "/", true, nil,
		},
		{
			"/a", "/a", true, nil,
		},
		{
			"/a/b", "/a/b", true, nil,
		},

		{
			"/a/{var}", "/a/b", true, map[string]string{"var": "b"},
		},
		{
			"/a/{var}/c", "/a/b/c", true, map[string]string{"var": "b"},
		},
		{
			"/a/{v/r}/c", "/a/b/c", true, map[string]string{"v/r": "b"},
		},
		{
			"/a/{var1}/c/{var2}", "/a/b/c/d", true, map[string]string{"var1": "b", "var2": "d"},
		},

		{
			"/", "/a", false, nil,
		},
		{
			"/a/b", "/a/c", false, nil,
		},
		{
			"/a/{var}", "/a/b/c", false, nil,
		},
		{
			"/a/c/{var}", "/a/b/c", false, nil,
		},
		{
			"/a/{var1}/{var2}", "/x/b/c", false, nil,
		},
		{
			"/a/{var1}/{var2}", "/a/b/c/d", false, nil,
		},

		{
			"/*", "/a", true, nil,
		},
		{
			"/*", "/", true, nil,
		},
		{
			"/a/*/c", "/a/b/c", true, nil,
		},
		{
			"/a/*", "/a/b", true, nil,
		},
		{
			"/a/*", "/a/b/c", false, nil,
		},
		{
			"/a/*", "/a/b/c/d", false, nil,
		},

		{
			"/a/**", "/a/b/c/d", true, nil,
		},
		{
			"/a/**", "/a", true, nil,
		},
		{
			"/a/**", "/", false, nil,
		},
		{
			"/a/**/c/d", "/a/b/c/d", true, nil,
		},
		{
			"/a/**/x/d", "/a/b/c/d", false, nil,
		},
		{
			"/**", "/", true, nil,
		},
		{
			"/**", "/a/b/c/d", true, nil,
		},
		{
			"/a/**", "/a/b/c/d", true, nil,
		},
	}
	for _, e := range tests {
		m, v := Parse(e.Path).Matches(e.Match)
		assert.Equal(t, e.Expect, m)
		assert.Equal(t, e.Vars, v)
	}
}

func TestPathsSep(t *testing.T) {
	sep := ':'
	tests := []struct {
		Path   string
		Match  string
		Expect bool
		Vars   Vars
	}{
		{
			":", ":", true, nil,
		},
		{
			"a", "a", true, nil,
		},
		{
			"a:b", "a:b", true, nil,
		},
		{
			"a:{var}", "a:b", true, map[string]string{"var": "b"},
		},
	}
	for _, e := range tests {
		m, v := ParseSeparator(e.Path, sep).Matches(e.Match)
		assert.Equal(t, e.Expect, m)
		assert.Equal(t, e.Vars, v)
	}
}
