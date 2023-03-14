package internal

import (
	"fmt"
	"strconv"
)

type PseudoPrint func(instr *SoraInstruction, doc *SoraDocument) string

var mnemonicToPseudo map[string]PseudoPrint = map[string]PseudoPrint{
	"nop":   PseudoNothing,
	"addiu": PseudoAssign,
	"addu":  PseudoAssign,
	"subu":  PseudoAssign,

	"move":  PseudoAssign,
	"andi":  PseudoAssign,
	"ori":   PseudoAssign,
	"or":    PseudoAssign,
	"xor":   PseudoAssign,
	"sll":   PseudoAssign,
	"sltiu": PseudoAssign,
	"slti":  PseudoAssign,
	"sltu":  PseudoAssign,
	"slt":   PseudoAssign,
	"sra":   PseudoAssign,
	"srl":   PseudoAssign,

	"li":  PseudoAssign,
	"lui": PseudoLoadUpper,
	"lw":  PseudoLoad,
	"bu":  PseudoLoad,

	"sw": PseudoStore,
	"sb": PseudoStore,
	"sh": PseudoStore,

	"beq":  PseudoJump,
	"beql": PseudoJump,
	"bne":  PseudoJump,
	"bnel": PseudoJump,
	"blez": PseudoJump,
	"bgtz": PseudoJump,
	"bltz": PseudoJump,
	"bgez": PseudoJump,

	"jr":  PseudoJump,
	"j":   PseudoJump,
	"jal": PseudoJump,

	"syscall": PseudoSyscall,
}

func Code(instr *SoraInstruction, doc *SoraDocument) string {
	if fn, ok := mnemonicToPseudo[instr.Mnemonic]; ok {
		return fn(instr, doc)
	}
	return ""
}

func PseudoNothing(instr *SoraInstruction, doc *SoraDocument) string {
	return "\t//"
}

func PseudoAssign(instr *SoraInstruction, doc *SoraDocument) string {
	op := ""
	arg2_is_dec := false
	arg1_signed := false
	arg2_signed := false

	switch instr.Mnemonic {
	case "addiu", "addu", "li":
		op = "+"
	case "subu":
		op = "-"
	case "andi":
		op = "&"
	case "ori", "or":
		op = "|"
	case "xor":
		op = "^"
	case "sll":
		op = "<<"
		arg2_is_dec = true
	case "sltiu", "sltu":
		op = "<"
	case "slti":
		op = "<"
		arg1_signed = true
	case "slt":
		op = "<"
		arg1_signed = true
		arg2_signed = true
	case "sra":
		op = ">>"
		arg1_signed = true
		arg2_is_dec = true
	case "srl":
		op = ">>"
		arg2_is_dec = true
	}

	if len(instr.Args) > 2 && op == "" {
		panic("invalid number of argument")
	}

	if len(instr.Args) > 3 {
		panic("invalid number of argument")
	}

	s := instr.Args[0].Str(false) + " = "

	if arg1_signed {
		s += "(s32)"
	}
	s += instr.Args[1].Str(false)

	if len(instr.Args) > 2 && !instr.Args[2].IsZero() {
		if op != "" {
			s += " " + op
		}

		s += " "
		if arg2_signed {
			s += "(s32)"
		}
		s += instr.Args[2].Str(arg2_is_dec)
	}

	return "\t; " + s
}

func PseudoLoadUpper(instr *SoraInstruction, doc *SoraDocument) string {
	ss := instr.Args[0].Str(false) + " = " + instr.Args[1].Str(false)

	//if instr.Args[1].IsZero() {
	//}

	if instr.Args[1].IsNumber() {
		ss += "0000"
	} else {
		ss += " << 16"
	}

	return "\t; " + ss
}

func PseudoLoad(instr *SoraInstruction, doc *SoraDocument) string {
	suffix := instr.Mnemonic[1:]
	sz := ""
	mask := ""

	if suffix == "w" {
		sz = "u32"
	} else if suffix == "bu" {
		sz = "u8"
		mask = " & 0xff"
	} else {
		panic("unknown suffix")
	}

	ss := instr.Args[0].Str(false) + " = (" + sz + ")" + instr.Args[1].Str(false)
	if mask != "" {
		ss += mask
	}

	return "\t; " + ss
}

func PseudoStore(instr *SoraInstruction, doc *SoraDocument) string {
	sz := ""
	suffix := instr.Mnemonic[1:]
	if suffix == "b" {
		sz = "u8"
	} else if suffix == "h" {
		sz = "u16"
	} else if suffix == "w" {
		sz = "u32"
	}

	ss := "(" + sz + ")" + instr.Args[1].Str(false) + " = " + instr.Args[0].Str(false)
	return "\t; " + ss
}

var maskToType map[byte]string = map[byte]string{
	'v': "void",
	'x': "u32",
	'i': "int",
	'f': "float",
	'X': "u64",
	'I': "int64",
	'F': "double",
	's': "const char*",
	'p': "(u32*)",
}

func PseudoSyscall(instr *SoraInstruction, doc *SoraDocument) string {
	moduleIndex, funcIndex := instr.GetSyscallNumber()

	ss := ""

	_, fun := doc.GetHLE(moduleIndex, funcIndex)
	if fun != nil {

		if len(fun.RetMask) > 0 {
			ss += "v0 = (" + maskToType[fun.RetMask[0]] + ")"
		}

		ss += doc.GetHLEFuncName(moduleIndex, funcIndex) + "("
		for arg_i := range fun.ArgMask {
			if arg_i > 0 {
				ss += ", "
			}
			ss += "(" + maskToType[fun.ArgMask[arg_i]] + ")" + doc.GetRegName(0, 4+arg_i)
		}

		for ret_i := range fun.RetMask {
			if ret_i == 0 {
				continue
			}

			if ret_i > 1 {
				ss += ", "
			}
			ss += "(*" + maskToType[fun.RetMask[ret_i]] + ")&v" + strconv.Itoa(ret_i)
		}

		ss += ")"
		// 0
	} else {
		ss = fmt.Sprintf("m0x%02x::f0x%03x()", moduleIndex, funcIndex)
		// -1
	}

	return "\t; " + ss
}

func PseudoJump(instr *SoraInstruction, doc *SoraDocument) string {
	return "\t; TODO"
}
