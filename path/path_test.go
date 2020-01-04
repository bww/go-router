package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaths(t *testing.T) {
	var m bool
	var v map[string]string

	m, _ = Parse("/").Matches("/")
	assert.Equal(t, true, m)
	m, _ = Parse("/a").Matches("/a")
	assert.Equal(t, true, m)
	m, _ = Parse("/a/b").Matches("/a/b")
	assert.Equal(t, true, m)

	m, v = Parse("/a/{var}").Matches("/a/b")
	assert.Equal(t, true, m)
	assert.Equal(t, map[string]string{"var": "b"}, v)
	m, v = Parse("/a/{var}/c").Matches("/a/b/c")
	assert.Equal(t, true, m)
	assert.Equal(t, map[string]string{"var": "b"}, v)
	m, v = Parse("/a/{v/r}/c").Matches("/a/b/c")
	assert.Equal(t, true, m)
	assert.Equal(t, map[string]string{"v/r": "b"}, v)
	m, v = Parse("/a/{var1}/c/{var2}").Matches("/a/b/c/d")
	assert.Equal(t, true, m)
	assert.Equal(t, map[string]string{"var1": "b", "var2": "d"}, v)

	m, _ = Parse("/").Matches("/a")
	assert.Equal(t, false, m)
	m, _ = Parse("/a").Matches("/")
	assert.Equal(t, false, m)
	m, _ = Parse("/a/b").Matches("/a/c")
	assert.Equal(t, false, m)
	m, _ = Parse("/a/{var}").Matches("/a/b/c")
	assert.Equal(t, false, m)
	m, _ = Parse("/a/c/{var}").Matches("/a/b/c")
	assert.Equal(t, false, m)
	m, _ = Parse("/a/{var1}/{var2}").Matches("/x/b/c")
	assert.Equal(t, false, m)
	m, _ = Parse("/a/{var1}/{var2}").Matches("/a/b/c/d")
	assert.Equal(t, false, m)

	m, _ = Parse("/*").Matches("/a")
	assert.Equal(t, true, m)
	m, _ = Parse("/*").Matches("/")
	assert.Equal(t, true, m)
	m, _ = Parse("/a/*/c").Matches("/a/b/c")
	assert.Equal(t, true, m)
	m, _ = Parse("/a/*").Matches("/a/b")
	assert.Equal(t, true, m)
	m, _ = Parse("/a/*").Matches("/a/b/c")
	assert.Equal(t, false, m)
	m, _ = Parse("/a/*").Matches("/a/b/c/d")
	assert.Equal(t, false, m)
	m, _ = Parse("/a/**").Matches("/a/b/c/d")
	assert.Equal(t, true, m)
	m, _ = Parse("/a/**").Matches("/a")
	assert.Equal(t, true, m)
	m, _ = Parse("/a/**").Matches("/")
	assert.Equal(t, false, m)
	m, _ = Parse("/a/**/c/d").Matches("/a/b/c/d")
	assert.Equal(t, true, m)
	m, _ = Parse("/**").Matches("/")
	assert.Equal(t, true, m)
	m, _ = Parse("/**").Matches("/a/b/c/d")
	assert.Equal(t, true, m)
	m, _ = Parse("/a/**").Matches("/a/b/c/d")
	assert.Equal(t, true, m)

}
