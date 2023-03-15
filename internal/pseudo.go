package internal

import (
	"fmt"
	"strconv"
)

type PseudoPrint func(instr *SoraInstruction, doc *SoraDocument) (string, int)

var mnemonicToPseudo map[string]PseudoPrint

func init() {
	mnemonicToPseudo = map[string]PseudoPrint{
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
}

// return
//   - 0: ok,
//   - -1: unimplemented,
//   - 1: skip delay shot.
func Code(instr *SoraInstruction, doc *SoraDocument) (string, int) {
	if fn, ok := mnemonicToPseudo[instr.Mnemonic]; ok {
		return fn(instr, doc)
	}
	return "", -1
}

func PseudoNothing(instr *SoraInstruction, doc *SoraDocument) (string, int) {
	return "// nop", 0
}

func PseudoAssign(instr *SoraInstruction, doc *SoraDocument) (string, int) {
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

	ss := instr.Args[0].Str(false) + " = "

	if arg1_signed {
		ss += "(s32)"
	}
	ss += instr.Args[1].Str(false)

	if len(instr.Args) > 2 && !instr.Args[2].IsZero() {
		if op != "" {
			ss += " " + op
		}

		ss += " "
		if arg2_signed {
			ss += "(s32)"
		}
		ss += instr.Args[2].Str(arg2_is_dec)
	}

	return ss, 0
}

func PseudoLoadUpper(instr *SoraInstruction, doc *SoraDocument) (string, int) {
	ss := instr.Args[0].Str(false) + " = " + instr.Args[1].Str(false)

	//if instr.Args[1].IsZero() {
	//}

	if instr.Args[1].IsNumber() {
		ss += "0000"
	} else {
		ss += " << 16"
	}

	return ss, 0
}

func PseudoLoad(instr *SoraInstruction, doc *SoraDocument) (string, int) {
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

	return ss, 0
}

func PseudoStore(instr *SoraInstruction, doc *SoraDocument) (string, int) {
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
	return ss, 0
}

var maskToType map[byte]string = map[byte]string{
	'v': "void",
	'x': "u32",
	'i': "s32",
	'f': "float",
	'X': "u64",
	'I': "s64",
	'F': "double",
	's': "const char*",
	'p': "(u32*)",
}

func PseudoSyscall(instr *SoraInstruction, doc *SoraDocument) (string, int) {
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
		return ss, 0
	}

	ss = fmt.Sprintf("m0x%02x::f0x%03x()", moduleIndex, funcIndex)
	return ss, -1

}

func PseudoJump(instr *SoraInstruction, doc *SoraDocument) (string, int) {
	jal_ra := instr.Address + 4
	if instr.Info.HasDelaySlot {
		jal_ra += 4
	}

	arg0 := instr.Args[0]

	var next_instr *SoraInstruction = nil
	jmp := false
	op := ""
	op_z := ""
	op_l := ""
	ss := ""

	if instr.Info.HasDelaySlot {
		next_instr = doc.InstrManager.Get(instr.Address + 4)
		if next_instr == nil {
			fmt.Printf("WARNING:\tPseudoJump delay shot instruction out of range\n")
			return "", -1
		}
	}

	switch instr.Mnemonic {
	case "jal":
		if next_instr != nil {
			_ss, _ := Code(next_instr, doc)
			ss += _ss + ";\n"
		}

		if arg0.IsZero() {
			return "", -1
		}

		ss += "v0 = " + arg0.Str(false) + "(...)"
		ss += fmt.Sprintf("\t/* { ra = 0x%08x; ", jal_ra)
		ss += fmt.Sprintf("goto %s; } */", arg0.CodeLabel(doc))

		jmp = true
	case "j":
		if next_instr != nil {
			_ss, _ := Code(next_instr, doc)
			ss += _ss + ";\n"
		}

		ss += fmt.Sprintf("goto %s;", arg0.CodeLabel(doc))

		jmp = true
	case "jr":
		if next_instr != nil {
			_ss, _ := Code(next_instr, doc)
			ss += _ss + ";\n"
		}

		if arg0.Type == ArgReg && arg0.Reg == "ra" {
			ss += "return v0"
			ss += "\t/* { goto -> ra; } */"
		} else {
			ss += fmt.Sprintf("goto %s;", arg0.Str(false))
		}

		jmp = true
	case "beq":
		op = "=="
	case "bne":
		op = "!="
	case "blez":
		op = "<= 0"
	case "bgtz":
		op_z = "> 0"
	case "bltz":
		op_z = "< 0"
	case "bgez":
		op_z = ">= 0"
	case "bnel":
		op_l = "!="
	case "beql":
		op_l = "=="
	}

	if !jmp {
		if op != "" {
			if next_instr != nil {
				_ss, _ := Code(next_instr, doc)
				ss += _ss + ";\n"
			}

			ss += "if (" + instr.Args[0].Str(false) + " " + op + " " + instr.Args[1].Str(false) + ") "
			ss += "goto " + instr.Args[2].CodeLabel(doc)
		} else if op_z != "" {
			if next_instr != nil {
				_ss, _ := Code(next_instr, doc)
				ss += _ss + ";\n"
			}

			ss += "if ((s32)" + instr.Args[0].Str(false) + " " + op_z + ") "
			ss += "goto " + instr.Args[1].CodeLabel(doc)

		} else if op_l != "" {
			ss += "if (" + instr.Args[0].Str(false) + " " + op_l + " " + instr.Args[1].Str(false) + ") {\n"
			if next_instr != nil {
				_ss, _ := Code(next_instr, doc)
				ss += "\t" + _ss + ";\n"
			}
			ss += "\tgoto " + instr.Args[2].Str(false) + "\n"
			ss += "}"

		} else {
			return "", -1
		}
	}

	if instr.Info.HasDelaySlot {
		return ss, 1
	}
	return ss, 0
}
