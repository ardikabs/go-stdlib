package tool_test

import (
	"testing"

	"github.com/ardikabs/golib/tool"
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
