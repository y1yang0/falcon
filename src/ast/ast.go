// Copyright (c) 2024 The Falcon Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
package ast

import (
	"falcon/utils"
	"fmt"
	"os"
	"reflect"
)

// -----------------------------------------------------------------------------
// Ast Root Interfaces

type AstNode interface {
	String() string
}

type AstExpr interface {
	AstNode
	GetType() *Type
	SetType(*Type)
}

type AstStmt interface {
	AstNode
}

type AstDecl interface {
	AstNode
}

// -----------------------------------------------------------------------------
// Expressions

type Expr struct {
	Type *Type
}

type UnaryExpr struct {
	Expr
	Left AstExpr
	Opt  TokenKind
}

type BinaryExpr struct {
	Expr
	Left  AstExpr
	Right AstExpr
	Opt   TokenKind
}

type IndexExpr struct {
	Expr
	Index AstExpr
	Name  string
}

type VarExpr struct {
	Expr
	Name string
}

type IntExpr struct {
	Expr
	Value int
}

type LongExpr struct {
	Expr
	Value int64
}

type ShortExpr struct {
	Expr
	Value int16
}

type DoubleExpr struct {
	Expr
	Value float64
}

type FloatExpr struct {
	Expr
	Value float32
}

type CharExpr struct {
	Expr
	Value int32
}

type BoolExpr struct {
	Expr
	Value bool
}

type ByteExpr struct {
	Expr
	Value byte
}

type VoidExpr struct {
	Expr
}

type NullExpr struct {
	Expr
}

type StrExpr struct {
	Expr
	Value string
}

type ArrayExpr struct {
	Expr
	Elems []AstExpr
}

type AssignExpr struct {
	Expr
	Left  AstExpr
	Right AstExpr
	Opt   TokenKind
}
type TernaryExpr struct {
	Expr
	Cond AstExpr
	Then AstExpr
	Else AstExpr
}
type FuncCallExpr struct {
	Expr
	Name string
	Args []AstExpr
}

func (e *Expr) String() string {
	return fmt.Sprintf("Expr{%v}", e.Type)
}

func (e *Expr) GetType() *Type {
	return e.Type
}

func (e *Expr) SetType(t *Type) {
	e.Type = t
}

func (a *ArrayExpr) String() string {
	return fmt.Sprintf("ArrayExpr")
}

func (a *AssignExpr) String() string {
	return fmt.Sprintf("AssignExpr{%v}", a.Opt.String())
}

func (n *NullExpr) String() string {
	return fmt.Sprintf("NullExpr")
}
func (f *FuncCallExpr) String() string {
	return fmt.Sprintf("FuncCallExpr{%v}", f.Name)
}

func (t *UnaryExpr) String() string {
	return fmt.Sprintf("UnaryExpr{%v}", t.Opt.String())
}

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("BinaryExpr{%v}", b.Opt.String())
}

func (i *IndexExpr) String() string {
	return fmt.Sprintf("IndexExpr{%v}", i.Name)
}

func (v *VarExpr) String() string {
	return fmt.Sprintf("VarExpr{%v}", v.Name)
}

func (t *TernaryExpr) String() string {
	return fmt.Sprintf("TernaryExpr")
}

func (i *IntExpr) String() string {
	return fmt.Sprintf("IntExpr{%v}", i.Value)
}

func (l *LongExpr) String() string {
	return fmt.Sprintf("LongExpr{%v}", l.Value)
}

func (s *ShortExpr) String() string {
	return fmt.Sprintf("ShortExpr{%v}", s.Value)
}

func (d *DoubleExpr) String() string {
	return fmt.Sprintf("DoubleExpr{%v}", d.Value)
}

func (f *FloatExpr) String() string {
	return fmt.Sprintf("FloatExpr{%v}", f.Value)
}

func (c *CharExpr) String() string {
	return fmt.Sprintf("CharExpr{%v}", c.Value)
}

func (b *BoolExpr) String() string {
	return fmt.Sprintf("BoolExpr{%v}", b.Value)
}

func (b *ByteExpr) String() string {
	return fmt.Sprintf("ByteExpr{%v}", b.Value)
}

func (v *VoidExpr) String() string {
	return fmt.Sprintf("VoidExpr")
}

func (s *StrExpr) String() string {
	return fmt.Sprintf("StrExpr{%v}", s.Value)
}

// -----------------------------------------------------------------------------
// Statements

type AssignStmt struct {
	Left  AstExpr
	Right AstExpr
}

type SimpleStmt struct {
	Expr AstExpr
}

type ReturnStmt struct {
	Expr AstExpr
}

type LetStmt struct {
	Var  *VarExpr
	Init AstExpr
}

type IfStmt struct {
	Cond AstExpr
	Then AstDecl
	Else AstDecl
}

type ForStmt struct {
	Init AstStmt
	Cond AstExpr
	Post AstExpr
	Body AstDecl
}

type WhileStmt struct {
	Cond AstExpr
	Body AstDecl
}

type BreakStmt struct {
}
type ContinueStmt struct {
}

func (f *ForStmt) String() string {
	return fmt.Sprintf("ForStmt")
}

func (w *WhileStmt) String() string {
	return fmt.Sprintf("WhileStmt")
}

func (b *BreakStmt) String() string {
	return fmt.Sprintf("BreakStmt")
}

func (c *ContinueStmt) String() string {
	return fmt.Sprintf("ContinueStmt")
}

func (s *SimpleStmt) String() string {
	return fmt.Sprintf("SimpleStmt")
}

func (s *AssignStmt) String() string {
	return fmt.Sprintf("AssignStmt")
}

func (s *ReturnStmt) String() string {
	return fmt.Sprintf("ReturnStmt")
}

func (s *LetStmt) String() string {
	return fmt.Sprintf("LetStmt")
}

func (s *IfStmt) String() string {
	return fmt.Sprintf("IfStmt")
}

// -----------------------------------------------------------------------------
// Declarations

type BlockDecl struct {
	AstDecl
	Name  string //optional name
	Stmts []AstStmt
}

type FuncDecl struct {
	AstDecl
	Name    string
	Params  []AstExpr
	Block   AstDecl
	RetType *Type
	Builtin bool
}

type RootDecl struct {
	AstDecl
	Source string
	Func   []AstDecl
	List   []AstNode
}

func (r *RootDecl) String() string {
	return fmt.Sprintf("RootDecl")
}

func (b *BlockDecl) String() string {
	return fmt.Sprintf("BlockDecl{%s}", b.Name)
}

func (f *FuncDecl) String() string {
	if f.Builtin {
		return fmt.Sprintf("FuncDecl{%v@builtin}", f.Name)
	} else {
		return fmt.Sprintf("FuncDecl{%v}", f.Name)
	}
}

// -----------------------------------------------------------------------------
// Lexical tokens

type TokenKind int

const (
	INVALID    TokenKind = iota // <invalid>
	TK_IDENT                    // <identifier>
	TK_EOF                      // <eof>
	LIT_INT                     // <integer>
	LIT_LONG                    // <long>
	LIT_SHORT                   // <short>
	LIT_DOUBLE                  // <decimal>
	LIT_FLOAT                   // <float>
	LIT_CHAR                    // <character>
	LIT_BOOL                    // <bool>
	LIT_BYTE                    // <byte>
	LIT_STR                     // <string>

	TK_BITAND // &
	TK_BITOR  // |
	TK_BITNOT // ~
	TK_BITXOR // ^
	TK_LOGAND // &&
	TK_LOGOR  // ||
	TK_LOGNOT // !
	TK_PLUS   // +
	TK_MINUS  // -
	TK_TIMES  // *
	TK_DIV    // /
	TK_MOD    // %
	TK_EQ     // ==
	TK_NE     // !=
	TK_GT     // >
	TK_GE     // >=
	TK_LT     // <
	TK_LE     // <=
	TK_RSHIFT // >>
	TK_LSHIFT // <<

	TK_ASSIGN     // =
	TK_PLUS_AGN   // +=
	TK_MINUS_AGN  // -=
	TK_TIMES_AGN  // *=
	TK_DIV_AGN    // /=
	TK_MOD_AGN    // %=
	TK_RSHIFT_AGN // >>=
	TK_LSHIFT_AGN // <<=
	TK_BITXOR_AGN // ^=
	TK_BITAND_AGN // &=
	TK_BITOR_AGN  // |=
	TK_MATCH      // =>
	TK_COMMA      //
	TK_LPAREN     // (
	TK_RPAREN     // )
	TK_LBRACE     // {
	TK_RBRACE     // }
	TK_LBRACKET   // [
	TK_RBRACKET   // ]
	TK_SEMICOLON  // ;
	TK_COLON      // :
	TK_DOT        // .
	TK_QUESTION   // ?

	KW_IF       // if
	KW_ELSE     // else
	KW_TRUE     // true
	KW_FALSE    // false
	KW_WHILE    // while
	KW_FOR      // for
	KW_NULL     // null
	KW_FUNC     // func
	KW_RETURN   // return
	KW_BREAK    // break
	KW_CONTINUE // continue
	KW_LET      // let

	KW_TYPE_INT    // int
	KW_TYPE_LONG   // long
	KW_TYPE_SHORT  // short
	KW_TYPE_DOUBLE // double
	KW_TYPE_FLOAT  // float
	KW_TYPE_CHAR   // char
	KW_TYPE_BOOL   // bool
	KW_TYPE_BYTE   // byte
	KW_TYPE_VOID   // void
	KW_TYPE_STR    // string
)

func (t TokenKind) IsCmpOp() bool {
	switch t {
	case TK_EQ, TK_NE, TK_GT, TK_GE, TK_LT, TK_LE:
		return true
	}
	return false
}

func (t TokenKind) IsShortCircuitOp() bool {
	switch t {
	case TK_LOGAND, TK_LOGOR, TK_LOGNOT:
		return true
	}
	return false
}

func (t TokenKind) String() string {
	switch t {
	case INVALID:
		return "<invalid>"
	case TK_IDENT:
		return "<identifier>"
	case TK_EOF:
		return "<eof>"
	case LIT_INT:
		return "<integer>"
	case LIT_STR:
		return "<string>"
	case LIT_DOUBLE:
		return "<decimal>"
	case LIT_CHAR:
		return "<character>"

	case TK_BITAND:
		return "&"
	case TK_BITOR:
		return "|"
	case TK_BITNOT:
		return "~"
	case TK_RSHIFT:
		return ">>"
	case TK_LSHIFT:
		return "<<"
	case TK_BITXOR:
		return "^"
	case TK_LOGAND:
		return "&&"
	case TK_LOGOR:
		return "||"
	case TK_LOGNOT:
		return "!"
	case TK_PLUS:
		return "+"
	case TK_MINUS:
		return "-"
	case TK_TIMES:
		return "*"
	case TK_DIV:
		return "/"
	case TK_MOD:
		return "%"
	case TK_EQ:
		return "=="
	case TK_NE:
		return "!="
	case TK_GT:
		return ">"
	case TK_GE:
		return ">="
	case TK_LT:
		return "<"
	case TK_LE:
		return "<="
	case TK_ASSIGN:
		return "="
	case TK_PLUS_AGN:
		return "+="
	case TK_MINUS_AGN:
		return "-="
	case TK_TIMES_AGN:
		return "*="
	case TK_DIV_AGN:
		return "/="
	case TK_MOD_AGN:
		return "%="
	case TK_RSHIFT_AGN:
		return ">>="
	case TK_LSHIFT_AGN:
		return "<<="
	case TK_BITXOR_AGN:
		return "^="
	case TK_BITAND_AGN:
		return "&="
	case TK_BITOR_AGN:
		return "|="
	case TK_MATCH:
		return "=>"
	case TK_COMMA:
		return ","
	case TK_LPAREN:
		return "("
	case TK_RPAREN:
		return ")"
	case TK_LBRACE:
		return "{"
	case TK_RBRACE:
		return "}"
	case TK_LBRACKET:
		return "["
	case TK_RBRACKET:
		return "]"
	case TK_SEMICOLON:
		return ";"
	case TK_COLON:
		return ":"
	case TK_DOT:
		return "."
	case TK_QUESTION:
		return "?"

	case KW_IF:
		return "if"
	case KW_ELSE:
		return "else"
	case KW_TRUE:
		return "true"
	case KW_FALSE:
		return "false"
	case KW_WHILE:
		return "while"
	case KW_FOR:
		return "for"
	case KW_NULL:
		return "null"
	case KW_FUNC:
		return "func"
	case KW_RETURN:
		return "return"
	case KW_BREAK:
		return "break"
	case KW_CONTINUE:
		return "continue"
	case KW_LET:
		return "let"

	case KW_TYPE_INT:
		return "int"
	case KW_TYPE_LONG:
		return "long"
	case KW_TYPE_SHORT:
		return "short"
	case KW_TYPE_DOUBLE:
		return "double"
	case KW_TYPE_FLOAT:
		return "float"
	case KW_TYPE_CHAR:
		return "char"
	case KW_TYPE_BOOL:
		return "bool"
	case KW_TYPE_BYTE:
		return "byte"
	case KW_TYPE_VOID:
		return "void"
	case KW_TYPE_STR:
		return "string"
	default:
		utils.Unimplement()
	}
	return ""
}

var Keywords = map[string]TokenKind{
	"func":     KW_FUNC,
	"if":       KW_IF,
	"else":     KW_ELSE,
	"true":     KW_TRUE,
	"false":    KW_FALSE,
	"while":    KW_WHILE,
	"for":      KW_FOR,
	"null":     KW_NULL,
	"return":   KW_RETURN,
	"break":    KW_BREAK,
	"continue": KW_CONTINUE,
	"let":      KW_LET,

	"int":    KW_TYPE_INT,
	"long":   KW_TYPE_LONG,
	"short":  KW_TYPE_SHORT,
	"double": KW_TYPE_DOUBLE,
	"float":  KW_TYPE_FLOAT,
	"char":   KW_TYPE_CHAR,
	"bool":   KW_TYPE_BOOL,
	"byte":   KW_TYPE_BYTE,
	"void":   KW_TYPE_VOID,
	"string": KW_TYPE_STR,
}

// -----------------------------------------------------------------------------
// Utils for ast manipulation

type AstWalker struct {
	// the root node to start walking
	Root AstNode
	// apply before visiting a node
	FuncPre func(AstNode, AstNode, int) interface{}
	// apply when visiting a node
	Func func(AstNode, AstNode, int) interface{}
	// apply when leaving a node
	FuncPost func(AstNode, AstNode, int) interface{}
}

func NewAstWalker(root AstNode,
	funcs ...func(AstNode, AstNode, int) interface{}) *AstWalker {
	if len(funcs) == 0 {
		panic("Must provide at least one function")
	}
	switch len(funcs) {
	case 1:
		return &AstWalker{
			Root: root,
			Func: funcs[0],
		}
	case 2:
		return &AstWalker{
			Root:     root,
			Func:     funcs[0],
			FuncPost: funcs[1],
		}
	case 3:
		return &AstWalker{
			Root:     root,
			Func:     funcs[0],
			FuncPre:  funcs[1],
			FuncPost: funcs[2],
		}
	}

	return nil
}

// WalkAst Walk walks the AST in depth-first order, calling f for each node.
func (walker *AstWalker) WalkAst(node AstNode, prev AstNode, depth int) {
	if node == nil {
		return
	}
	if walker.FuncPre != nil {
		walker.FuncPre(node, prev, depth)
	}
	walker.Func(node, prev, depth)
	switch v := node.(type) {
	case *BreakStmt, *ContinueStmt:
		// Donothing
	case *SimpleStmt:
		walker.WalkAst(v.Expr, v, depth+1)
	case *AssignStmt:
		walker.WalkAst(v.Left, v, depth+1)
		walker.WalkAst(v.Right, v, depth+1)
	case *ReturnStmt:
		walker.WalkAst(v.Expr, v, depth+1)
	case *LetStmt:
		walker.WalkAst(v.Var, v, depth+1)
		walker.WalkAst(v.Init, v, depth+1)
	case *IfStmt:
		walker.WalkAst(v.Cond, v, depth+1)
		walker.WalkAst(v.Then, v, depth+1)
		walker.WalkAst(v.Else, v, depth+1)
	case *ForStmt:
		walker.WalkAst(v.Init, v, depth+1)
		walker.WalkAst(v.Cond, v, depth+1)
		walker.WalkAst(v.Post, v, depth+1)
		walker.WalkAst(v.Body, v, depth+1)
	case *WhileStmt:
		walker.WalkAst(v.Cond, v, depth+1)
		walker.WalkAst(v.Body, v, depth+1)

	case *UnaryExpr:
		walker.WalkAst(v.Left, v, depth+1)
	case *BinaryExpr:
		walker.WalkAst(v.Left, v, depth+1)
		walker.WalkAst(v.Right, v, depth+1)
	case *TernaryExpr:
		walker.WalkAst(v.Cond, v, depth+1)
		walker.WalkAst(v.Then, v, depth+1)
		walker.WalkAst(v.Else, v, depth+1)
	case *AssignExpr:
		walker.WalkAst(v.Left, v, depth+1)
		walker.WalkAst(v.Right, v, depth+1)
	case *IntExpr, *LongExpr, *ShortExpr, *DoubleExpr, *FloatExpr,
		*CharExpr, *BoolExpr, *ByteExpr, *VoidExpr, *NullExpr, *StrExpr:
	case *ArrayExpr:
		for _, elem := range v.Elems {
			walker.WalkAst(elem, v, depth+1)
		}
	case *VarExpr:
	case *FuncCallExpr:
		for _, elem := range v.Args {
			walker.WalkAst(elem, v, depth+1)
		}
	case *IndexExpr:
		walker.WalkAst(v.Index, v, depth+1)

	case *RootDecl:
		for _, elem := range v.List {
			walker.WalkAst(elem, v, depth+1)
		}
		for _, elem := range v.Func {
			walker.WalkAst(elem, v, depth+1)
		}
	case *FuncDecl:
		for _, elem := range v.Params {
			walker.WalkAst(elem, v, depth+1)
		}
		walker.WalkAst(v.Block, v, depth+1)
	case *BlockDecl:
		for _, elem := range v.Stmts {
			walker.WalkAst(elem, v, depth+1)
		}
	default:
		utils.Unimplement()
	}
	if walker.FuncPost != nil {
		walker.FuncPost(node, prev, depth)
	}
}

func PrintAst(root *RootDecl, showTypes bool) {
	printer := func(node AstNode, _ AstNode, ident int) interface{} {
		if node == nil {
			return nil
		}
		for i := 0; i < ident; i++ {
			print("..")
		}
		str := node.String()
		if showTypes {
			if expr, ok := node.(AstExpr); ok {
				str += fmt.Sprintf(" :: %v", expr.GetType())
			}
		}
		println(str)
		return nil
	}
	walker := NewAstWalker(root, printer)
	walker.WalkAst(root, root, 0)
}

func DumpAstToDotFile(name string, root *RootDecl) {
	f, err := os.Create(fmt.Sprintf("ast_%s.dot", name))
	if err != nil {
		panic(err)
	}
	defer func() {
		utils.ExecuteCmd(".",
			"dot",
			"-Tpng",
			fmt.Sprintf("ast_%s.dot", name),
			"-o",
			fmt.Sprintf("ast_%s.png", name),
		)
		f.Close()
		os.Remove(fmt.Sprintf("ast_%s.dot", name))
	}()
	f.WriteString("digraph G {\n")
	f.WriteString("  graph [ dpi = 500 ];\n")
	astWriter := func(node AstNode, prev AstNode, depth int) interface{} {
		edge := fmt.Sprintf("  %v_%p -> %v_%p\n",
			reflect.TypeOf(prev).Elem().Name(),
			prev,
			reflect.TypeOf(node).Elem().Name(),
			node,
		)
		n := fmt.Sprintf("  %v_%p [label=\"%v\"]\n",
			reflect.TypeOf(node).Elem().Name(),
			node,
			node.String(),
		)
		println("===" + edge)
		f.WriteString(edge)
		f.WriteString(n + "\n")
		return nil
	}
	walker := NewAstWalker(root, astWriter)
	walker.WalkAst(root, root, 0)
	f.WriteString("}\n")
}
