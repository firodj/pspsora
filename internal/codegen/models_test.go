package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAja(t *testing.T) {

	s := ASTAssign{}
	s.Left = &ASTSymbolRef{
		ASTSymbol{
			Name: "a",
		},
	}

	a := ASTBinary{
		Op: "+",
	}

	a.Left = &ASTUnary{
		Op: "u32",
		Expr: &ASTSymbolRef{
			ASTSymbol{
				Name: "b",
			},
		},
	}

	a.Right = &ASTNumber{Value: 2}

	s.Right = &a

	assert.Equal(t, "a = u32(b) + 0x2", s.String())
}
