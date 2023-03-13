package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDizz(t *testing.T) {
	doc := newSoraDocument()
	doc.SymMap.AddModule("kernel", 0x8804000, 0x29a800)
	doc.FunManager.CreateNewFunction(0x08a38a70, 8)

	t.Run("when addiu", func(t *testing.T) {
		dizz := "addiu\tsp, sp, -0x20"
		m, args := doc.ParseDizz(dizz)

		assert.Equal(t, "addiu", m)
		assert.Len(t, args, 3)

		assert.Equal(t, ArgReg, args[0].Type)
		assert.Equal(t, "sp", args[0].Reg)
		assert.Equal(t, ArgReg, args[1].Type)
		assert.Equal(t, "sp", args[1].Reg)
		assert.Equal(t, ArgImm, args[2].Type)
		assert.Equal(t, -32, args[2].ValOfs)
	})

	t.Run("when jr ra", func(t *testing.T) {
		dizz := "jr\t->ra"
		m, args := doc.ParseDizz(dizz)

		assert.Equal(t, "jr", m)
		assert.Len(t, args, 1)

		assert.Equal(t, ArgReg, args[0].Type)
		assert.Equal(t, "ra", args[0].Reg)
	})

	t.Run("when syscall", func(t *testing.T) {
		dizz := "syscall\tSysMemUserForUser::sceKernelSetCompiledSdkVersion380_390"
		m, args := doc.ParseDizz(dizz)

		assert.Equal(t, "syscall", m)
		assert.Len(t, args, 1)

		assert.Equal(t, ArgUnknown, args[0].Type)
		assert.Equal(t, "SysMemUserForUser::sceKernelSetCompiledSdkVersion380_390", args[0].Label)
	})

	t.Run("when jal tgt", func(t *testing.T) {
		dizz := "jal\t->$08a38a70"
		m, args := doc.ParseDizz(dizz)

		assert.Equal(t, "jal", m)
		assert.Len(t, args, 1)

		assert.Equal(t, ArgImm, args[0].Type)
		assert.True(t, args[0].IsCodeLocation)
		assert.Equal(t, 0x08a38a70, args[0].ValOfs)
		assert.Equal(t, "z_un_08a38a70", args[0].Label)
	})

	t.Run("when beq", func(t *testing.T) {
		dizz := "beq\tt6, zero, ->$088041dc"
		m, args := doc.ParseDizz(dizz)

		assert.Equal(t, "beq", m)
		assert.Len(t, args, 3)

		assert.Equal(t, ArgReg, args[0].Type)
		assert.Equal(t, "t6", args[0].Reg)
		assert.Equal(t, ArgReg, args[1].Type)
		assert.Equal(t, "zero", args[1].Reg)
		assert.Equal(t, ArgImm, args[2].Type)
		assert.True(t, args[2].IsCodeLocation)
		assert.Equal(t, 0x088041dc, args[2].ValOfs)
	})
}
