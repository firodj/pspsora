package internal

import (
	"fmt"
	"strings"
)

type SoraArgType string

const (
	ArgNone    SoraArgType = ""
	ArgImm     SoraArgType = "imm"
	ArgReg     SoraArgType = "reg"
	ArgMem     SoraArgType = "mem"
	ArgUnknown SoraArgType = "unk"
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

func (arg *SoraArgument) IsNegative() bool {
	return arg.Type == ArgImm && arg.ValOfs < 0
}

func (arg *SoraArgument) IsNumber() bool {
	return arg.Type == ArgImm
}

func (arg *SoraArgument) IsZero() bool {
	return (arg.Type == ArgImm && arg.ValOfs == 0) || (arg.Type == ArgReg && arg.Reg == "zero")
}
