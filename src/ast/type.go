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

var BasicTypes = map[TypeKind]*Type{
	TypeInt:    {Kind: TypeInt},
	TypeLong:   {Kind: TypeLong},
	TypeShort:  {Kind: TypeShort},
	TypeDouble: {Kind: TypeDouble},
	TypeFloat:  {Kind: TypeFloat},
	TypeChar:   {Kind: TypeChar},
	TypeBool:   {Kind: TypeBool},
	TypeByte:   {Kind: TypeByte},
	TypeVoid:   {Kind: TypeVoid},
	TypeString: {Kind: TypeString},
}

var (
	TInt    = BasicTypes[TypeInt]
	TLong   = BasicTypes[TypeLong]
	TShort  = BasicTypes[TypeShort]
	TDouble = BasicTypes[TypeDouble]
	TFloat  = BasicTypes[TypeFloat]
	TChar   = BasicTypes[TypeChar]
	TBool   = BasicTypes[TypeBool]
	TByte   = BasicTypes[TypeByte]
	TVoid   = BasicTypes[TypeVoid]
	TString = BasicTypes[TypeString]
)

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

//-----------------------------------------------------------------------------
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

	if lt == BasicTypes[TypeDouble] || rt == BasicTypes[TypeDouble] {
		return BasicTypes[TypeDouble]
	}
	if lt == BasicTypes[TypeLong] || rt == BasicTypes[TypeLong] {
		return BasicTypes[TypeLong]
	}
	return rt
}

func (infer *Infer) infer(node AstNode, _ AstNode, depth int) interface{} {
	// Expression must return inferred type even if it's undetermined
	switch node.(type) {
	case *IntExpr, *LongExpr, *DoubleExpr,
		*FloatExpr, *CharExpr,
		*BoolExpr, *ByteExpr,
		*StrExpr, *ArrayExpr:
		if node.(AstExpr).GetType() == nil {
			syntaxError("literal must hold the type info")
		}
		return node.(AstExpr).GetType()
	case *IndexExpr:
		e := node.(*IndexExpr)
		arrType := infer.getLetType(e.Name)
		elemType := arrType.ElemType
		e.SetType(elemType)
		return elemType
	case *TernaryExpr:
		e := node.(*TernaryExpr)
		thenExpr := e.Then
		thenType := infer.infer(thenExpr, e, depth+1)
		e.SetType(thenType.(*Type))
		return thenType
	case *FuncCallExpr:
		e := node.(*FuncCallExpr)
		retType := infer.funcScopes[e.Name]
		e.SetType(retType)
		return retType
	case *UnaryExpr:
		e := node.(*UnaryExpr)
		leftType := infer.infer(e.Left, e, depth+1)
		if leftType.(*Type) != nil {
			e.SetType(leftType.(*Type))
		}
		return leftType
	case *BinaryExpr:
		e := node.(*BinaryExpr)
		// cmp op and short-circuit op are special cases, they always produce
		// bool type.
		if e.Opt.IsCmpOp() || e.Opt.IsShortCircuitOp() {
			e.SetType(BasicTypes[TypeBool])
			return BasicTypes[TypeBool]
		}
		leftType := infer.infer(e.Left, e, depth+1)
		rightType := infer.infer(e.Right, e, depth+1)
		finalType := infer.resolveType(e.Opt, leftType, rightType)
		if finalType != nil {
			e.SetType(finalType)
		}
		return rightType
	case *AssignExpr:
		e := node.(*AssignExpr)
		leftType := infer.infer(e.Left, e, depth+1)
		rightType := infer.infer(e.Right, e, depth+1)
		finalType := infer.resolveType(e.Opt, leftType, rightType)
		// rightType := infer.infer(e.Right, e, depth+1)
		if finalType != nil {
			e.SetType(finalType)

			left := e.Left
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
		v := node.(*VarExpr)
		if vt := v.GetType(); vt != nil {
			infer.setVarType(v.Name, vt)
			return vt
		}
		vt := infer.getLetType(v.Name)
		if vt != nil {
			// record it
			infer.setVarType(v.Name, vt)
			// also let var expr awares of its type
			v.SetType(vt)
		}
		return vt

	case *LetStmt:
		s := node.(*LetStmt)
		name := infer.infer(s.Var, s, depth+1)
		if name.(*Type) != nil {
			// type is explicitly declared, no need to infer
			varExpr := s.Var
			varName := varExpr.Name
			infer.setVarType(varName, name.(*Type))
		} else {
			// infer type from the right side
			right := infer.infer(s.Init, s, depth+1)
			if right != nil {
				varExpr := s.Var
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
func InferTypes(debug bool, roots ...*RootDecl) {
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
type TypeChecker struct {
	root    *RootDecl
	current *FuncDecl
	funcs   []*FuncDecl
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
	// Return value matches the func return type
	if s, yes := node.(*ReturnStmt); yes {
		retVal := s.Expr
		if retVal != nil {
			retValType := retVal.GetType()
			if retValType != tc.current.RetType {
				syntaxError(fmt.Sprintf("bad return type: %v for %v", s, tc.current.Name))
			}
		}
	}
	// Type of arguments match the function signature
	if s, yes := node.(*FuncCallExpr); yes {
		var funcDecl *FuncDecl
		for _, f := range tc.funcs {
			if f.Name == s.Name {
				funcDecl = f
				break
			} else if strings.HasPrefix(s.Name, "rt_") {
				sname := strings.TrimPrefix(s.Name, "rt_")
				if f.Name == sname {
					funcDecl = f
					break
				}
			}
		}
		if funcDecl == nil {
			syntaxError(fmt.Sprintf("call to %v function not found", s))
		}
		if len(s.Args) != len(funcDecl.Params) {
			syntaxError(fmt.Sprintf("argument count mismatch: %v", s))
		}
		for i, arg := range s.Args {
			argType := arg.GetType()
			paramType := funcDecl.Params[i].GetType()
			if argType.Kind != paramType.Kind {
				syntaxError(fmt.Sprintf("argument type mismatch: %v", s))
			}
		}

	}
	// Logical expressions must be boolean on both sides
	if s, yes := node.(*BinaryExpr); yes {
		if s.Opt.IsLogicalOp() {
			leftType := s.Left.GetType()
			rightType := s.Right.GetType()
			if leftType.Kind != TypeBool || rightType.Kind != TypeBool {
				syntaxError(fmt.Sprintf("logical expression must be boolean: %v", s))
			}
		}

	}
	return nil
}

func TypeCheck(debug bool, roots ...*RootDecl) {
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
