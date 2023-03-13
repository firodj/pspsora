package internal

type PseudoPrint func(instr *SoraInstruction) string

var mnemonicToPseudo map[string]PseudoPrint = map[string]PseudoPrint{
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

	"li": PseudoAssign,
}

func Code(instr *SoraInstruction) string {
	if fn, ok := mnemonicToPseudo[instr.Mnemonic]; ok {
		return fn(instr)
	}
	return ""
}

func PseudoAssign(instr *SoraInstruction) string {
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
