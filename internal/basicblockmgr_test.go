package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	bbmanager := NewBasicBlockManager(nil)

	addr := uint32(0x800001)
	bb := bbmanager.Create(addr)
	bb.LastAddress = addr
	assert.NotNil(t, bb)

	bb2 := bbmanager.Create(addr)
	assert.Nil(t, bb2)
}

func TestGet(t *testing.T) {
	bbmanager := NewBasicBlockManager(nil)

	bb := bbmanager.Create(0x800020)
	bb.LastAddress = 0x80002C

	bb = bbmanager.Create(0x800010)
	bb.LastAddress = 0x80001C

	bb = bbmanager.Get(0x800000)
	assert.Nil(t, bb)

	bb = bbmanager.Get(0x800010)
	assert.NotNil(t, bb)
	assert.Equal(t, uint32(0x800010), bb.Address)

	bb = bbmanager.Get(0x800018)
	assert.NotNil(t, bb)
	assert.Equal(t, uint32(0x800010), bb.Address)

	bb = bbmanager.Get(0x800030)
	assert.Nil(t, bb)
}

func TestSplitAt(t *testing.T) {
	bbmanager := NewBasicBlockManager(nil)

	bb := bbmanager.Create(0x800008)
	bb.LastAddress = 0x80001C
	bb.BranchAddress = 0x800018

	prev, split := bbmanager.SplitAt(0x800014)

	assert.Equal(t, uint32(0x800008), prev.Address)
	assert.Equal(t, uint32(0x800014-4), prev.LastAddress)

	assert.Equal(t, uint32(0x800014), split.Address)
	assert.Equal(t, uint32(0x800018), split.BranchAddress)
	assert.Equal(t, uint32(0x80001C), split.LastAddress)
}
