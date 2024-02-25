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
package test

import (
	"falcon/ast"
	"testing"
)

func MustBeType(t *testing.T, e ast.AstExpr, ty *ast.Type) {
	if e.GetType() != ty {
		t.Fatalf("Expect type %v, got %v", ty, e.GetType())
	}
}

func TestBasicType(t *testing.T) {
	root := ast.ParseText("0")
	ast.InferTypes(false, root)
	MustBeType(t, checkSimpleStmt(root).Expr, ast.BasicTypes[ast.TypeInt])

	root = ast.ParseText("true")
	ast.InferTypes(false, root)
	MustBeType(t, checkSimpleStmt(root).Expr, ast.BasicTypes[ast.TypeBool])

	root = ast.ParseText("false")
	ast.InferTypes(false, root)
	MustBeType(t, checkSimpleStmt(root).Expr, ast.BasicTypes[ast.TypeBool])

	root = ast.ParseText("3.14")
	ast.InferTypes(false, root)
	MustBeType(t, checkSimpleStmt(root).Expr, ast.BasicTypes[ast.TypeDouble])

	root = ast.ParseText("'\t'")
	ast.InferTypes(false, root)
	MustBeType(t, checkSimpleStmt(root).Expr, ast.BasicTypes[ast.TypeChar])

	root = ast.ParseText(`"this is string"`)
	ast.InferTypes(false, root)
	MustBeType(t, checkSimpleStmt(root).Expr, ast.BasicTypes[ast.TypeString])

	ast.PrintAst(root, true)
}

func TestInferBinaryExpr(t *testing.T) {
	root := ast.ParseText("3+2-2")
	ast.InferTypes(false, root)
	a := checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeInt])

	root = ast.ParseText("-3+2-5")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeInt])

	root = ast.ParseText("3+(2-2)")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeInt])

	root = ast.ParseText("3+2*2")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeInt])

	root = ast.ParseText("3/2*2")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeInt])

	root = ast.ParseText("3/2.0*2")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeDouble])

	root = ast.ParseText("3/2.0*2.0")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeDouble])

	root = ast.ParseText("3/2.0*2.0+3")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeDouble])

	root = ast.ParseText("3/2.0*2.0+3.0")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeDouble])

	root = ast.ParseText("3/2.0*2.0+3.0-2")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeDouble])

	root = ast.ParseText("3/2.0*2.0+3.0-2.0")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeDouble])
}

func TestInferUnaryExpr(t *testing.T) {
	root := ast.ParseText("!true")
	ast.InferTypes(false, root)
	a := checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeBool])

	root = ast.ParseText("!3")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeInt])

	root = ast.ParseText("!false")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeBool])

	root = ast.ParseText("-3")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeInt])

	root = ast.ParseText("-3.0")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeDouble])

	root = ast.ParseText("-0")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeInt])

	root = ast.ParseText("-3.0")
	ast.InferTypes(false, root)
	a = checkSimpleStmt(root)
	MustBeType(t, a.Expr, ast.BasicTypes[ast.TypeDouble])
}
