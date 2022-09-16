package tool_test

import (
	"regexp"
	"testing"

	"github.com/ardikabs/golib/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestIn(t *testing.T) {

	t.Run("true", func(t *testing.T) {
		val := tool.In("woman", "man", "woman")
		assert.True(t, val)
	})

	t.Run("false", func(t *testing.T) {
		val := tool.In(5, 1, 2, 3, 4)
		assert.False(t, val)
	})
}

func TestMatches(t *testing.T) {

	t.Run("true", func(t *testing.T) {
		fakeRX := regexp.MustCompile(`^\w+$`)
		val := tool.Matches("abcdefg", fakeRX)
		assert.True(t, val)
	})

	t.Run("false", func(t *testing.T) {
		fakeEmailRX := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
		val := tool.Matches("abcdefg", fakeEmailRX)
		assert.False(t, val)
	})
}

func TestUnique(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		val := tool.Unique([]string{"a", "b", "c", "d"})
		assert.True(t, val)
	})

	t.Run("false", func(t *testing.T) {
		val := tool.Unique([]int{1, 1, 1, 1, 1, 2, 2, 2, 3, 3})
		assert.False(t, val)
	})
}

func TestRFC3339(t *testing.T) {
	t.Run("true for UTC", func(t *testing.T) {
		val := tool.RFC3339("1996-12-25T18:34:05Z")
		assert.True(t, val)
	})

	t.Run("true for UTC+7", func(t *testing.T) {
		val := tool.RFC3339("1996-12-25T18:34:05+07:00")
		assert.True(t, val)
	})

	t.Run("true for UTC+5:30", func(t *testing.T) {
		val := tool.RFC3339("1996-12-25T18:34:05+05:30")
		assert.True(t, val)
	})

	t.Run("false", func(t *testing.T) {
		val := tool.RFC3339("1996-12-25")
		assert.False(t, val)
	})
}
