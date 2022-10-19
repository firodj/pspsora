package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAja(t *testing.T) {
	q := Queue[string]{}

	assert.Equal(t, 0, q.Len())
	q.Push("a")
	q.Push("b")

	assert.Equal(t, 2, q.Len())
	x := q.Pop()
	assert.Equal(t, 1, q.Len())
	assert.Equal(t, "b", x)

	x = q.Pop()
	assert.Equal(t, 0, q.Len())
	assert.Equal(t, "a", x)
}
