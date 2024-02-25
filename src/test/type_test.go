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

func TestLong(t *testing.T) {
	source := `
	func main(){
		let n1 long
		let n2 long
		assert_long(n1, n2)
		let a long = 55L
		let b long = 10L
		let p = a + b
		let q = a - b
		let r = a * b
		let s = a / b
		let t = a % b
		let u = a & b
		let v = a | b
		let w = a ^ b
		let x = a << b
		let y = a >> b
		let aa = a == b
		let ab = a != b
		let ac = a < b
		let ad = a > b
		let ae = a <= b
		let af = a >= b
		assert_long(p, 65L)
		assert_long(q, 45L)
		assert_long(r, 550L)
		assert_long(s, 5L)
		assert_long(t, 5L)
		assert_long(u, 2L)
		assert_long(v, 63L)
		assert_long(w, 61L)
		assert_long(x, 56320L)
		assert_long(y, 0L)
		assert_bool(aa, false)
		assert_bool(ab, true)
		assert_bool(ac, false)
		assert_bool(ad, true)
		assert_bool(ae, false)
		assert_bool(af, true)
	}
	`
	ExecExpect(source)
}

func TestShort(t *testing.T) {
	source := `
	func main(){
		let n1 short
		let n2 short
		assert_short(n1, n2)
		let a short = 55S
		let b short = 10S
		let p = a + b
		let q = a - b
		let r = a * b
		let s = a / b
		let t = a % b
		let u = a & b
		let v = a | b
		let w = a ^ b
		let x = a << b
		let y = a >> b
		let aa = a == b
		let ab = a != b
		let ac = a < b
		let ad = a > b
		let ae = a <= b
		let af = a >= b
		assert_short(p, 65S)
		assert_short(q, 45S)
		assert_short(r, 550S)
		assert_short(s, 5S)
		assert_short(t, 5S)
		assert_short(u, 2S)
		assert_short(v, 63S)
		assert_short(w, 61S)
		//assert_short(x, 56320S)
		assert_short(y, 0S)
		assert_bool(aa, false)
		assert_bool(ab, true)
		assert_bool(ac, false)
		assert_bool(ad, true)
		assert_bool(ae, false)
		assert_bool(af, true)
	}
	`
	ExecExpect(source)
}
