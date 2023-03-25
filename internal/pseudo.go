package internal

import (
	"fmt"
	"strconv"

	"github.com/firodj/pspsora/internal/codegen"
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
		"lbu": PseudoLoad,

		"sw": PseudoStore,
		"sb": PseudoStore,
		"sh": PseudoStore,

		"beq":  PseudoCondJump,
		"beql": PseudoCondJump,
		"bne":  PseudoCondJump,
		"bnel": PseudoCondJump,
		"blez": PseudoCondJump,
		"bgtz": PseudoCondJump,
		"bltz": PseudoCondJump,
		"bgez": PseudoCondJump,

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
	//arg2_is_dec := false
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
		//arg2_is_dec = true
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
		//arg2_is_dec = true
	case "srl":
		op = ">>"
		//arg2_is_dec = true
	}

	if len(instr.Args) > 2 && op == "" {
		panic("invalid number of argument")
	}

	if len(instr.Args) > 3 {
		panic("invalid number of argument")
	}

	s := codegen.ASTAssign{}

	s_left := instr.Args[0].ToPseudo()

	s_right := instr.Args[1].ToPseudo()

	if arg1_signed {
		s_right = &codegen.ASTUnary{
			Op:   "s32",
			Expr: s_right,
		}
	}

	if len(instr.Args) > 2 && !instr.Args[2].IsZero() {
		// Binary
		if op == "" {
			panic("missing op")
		}

		s_right1 := codegen.ASTBinary{
			Op:   op,
			Left: s_right,
		}

		s_right2 := instr.Args[2].ToPseudo()

		if arg2_signed {
			s_right2 = &codegen.ASTUnary{
				Op:   "s32",
				Expr: s_right2,
			}
		}

		s_right1.Right = s_right2
		s_right = &s_right1
	}

	s.Left = s_left
	s.Right = s_right

	return "--- " + s.String(), 0
}

func PseudoLoadUpper(instr *SoraInstruction, doc *SoraDocument) (string, int) {
	s := codegen.ASTAssign{}

	s_left := instr.Args[0].ToPseudo()

	var s_right codegen.ASTNode

	s_right1 := instr.Args[1].ToPseudo()

	s_right2 := &codegen.ASTNumber{
		Value: 16,
	}

	s_right = &codegen.ASTBinary{
		Op:    "<<",
		Left:  s_right1,
		Right: s_right2,
	}
	s.Right = s_right

	s.Left = s_left
	s.Right = s_right

	return "--- " + s.String(), 0
}

func PseudoLoad(instr *SoraInstruction, doc *SoraDocument) (string, int) {
	s := codegen.ASTAssign{}

	s_left := instr.Args[0].ToPseudo()
	s_right := instr.Args[1].ToPseudo()

	suffix := instr.Mnemonic[1:]
	switch suffix {
	case "w":
		s_right.(*codegen.ASTPointer).Sz = "u32"
	case "bu":
		s_right.(*codegen.ASTPointer).Sz = "u8"
		//mask = " & 0xff"
	case "hu":
		s_right.(*codegen.ASTPointer).Sz = "u16"
	case "b":
		s_right.(*codegen.ASTPointer).Sz = "s8"
		//mask = " & 0xff"
	case "h":
		s_right.(*codegen.ASTPointer).Sz = "s16"
	default:
		panic("unknown suffix: " + suffix)
	}

	s.Left = s_left
	s.Right = s_right

	return "--- " + s.String(), 0
}

func PseudoStore(instr *SoraInstruction, doc *SoraDocument) (string, int) {
	s := codegen.ASTAssign{}

	s_left := instr.Args[1].ToPseudo()
	s_right := instr.Args[0].ToPseudo()

	suffix := instr.Mnemonic[1:]
	switch suffix {
	case "b":
		s_left.(*codegen.ASTPointer).Sz = "u8"
	case "h":
		s_left.(*codegen.ASTPointer).Sz = "u16"
	case "w":
		s_left.(*codegen.ASTPointer).Sz = "u32"
	default:
		panic("unknown suffix: " + suffix)
	}

	s.Left = s_left
	s.Right = s_right

	return "--- " + s.String(), 0
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

	//ss := ""

	_, fun := doc.GetHLE(moduleIndex, funcIndex)
	if fun != nil {
		s_name := codegen.ASTSymbolRef{}
		s_name.Name = doc.GetHLEFuncName(moduleIndex, funcIndex)
		s := codegen.ASTCall{
			Expr: &s_name,
		}

		for arg_i := range fun.ArgMask {
			//if arg_i > 0 {
			//	ss += ", "
			//}
			a_name := codegen.ASTSymbolRef{}
			a_name.Name = doc.GetRegName(0, 4+arg_i)
			arg := codegen.ASTUnary{
				Op:   maskToType[fun.ArgMask[arg_i]],
				Expr: &a_name,
			}
			//ss += "(" + maskToType[fun.ArgMask[arg_i]] + ")" + doc.GetRegName(0, 4+arg_i)
			s.Args = append(s.Args, &arg)
		}

		if len(fun.RetMask) > 0 {
			sa := codegen.ASTAssign{}
			sa.Right = &s

			for ret_i := range fun.RetMask {
				if ret_i == 0 {
					s_left := codegen.ASTSymbolRef{}
					s_left.Name = "v0"
					sa.Left = &s_left
					continue
				}

				//if ret_i > 1 {
				//	ss += ", "
				//}
				//ss += "(*" + maskToType[fun.RetMask[ret_i]] + ")&v" + strconv.Itoa(ret_i)
				ret_a := codegen.ASTSymbolRef{}
				ret_a.Name = "v" + strconv.Itoa(ret_i)
				arg := codegen.ASTUnary{
					Op:   "*" + maskToType[fun.RetMask[ret_i]] + "&",
					Expr: &ret_a,
				}
				s.Args = append(s.Args, &arg)
			}

			//s_left := codegen.ASTSymbolRef{}
			//s_left.Name = "v0"

			//ss += "v0 = (" + maskToType[fun.RetMask[0]] + ")"
			//s.Left = &s_left
		}

		//ss += +"("

		//ss += ")"
		return "--- " + s.String(), 0
	}

	s_name := codegen.ASTSymbolRef{}
	s_name.Name = fmt.Sprintf("m0x%02x::f0x%03x()", moduleIndex, funcIndex)
	s := codegen.ASTCall{
		Expr: &s_name,
	}

	return "--- " + s.String(), 0
}

func PseudoJump(instr *SoraInstruction, doc *SoraDocument) (string, int) {
	jal_ra := instr.Address + 4
	if instr.Info.HasDelaySlot {
		jal_ra += 4
	}

	arg0 := instr.Args[0]
	ss := ""
	ss_next := ""

	if instr.Info.HasDelaySlot {
		next_instr := doc.InstrManager.Get(instr.Address + 4)
		if next_instr == nil {
			fmt.Printf("WARNING:\tPseudoJump delay shot instruction out of range\n")
			return "", -1
		}
		ss_next, _ = Code(next_instr, doc)
	}

	switch instr.Mnemonic {
	case "jal":
		if ss_next != "" {
			ss += ss_next + "\n"
		}

		if arg0.IsZero() {
			return "", -1
		}

		ss += "v0 = " + arg0.Str(false) + "(...);"
		ss += fmt.Sprintf("\t/* { ra = 0x%08x; ", jal_ra)
		ss += fmt.Sprintf("goto %s; } */", arg0.CodeLabel(doc))

	case "j":
		if ss_next != "" {
			ss += ss_next + "\n"
		}
		ss += fmt.Sprintf("goto %s;", arg0.CodeLabel(doc))
	case "jr":
		if ss_next != "" {
			ss += ss_next + "\n"
		}
		if arg0.Type == ArgReg && arg0.Reg == "ra" {
			ss += "return v0;"
			ss += "\t/* { goto -> ra; } */"
		} else {
			ss += fmt.Sprintf("goto %s;", arg0.Str(false))
		}
	default:
		panic("unknown jump")
	}

	if instr.Info.HasDelaySlot {
		return ss, 1
	}
	return ss, 0
}

func PseudoCondJump(instr *SoraInstruction, doc *SoraDocument) (string, int) {
	j_else := instr.Address + 4
	then_taken := false
	else_taken := false

	arg0 := instr.Args[0]

	op := ""
	op_l := ""
	ss := ""
	ss_next := ""

	var arg1 *SoraArgument = nil
	if len(instr.Args) >= 2 {
		arg1 = instr.Args[1]
	}

	if instr.Info.HasDelaySlot {
		next_instr := doc.InstrManager.Get(instr.Address + 4)
		if next_instr == nil {
			fmt.Printf("WARNING:\tPseudoCondJump delay shot instruction out of range\n")
			return "", -1
		}
		j_else += 4
		ss_next, _ = Code(next_instr, doc)
	} else {
		panic("unimplmented without delayslot")
	}

	theBB := doc.BBManager.Get(instr.Address)
	if theBB == nil {
		fmt.Printf("ERROR:\tunable to get BB for: 0x%08x", instr.Address)
		return "", -1
	}
	xref_tos := doc.BBManager.GetExitRefs(theBB.Address)
	for _, xref := range xref_tos {
		if xref.To == j_else {
			else_taken = true
		} else if !xref.IsAdjacent {
			then_taken = true
		}
	}

	switch instr.Mnemonic {
	case "beq":
		op = "=="
	case "bne":
		op = "!="
	case "blez":
		arg1 = NewSoraArgument("0", nil)
		op = "<="
	case "bgtz":
		arg1 = NewSoraArgument("0", nil)
		op = ">"
	case "bltz":
		arg1 = NewSoraArgument("0", nil)
		op = "<"
	case "bgez":
		arg1 = NewSoraArgument("0", nil)
		op = ">="
	case "bnel":
		op_l = "!="
	case "beql":
		op_l = "=="
	default:
		panic("unknown conditional jump")
	}

	argElse := NewSoraArgument(fmt.Sprintf("->$%08x", j_else), doc.SymMap.GetLabelName)

	if op != "" {
		ss += "zf = " + arg0.Str(false) + " " + op + " " + arg1.Str(false) + ";\n"
		ss += ss_next + ";\n"

		if then_taken && else_taken {
			ss += "if (zf) { goto " + instr.Args[2].CodeLabel(doc) + "; }"
		} else if then_taken {
			ss += "assert(zf); goto " + instr.Args[2].CodeLabel(doc) + ";"
		} else {
			ss += "assert(!zf); // goto " + argElse.CodeLabel(doc) + ";"
		}
	} else if op_l != "" {
		ss += "zf = " + arg0.Str(false) + " " + op_l + " " + arg1.Str(false) + ";\n"

		if then_taken && else_taken {
			ss += "if (!zf) { goto " + argElse.CodeLabel(doc) + "; }\n"
			ss += ss_next + "\n"
			ss += "goto " + instr.Args[2].CodeLabel(doc) + ";"
		} else if then_taken {
			ss += "assert(zf);\n"
			ss += ss_next + "\n"
			ss += "goto " + instr.Args[2].CodeLabel(doc) + ";"
		} else {
			ss += "assert(!zf); // goto " + argElse.CodeLabel(doc) + ";"
		}
	} else {
		return "", -1
	}

	if instr.Info.HasDelaySlot {
		return ss, 1
	}

	return ss, 0
}
