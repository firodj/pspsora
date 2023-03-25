package codegen

import "fmt"

type NodeType string

const (
	TypeASTBinary    NodeType = "binary"
	TypeASTUnary     NodeType = "unary"
	TypeASTAssign    NodeType = "assign"
	TypeASTNumber    NodeType = "number"
	TypeASTSymbolRef NodeType = "symbol_ref"
	TypeASTPointer   NodeType = "pointer"
)

type ASTNode interface {
	Type() NodeType
}

//

type ASTBinary struct {
	ASTNode
	Left  ASTNode
	Right ASTNode
	Op    string
}

func (a *ASTBinary) Type() NodeType {
	return TypeASTBinary
}

func (a *ASTBinary) String() string {
	return fmt.Sprintf("%s %s %s", a.Left, a.Op, a.Right)
}

//

type ASTUnary struct {
	ASTNode
	Op   string
	Expr ASTNode
}

func (a *ASTUnary) Type() NodeType {
	return TypeASTUnary
}

func (a *ASTUnary) String() string {
	return fmt.Sprintf("%s(%s)", a.Op, a.Expr)
}

//

type ASTAssign struct {
	ASTBinary
}

func (a *ASTAssign) Type() NodeType {
	return TypeASTAssign
}

func (a *ASTAssign) String() string {
	return fmt.Sprintf("%s = %s", a.Left, a.Right)
}

//

type ASTNumber struct {
	ASTNode
	Value int
}

func (a *ASTNumber) Type() NodeType {
	return TypeASTNumber
}

func (a *ASTNumber) String() string {
	return fmt.Sprintf("%#x", a.Value)
}

//

type ASTPointer struct {
	ASTNode
	Sz   string
	Expr ASTNode
}

func (a *ASTPointer) Type() NodeType {
	return TypeASTPointer
}

func (a *ASTPointer) String() string {
	return fmt.Sprintf("*(%s*)&mem[%s]", a.Sz, a.Expr)
}

//

type ASTSymbol struct {
	ASTNode
	Name string
}

type ASTSymbolRef struct {
	ASTSymbol
}

func (a *ASTSymbolRef) Type() NodeType {
	return TypeASTSymbolRef
}

func (a *ASTSymbolRef) String() string {
	return a.Name
}

//

type ASTCall struct {
	ASTNode
	Expr ASTNode
	Args []ASTNode
}

func (a *ASTCall) Type() NodeType {
	return TypeASTSymbolRef
}

func (a *ASTCall) String() string {
	s := fmt.Sprintf("%s(", a.Expr)
	for i := range a.Args {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%s", a.Args[i])
	}
	s += ")"
	return s
}
