package internal

import (
	"fmt"
	"strings"

	"github.com/firodj/pspsora/internal/codegen"
)

type SoraArgType string

const (
	ArgNone SoraArgType = ""
	ArgImm  SoraArgType = "imm"
	ArgReg  SoraArgType = "reg"
	ArgMem  SoraArgType = "mem"
)

type SoraArgument struct {
	Type           SoraArgType
	Label          string
	ValOfs         int
	Reg            string
	IsCodeLocation bool
}

func NewSoraArgument(opr string, labellookup func(uint32) *string) (arg *SoraArgument) {
	if len(opr) == 2 {
		return &SoraArgument{
			Type: ArgReg,
			Reg:  opr,
		}
	}

	arg = &SoraArgument{}

	lookup := false
	if strings.HasPrefix(opr, "->$") {
		opr = "0x" + opr[3:]
		arg.IsCodeLocation = true
		lookup = true
	} else if strings.HasPrefix(opr, "->") {
		opr = opr[2:]
		arg.IsCodeLocation = true
	}

	var imm int
	var rs string

	n, _ := fmt.Sscanf(opr, "%v(%s)", &imm, &rs)
	if n >= 1 {
		arg.Type = ArgImm
		arg.ValOfs = imm
		if n >= 2 {
			arg.Type = ArgMem
			arg.Reg = rs[:len(rs)-1]
		}
	} else {
		arg.Type = ArgReg
		arg.Reg = opr
	}

	if lookup {
		if labellookup != nil {
			label := labellookup(uint32(arg.ValOfs))
			if label != nil {
				arg.Label = *label
			}
		}
	}
	return
}

func (arg *SoraArgument) ValueStr(isDec bool) string {
	ss := ""
	n := arg.ValOfs

	if arg.ValOfs < 0 {
		ss += "-"
		n = -n
	}

	if !isDec {
		ss += fmt.Sprintf("0x%x", n)
	} else {
		ss += fmt.Sprintf("%d", n)
	}

	return ss
}

func (arg *SoraArgument) Str(isDec bool) string {
	ss := ""

	switch arg.Type {
	case ArgImm:
		if arg.Label != "" {
			ss += arg.Label
		} else {
			ss += arg.ValueStr(isDec)
		}
	case ArgReg:
		if arg.Reg == "zero" {
			ss += "0"
		} else {
			ss += arg.Reg
		}
	case ArgMem:
		ss += "[" + arg.Reg
		if arg.ValOfs != 0 {
			ss += " + " + arg.ValueStr(isDec)
		}
		ss += "]"
	default:
		ss += "??"
	}

	return ss
}

func (arg *SoraArgument) CodeLabel(doc *SoraDocument) string {
	if arg.IsCodeLocation {
		return doc.GetLabelName(uint32(arg.ValOfs))
	}
	return arg.Str(false)
}

func (arg *SoraArgument) IsNegative() bool {
	return arg.Type == ArgImm && arg.ValOfs < 0
}

func (arg *SoraArgument) IsNumber() bool {
	return arg.Type == ArgImm
}

func (arg *SoraArgument) IsZero() bool {
	return (arg.Type == ArgImm && arg.ValOfs == 0) || (arg.Type == ArgReg && arg.Reg == "zero")
}

func (arg *SoraArgument) ToPseudo() codegen.ASTNode {
	switch arg.Type {
	case ArgImm:
		if arg.Label != "" {
			e := codegen.ASTSymbolRef{}
			e.Name = arg.Label
			return &e
		} else {
			e := codegen.ASTNumber{
				Value: arg.ValOfs,
			}
			return &e
		}
	case ArgReg:
		if arg.Reg == "zero" {
			e := codegen.ASTNumber{
				Value: arg.ValOfs,
			}
			return &e
		} else {
			e := codegen.ASTSymbolRef{}
			e.Name = arg.Reg
			return &e
		}
	case ArgMem:
		b := codegen.ASTSymbolRef{}
		b.Name = arg.Reg
		e := codegen.ASTPointer{
			Sz:   "u32",
			Expr: &b,
		}
		if arg.ValOfs != 0 {
			n := codegen.ASTNumber{
				Value: arg.ValOfs,
			}
			s := codegen.ASTBinary{
				Op:    "+",
				Left:  &b,
				Right: &n,
			}
			e.Expr = &s
		}
		return &e
	}

	panic("unknown argument type")

	//return nil
}
