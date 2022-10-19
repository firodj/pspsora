#include <iostream>
#include <sstream>

#include "Core/MIPS/MIPSDebugInterface.h"
#include "MyDocument.hpp"
#include "Pseudo.hpp"

struct PseudoTableRow {
  const char *mnemonic;
  int (MyDocument::*funcPseudo)(MyInstruction *instr);
};

static const std::map<std::string, PseudoTableRow> mapTables = {
  { "addiu", { "addiu", &MyDocument::PseuDoAssign } },
  { "addu", { "addu", &MyDocument::PseuDoAssign } },
  { "subu", { "subu", &MyDocument::PseuDoAssign } },

  { "move", { "move", &MyDocument::PseuDoAssign } },
  { "andi", { "andi", &MyDocument::PseuDoAssign } },
  { "ori", { "ori", &MyDocument::PseuDoAssign } },
  { "or", { "or", &MyDocument::PseuDoAssign } },
  { "xor", { "xor", &MyDocument::PseuDoAssign } },
  { "sll", { "sll", &MyDocument::PseuDoAssign } },
  { "sltiu", { "sltiu", &MyDocument::PseuDoAssign } },
  { "slti", { "slti", &MyDocument::PseuDoAssign } },
  { "sltu", { "sltu", &MyDocument::PseuDoAssign } },
  { "slt", { "slt", &MyDocument::PseuDoAssign } },
  { "sra", { "sra", &MyDocument::PseuDoAssign } },
  { "srl", { "sra", &MyDocument::PseuDoAssign } },

  { "lui", { "lui", &MyDocument::PseuDoLoadUpper } },
  { "lw", { "lw", &MyDocument::PseuDoLoad } },
  { "lbu", { "lbu", &MyDocument::PseuDoLoad }},
  { "li", { "li", &MyDocument::PseuDoAssign } },

  { "sw", { "sw", &MyDocument::PseuDoStore } },
  { "sb", { "sb", &MyDocument::PseuDoStore } },
  { "sh", { "sh", &MyDocument::PseuDoStore } },

  { "nop", { "nop", &MyDocument::PseuDoNothing } },

  { "beq", { "beq", &MyDocument::PseuDoJump } },
  { "beql", { "beql", &MyDocument::PseuDoJump } },
  { "bne", { "bne", &MyDocument::PseuDoJump } },
  { "bnel", { "bnel", &MyDocument::PseuDoJump } },
  { "blez", { "blez", &MyDocument::PseuDoJump } },
  { "bgtz", { "bgtz", &MyDocument::PseuDoJump } },
  { "bltz", { "bltz", &MyDocument::PseuDoJump } },
  { "bgez", { "bgez", &MyDocument::PseuDoJump } },

  { "jr", { "jr", &MyDocument::PseuDoJump } },
  { "j", { "j", &MyDocument::PseuDoJump } },
  { "jal", { "jal", &MyDocument::PseuDoJump } },

  { "syscall", { "syscall", &MyDocument::PseudoSyscall }}
};

int MyDocument::PseuDoNothing(MyInstruction *instr) {
  std::cout << "\t//" << std::endl;
  return 0;
}

int MyDocument::PseuDoAssign(MyInstruction *instr) {
  std::string op = "";
  bool arg2_is_dec = false;
  bool arg1_signed = false;
  bool arg2_signed = false;

  if (instr->mnemonic_ == "addiu") { op = "+"; }
  else if (instr->mnemonic_ == "addu" || instr->mnemonic_ == "li") { op = "+"; }
  else if (instr->mnemonic_ == "subu") { op = "-"; }
  else if (instr->mnemonic_ == "andi") { op = "&"; }
  else if (instr->mnemonic_ == "ori") { op = "|"; }
  else if (instr->mnemonic_ == "or") { op = "|"; }
  else if (instr->mnemonic_ == "xor") { op = "^"; }
  else if (instr->mnemonic_ == "sll") { op = "<<"; arg2_is_dec = true; }
  else if (instr->mnemonic_ == "sltiu") { op = "<"; }
  else if (instr->mnemonic_ == "sltu") { op = "<"; }
  else if (instr->mnemonic_ == "slti") { op = "<"; arg1_signed = true; }
  else if (instr->mnemonic_ == "slt") { op = "<"; arg1_signed = true; arg2_signed = true; }
  else if (instr->mnemonic_ == "sra") { op = ">>"; arg1_signed = true; arg2_is_dec = true; }
  else if (instr->mnemonic_ == "srl") { op = ">>"; arg2_is_dec = true; }

  if (instr->arguments_.size() > 2 && op.empty()) { return -1; }
  if (instr->arguments_.size() > 3) { return -1; }

  std::cout << '\t' << instr->arguments_[0].Str() << " = ";
  if (arg1_signed) {
    std::cout << "(s32)";
  }
  std::cout << instr->arguments_[1].Str();

  if (instr->arguments_.size() > 2 && !instr->arguments_[2].IsZero()) {
    if (!op.empty()) {
      std::cout << " " << op;
    }

    std::cout << " ";
    if (arg2_signed) {
      std::cout << "(s32)";
    }
    std::cout << instr->arguments_[2].Str(arg2_is_dec);
  }

  //std::cout << "\t// " << instr->AsString();
  std::cout << std::endl;

  return 0;
}

int MyDocument::PseuDoLoadUpper(MyInstruction *instr) {
  std::stringstream ss;
  ss << instr->arguments_[0].Str() << " = " << instr->arguments_[1].Str();

  if (instr->arguments_[1].IsZero()) { ; }
  if (instr->arguments_[1].IsNumber()) { ss << "0000"; }
  else { ss << " << 16"; }

  std::cout << '\t' << ss.str() << std::endl;
  return 0;
}

int MyDocument::PseuDoLoad(MyInstruction *instr) {
  std::string suffix = instr->mnemonic_.substr(1);
  std::string sz = "", mask = "";

  if (suffix == "w") { sz = "u32"; }
  else if (suffix == "bu") { sz = "u8"; mask = " & 0xff"; }
  else return -1;

  std::cout << "\t" << instr->arguments_[0].Str() << " = ("<<sz<<")" << instr->arguments_[1].Str();
  if (!mask.empty()) std::cout << mask;
  std::cout << std::endl;

  return 0;
}

int MyDocument::PseuDoStore(MyInstruction *instr) {
  std::string sz = "";
  if (instr->mnemonic_.back() == 'b') { sz = "u8"; }
  if (instr->mnemonic_.back() == 'h') { sz = "u16"; }
  if (instr->mnemonic_.back() == 'w') { sz = "u32"; }
  std::cout << "\t("<<sz<<")" << instr->arguments_[1].Str() << " = " << instr->arguments_[0].Str() << std::endl;
  return 0;
}

// return  0: ok,
//        -1: unimplemented,
//         1: skip delay shot.
//
int MyDocument::DumpPseudo(u32 addr) {
  if (!instrManager_.InstrIsExists(addr)) return -1;
  MyInstruction *instr = instrManager_.FetchInstruction(addr);
  auto it = mapTables.find(instr->mnemonic_);
  if (it == mapTables.end()) {
    std::cout << "\nWARNING\tDumpPseudo\tUnimplemented mnemonic: " << instr->mnemonic_ << std::endl;
    std::cout << instr->AsString(true) << std::endl;
    return -1;
  }
  return (this->*it->second.funcPseudo)(instr);
}

int MyDocument::PseuDoJump(MyInstruction *instr)
{
  int jal_ra = instr->addr_ + 4;
  if (instr->info_.hasDelaySlot) {
		jal_ra += 4;
  }
  auto &arg0 = instr->arguments_[0];
  MyInstruction *next_instr = nullptr;
  std::string op = "";
  std::string op_z = "";
  std::string op_l = "";

  if (instr->info_.hasDelaySlot) {
    next_instr = instrManager_.GetInstruction(instr->addr_ + 4);
    if (!next_instr) {
      std::cout << "WARNING\tPseuDoJump\tDelay shot instruction out of range" << std::endl;
      return -1;
    }
  }

  if (instr->mnemonic_ == "beq") op = "==";
  else if (instr->mnemonic_ == "bne") op = "!=";

  else if (instr->mnemonic_ == "blez") op_z = "<= 0";
  else if (instr->mnemonic_ == "bgtz") op_z = "> 0";
  else if (instr->mnemonic_ == "bltz") op_z = "< 0";
  else if (instr->mnemonic_ == "bgez") op_z = ">= 0";

  else if (instr->mnemonic_ == "bnel") op_l = "!=";
  else if (instr->mnemonic_ == "beql") op_l = "==";

  if (instr->mnemonic_ == "jal") {
    if (next_instr) DumpPseudo(next_instr->addr_);

    if (arg0.IsZero()) return -1;

    std::cout << "\tv0 = " << arg0.Str() << "(...)";
		std::cout << "\t// ra = " << jal_ra << ";";
		std::cout << " goto -> " << arg0.value_ << ";";
		std::cout << std::endl;
  } else if (instr->mnemonic_ == "j") {
    if (next_instr) DumpPseudo(next_instr->addr_);

    std::cout << "\tgoto -> " << arg0.Str() << std::endl;
  } else if (instr->mnemonic_ == "jr") {
    if (next_instr) DumpPseudo(next_instr->addr_);

    if (arg0.type_ == ArgReg && arg0.reg_ == "ra") {
			std::cout << "\treturn v0";
			std::cout << "\t// goto -> ra;";
			std::cout << std::endl;
		} else {
			std::cout << "\tgoto -> " << arg0.Str() << std::endl;
		}
  } else if (!op.empty()) {
		if (next_instr) DumpPseudo(next_instr->addr_);

    std::cout << '\t' << "if (" << instr->arguments_[0].Str() << " " << op << " " << instr->arguments_[1].Str() << ") ";
    std::cout << "goto -> " << instr->arguments_[2].Str() << std::endl;

  } else if (!op_z.empty()) {
		if (next_instr) DumpPseudo(next_instr->addr_);

    std::cout << '\t' << "if ((s32)" << instr->arguments_[0].Str() << " " << op_z << ") ";
		std::cout << "goto -> " << instr->arguments_[1].Str() << std::endl;
  } else if (!op_l.empty()) {
    std::cout << '\t' << "if (" << instr->arguments_[0].Str() << " " << op_l << " " << instr->arguments_[1].Str() << ") {" << std::endl;
		if (next_instr) {
			std::cout << '\t';
			DumpPseudo(next_instr->addr_);
		}
		std::cout << "\t\t" << "goto -> " << instr->arguments_[2].Str() << std::endl;
		std::cout << '\t' << "}" << std::endl;
  } else {
    return -1;
  }
	return instr->info_.hasDelaySlot ? +1 : 0;
}

int MyDocument::PseudoSyscall(MyInstruction *instr) {
  MyHLEFunction *hlefun = GetFunc(instr->arguments_[0].Str());
  if (hlefun) {
    std::cout << '\t';
    if (hlefun->retmask.size() > 0) {
      std::cout << "v0 = (" << hlefun->retmask[0] << ")";
    }

    std::cout << instr->arguments_[0].Str() << "(";

    for (int arg_i = 0; arg_i < hlefun->argmask.size(); arg_i++) {
      if (arg_i > 0)
        std::cout << ", ";
      std::cout << hlefun->argmask[arg_i] << " "
                << currentDebugMIPS->GetRegName(0, 4 + arg_i);
    }

    for (int ret_i = 1; ret_i < hlefun->retmask.size(); ret_i++) {
      if (ret_i > 1)
        std::cout << ", ";
      std::cout << hlefun->retmask[ret_i]
                  << "& v" << ret_i;
    }

    std::cout << ")" << std::endl;
  } else {
		std::cout << "\t" << instr->arguments_[0].Str() << "(...)" << std::endl;
    return -1;
  }
  return 0;
}
