package internal

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestFindFirstNull(t *testing.T) {
	s5 := []byte{0x73, 0x74, 0x61, 0x72, 0x74, 0x00, 0x00}
	assert.Equal(t, 5, FindFirstNull(s5))

	s3 := []byte{0x73, 0x74, 0x61, 0x00, 0x00, 0x00, 0x00}
	assert.Equal(t, 3, FindFirstNull(s3))

	s1 := []byte{0x73, 0x74, 0x61, 0x50, 0x61, 0x20, 0x00}
	assert.Equal(t, 6, FindFirstNull(s1))

	s0 := []byte{0x00, 0x00, 0x00, 0x00}
	assert.Equal(t, 0, FindFirstNull(s0))

	s2 := []byte{0x60, 0x70, 0x50, 0x60}
	assert.Equal(t, 4, FindFirstNull(s2))

	se := []byte{}
	assert.Equal(t, -1, FindFirstNull(se))
}

func TestBytesToString(t *testing.T) {
	s := []byte{0x73, 0x74, 0x61, 0x72, 0x74, 0x00, 0x00}

	z := string(s)

	spew.Dump(z)
	assert.True(t, true)
}
