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
	"strings"
)

// -----------------------------------------------------------------------------
// Types System

type TypeKind int

const (
	TypeInt TypeKind = iota
	TypeLong
	TypeShort
	TypeDouble
	TypeFloat
	TypeChar
	TypeBool
	TypeByte
	TypeVoid
	TypeString
	TypeArray
)

type Type struct {
	Kind     TypeKind
	ElemType *Type
}

// Pre-defined basic types
var (
	TInt    = &Type{Kind: TypeInt}
	TLong   = &Type{Kind: TypeLong}
	TShort  = &Type{Kind: TypeShort}
	TDouble = &Type{Kind: TypeDouble}
	TFloat  = &Type{Kind: TypeFloat}
	TChar   = &Type{Kind: TypeChar}
	TBool   = &Type{Kind: TypeBool}
	TByte   = &Type{Kind: TypeByte}
	TVoid   = &Type{Kind: TypeVoid}
	TString = &Type{Kind: TypeString}
)

func (t *Type) IsInt() bool    { return t == TInt }
func (t *Type) IsLong() bool   { return t == TLong }
func (t *Type) IsShort() bool  { return t == TShort }
func (t *Type) IsDouble() bool { return t == TDouble }
func (t *Type) IsFloat() bool  { return t == TFloat }
func (t *Type) IsChar() bool   { return t == TChar }
func (t *Type) IsBool() bool   { return t == TBool }
func (t *Type) IsByte() bool   { return t == TByte }
func (t *Type) IsVoid() bool   { return t == TVoid }
func (t *Type) IsString() bool { return t == TString }
func (t *Type) IsArray() bool  { return t.Kind == TypeArray }

func (t *Type) String() string {
	switch t.Kind {
	case TypeInt:
		return "int"
	case TypeLong:
		return "long"
	case TypeShort:
		return "short"
	case TypeDouble:
		return "double"
	case TypeFloat:
		return "float"
	case TypeChar:
		return "char"
	case TypeBool:
		return "bool"
	case TypeByte:
		return "byte"
	case TypeVoid:
		return "void"
	case TypeString:
		return "string"
	case TypeArray:
		return fmt.Sprintf("[]%v", t.ElemType)
	default:
		utils.Unimplement()
	}
	return ""
}

// -----------------------------------------------------------------------------
// Types Inference
// It annotates the AST with type information by propagating types from the leaves
// and applying some special rules for certain nodes.

type Stack struct {
	items []interface{}
}

func NewStack() *Stack {
	return &Stack{}
}

func (s *Stack) Push(item interface{}) {
	s.items = append(s.items, item)
}

func (s *Stack) Top() interface{} {
	length := len(s.items)
	if length == 0 {
		return nil
	}
	return s.items[length-1]
}

func (s *Stack) Pop() interface{} {
	length := len(s.items)
	if length == 0 {
		panic("Stack is empty")
	}

	item, items := s.items[length-1], s.items[:length-1]
	s.items = items
	return item
}

type Infer struct {
	varScopes  *Stack           // mapping var name to its type
	funcScopes map[string]*Type // return type for all functions
}

func (infer *Infer) numScopes() int {
	return len(infer.varScopes.items)
}

func (infer *Infer) enterScope() map[string]*Type {
	names := make(map[string]*Type)
	infer.varScopes.Push(names)
	return names
}

func (infer *Infer) exitScope() {
	infer.varScopes.Pop()
}

func (infer *Infer) getLetType(name string) *Type {
	for i := len(infer.varScopes.items) - 1; i >= 0; i-- {
		if names, ok := infer.varScopes.items[i].(map[string]*Type); ok {
			if t, ok := names[name]; ok {
				return t
			}
		}
	}
	return nil
}

func (infer *Infer) setVarType(name string, t *Type) {
	// Find variable from the innermost scope to the outermost scope and
	// set its type.
	for i := len(infer.varScopes.items) - 1; i >= 0; i-- {
		if names, ok := infer.varScopes.items[i].(map[string]*Type); ok {
			if _, ok := names[name]; ok {
				names[name] = t

				return
			}
		}
	}
	// Otherwise, set the type in the current scope.
	names := infer.varScopes.Top().(map[string]*Type)
	names[name] = t
}

func (infer *Infer) inferPre(node AstNode, _ AstNode, depth int) interface{} {
	switch node.(type) {
	case *FuncDecl:
		names := infer.enterScope()
		funcDecl := node.(*FuncDecl)
		for _, param := range funcDecl.Params {
			names[param.(*VarExpr).Name] = param.GetType()
		}
	case *BlockDecl:
		infer.enterScope()

	case *ForStmt, *IfStmt, *WhileStmt:
		infer.enterScope()
	}
	return nil
}

func (infer *Infer) inferPost(node AstNode, _ AstNode, depth int) interface{} {
	switch node.(type) {
	case *FuncDecl:
		infer.exitScope()
	case *BlockDecl:
		infer.exitScope()

	case *ForStmt, *IfStmt, *WhileStmt:
		infer.exitScope()
	}
	return nil
}

// Arithmetic conversion rules
func (infer *Infer) resolveType(opt TokenKind, left, right interface{}) *Type {
	var lt, rt *Type
	if left.(*Type) != nil {
		lt = left.(*Type)
	}
	if right.(*Type) != nil {
		rt = right.(*Type)
	}

	// string
	if lt == TString || rt == TString {
		return TString
	}
	return rt
}

func (infer *Infer) infer(node AstNode, _ AstNode, depth int) interface{} {
	// Expression must return inferred type even if it's undetermined
	switch node := node.(type) {
	case *IntExpr, *LongExpr, *DoubleExpr,
		*FloatExpr, *CharExpr,
		*BoolExpr, *ByteExpr,
		*StrExpr, *ArrayExpr:
		if node.(AstExpr).GetType() == nil {
			syntaxError("literal must hold the type info")
		}
		return node.(AstExpr).GetType()
	case *IndexExpr:
		varType := infer.getLetType(node.Name)
		if varType == TString {
			// indexing string returns char
			node.SetType(TChar)
			return TChar
		} else {
			// indexing array returns the element type
			utils.Assert(varType.Kind == TypeArray, "indexing non-array type")
			elemType := varType.ElemType
			node.SetType(elemType)
			return elemType
		}
	case *ConditionalExpr:
		thenExpr := node.Then
		thenType := infer.infer(thenExpr, node, depth+1)
		node.SetType(thenType.(*Type))
		return thenType
	case *FuncCallExpr:
		retType := infer.funcScopes[node.Name]
		node.SetType(retType)
		return retType
	case *UnaryExpr:
		leftType := infer.infer(node.Left, node, depth+1)
		if leftType.(*Type) != nil {
			node.SetType(leftType.(*Type))
		}
		return leftType
	case *BinaryExpr:
		// cmp op and short-circuit op are special cases, they always produce
		// bool type.
		if node.Opt.IsCmpOp() || node.Opt.IsShortCircuitOp() {
			node.SetType(TBool)
			return TBool
		}
		leftType := infer.infer(node.Left, node, depth+1)
		rightType := infer.infer(node.Right, node, depth+1)
		finalType := infer.resolveType(node.Opt, leftType, rightType)
		if finalType != nil {
			node.SetType(finalType)
			return finalType
		}
		return rightType
	case *AssignExpr:
		leftType := infer.infer(node.Left, node, depth+1)
		rightType := infer.infer(node.Right, node, depth+1)
		finalType := infer.resolveType(node.Opt, leftType, rightType)
		// rightType := infer.infer(e.Right, e, depth+1)
		if finalType != nil {
			node.SetType(finalType)

			left := node.Left
			switch left.(type) {
			case *VarExpr:
				// let p = 3.14
				// type of p is inferred from the right side
				v := left.(*VarExpr)
				infer.setVarType(v.Name, finalType)
			case *IndexExpr:
				// TODO: Implement it
			default:
				utils.Unimplement()
			}
		}
		return rightType
	case *VarExpr:
		// Letiable can be redeclared in different scopes, so we need to track
		// the type of the variable other than fetching from type of node directly.
		if vt := node.GetType(); vt != nil {
			infer.setVarType(node.Name, vt)
			return vt
		}
		vt := infer.getLetType(node.Name)
		if vt != nil {
			// record it
			infer.setVarType(node.Name, vt)
			// also let var expr awares of its type
			node.SetType(vt)
		}
		return vt

	case *LetStmt:
		name := infer.infer(node.Var, node, depth+1)
		if name.(*Type) != nil {
			// type is explicitly declared, no need to infer
			varExpr := node.Var
			varName := varExpr.Name
			infer.setVarType(varName, name.(*Type))
		} else {
			// infer type from the right side
			right := infer.infer(node.Init, node, depth+1)
			if right != nil {
				varExpr := node.Var
				varName := varExpr.Name
				infer.setVarType(varName, right.(*Type))
			}
		}
	}
	return nil
}

// InferTypes performs type inference on the given AST declaration.
// It creates an instance of the Infer struct and uses it to walk the AST,
// inferring types for expressions and verifying the balance of scopes.
// After type inference, it checks if all expressions are typed.
func InferTypes(debug bool, roots ...*PackageDecl) {
	infer := &Infer{}

	// Register return type for all functions
	infer.funcScopes = make(map[string]*Type)
	for _, root := range roots {
		for _, funcDecl := range root.Func {
			funcDecl := funcDecl.(*FuncDecl)
			infer.funcScopes[funcDecl.Name] = funcDecl.RetType
			if funcDecl.Builtin {
				// manually register return type for built-in functions
				call := funcDecl.Block.(*BlockDecl).Stmts[0].(*SimpleStmt).Expr.(*FuncCallExpr)
				callName := call.Name
				infer.funcScopes[callName] = funcDecl.RetType
			}
		}
	}

	// Deduces expression type for given AST
	for _, root := range roots {
		infer.varScopes = NewStack()
		walker := NewAstWalker(root, infer.infer, infer.inferPre, infer.inferPost)
		walker.WalkAst(root, root, 0)
		if infer.numScopes() != 0 {
			syntaxError("scope is unbalanced after type inference")
		}

		if debug {
			fmt.Printf("== TypedAst(%s) ==\n", root.Source)
			PrintAst(root, true)
		}

		// Good, we have typed AST!
	}
}

// -----------------------------------------------------------------------------
// Type Checker
//
// It performs type checks on the AST, ensuring that all expressions are typed
// and obey the rules of the language for certain AST construction component.
type TypeChecker struct {
	root    *PackageDecl
	current *FuncDecl
	funcs   []*FuncDecl
}

func (tc *TypeChecker) checkBinaryExpr(node *BinaryExpr) {
	// Logical expressions must be boolean on both sides
	if node.Opt.IsLogicalOp() {
		leftType := node.Left.GetType()
		rightType := node.Right.GetType()
		if leftType.Kind != TypeBool || rightType.Kind != TypeBool {
			syntaxError(fmt.Sprintf("logical expression must be boolean: %v", node))
		}
	}
}

func (tc *TypeChecker) checkUnaryExpr(node *UnaryExpr) {
	// Logical expressions must be boolean
	if node.Opt.IsLogicalOp() {
		leftType := node.Left.GetType()
		if leftType.Kind != TypeBool {
			syntaxError(fmt.Sprintf("logical expression must be boolean: %v", node))
		}
	}
}

func (tc *TypeChecker) checkFuncCallExpr(node *FuncCallExpr) {
	// Type of arguments match the function signature
	var funcDecl *FuncDecl
	for _, f := range tc.funcs {
		if f.Name == node.Name {
			funcDecl = f
			break
		} else if strings.HasPrefix(node.Name, "rt_") {
			sname := strings.TrimPrefix(node.Name, "rt_")
			if f.Name == sname {
				funcDecl = f
				break
			}
		}
	}
	if funcDecl == nil {
		syntaxError(fmt.Sprintf("call to %v function not found", node))
	}
	if len(node.Args) != len(funcDecl.Params) {
		syntaxError(fmt.Sprintf("argument count mismatch: %v", node))
	}
	for i, arg := range node.Args {
		argType := arg.GetType()
		paramType := funcDecl.Params[i].GetType()
		if argType.Kind != paramType.Kind {
			syntaxError(fmt.Sprintf("argument type mismatch: %v", node))
		}
	}
}

func (tc *TypeChecker) checkReturnStmt(node *ReturnStmt) {
	// Return value matches the func return type
	retVal := node.Expr
	if retVal != nil {
		retValType := retVal.GetType()
		if retValType != tc.current.RetType {
			syntaxError(fmt.Sprintf("bad return type: %v for %v", node, tc.current.Name))
		}
	}
}

func (tc *TypeChecker) check(node AstNode, _ AstNode, depth int) interface{} {
	if _, yes := node.(*FuncDecl); yes {
		tc.current = node.(*FuncDecl)
	}
	// All expressions must be typed
	if e, yes := node.(AstExpr); yes {
		if e.GetType() == nil {
			PrintAst(tc.root, true)
			syntaxError(fmt.Sprintf("expression is not typed: %v", node))
		}
	}
	switch node := node.(type) {
	case *ReturnStmt:
		tc.checkReturnStmt(node)
	case *BinaryExpr:
		tc.checkBinaryExpr(node)
	case *UnaryExpr:
		tc.checkUnaryExpr(node)
	case *FuncCallExpr:
		tc.checkFuncCallExpr(node)
	}
	return nil
}

func TypeCheck(debug bool, roots ...*PackageDecl) {
	typeChecker := &TypeChecker{}
	// Register all functions
	typeChecker.funcs = make([]*FuncDecl, 0)
	for _, root := range roots {
		for _, funcDecl := range root.Func {
			funcDecl := funcDecl.(*FuncDecl)
			typeChecker.funcs = append(typeChecker.funcs, funcDecl)
		}
	}

	for _, root := range roots {
		typeChecker.root = root
		walker := NewAstWalker(root, typeChecker.check, nil, nil)
		walker.WalkAst(root, root, 0)
	}
}
