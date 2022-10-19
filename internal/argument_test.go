package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSoraArgument(t *testing.T) {
	arg := NewSoraArgument("sp", nil)
	assert.Equal(t, ArgReg, arg.Type)
	assert.Equal(t, "sp", arg.Reg)
	assert.False(t, arg.IsCodeLocation)

	arg = NewSoraArgument("0x14", nil)
	assert.Equal(t, ArgImm, arg.Type)
	assert.Equal(t, 20, arg.ValOfs)
	assert.False(t, arg.IsCodeLocation)

	arg = NewSoraArgument("->ra", nil)
	assert.True(t, arg.IsCodeLocation)
	assert.Equal(t, ArgReg, arg.Type)
	assert.Equal(t, "ra", arg.Reg)

	arg = NewSoraArgument("-0x14(sp)", nil)
	assert.Equal(t, ArgMem, arg.Type)
	assert.Equal(t, -20, arg.ValOfs)
	assert.Equal(t, "sp", arg.Reg)

	arg = NewSoraArgument("->$08a38a70", func(addr uint32) *string {
		assert.Equal(t, uint32(0x08a38a70), addr)
		label := "z_unknown"
		return &label
	})
	assert.True(t, arg.IsCodeLocation)
	assert.Equal(t, ArgImm, arg.Type)
	assert.Equal(t, "z_unknown", arg.Label)
}
