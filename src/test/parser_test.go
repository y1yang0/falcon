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
	"falcon/utils"
	"testing"
)

func MustBe(cond bool) {
	if !cond {
		panic("unexpect")
	}
}

func MustNotNull(v interface{}) {
	if v == nil {
		panic("Unexpected nil")
	}
}

func MustNull(v interface{}) {
	if v != nil {
		panic("Unexpected non nil")
	}
}

func SyntaxError() {
	panic("SyntaxError")
}

func checkSimpleStmt(root ast.AstDecl) *ast.SimpleStmt {
	st, ok := root.(*ast.PackageDecl).List[0].(*ast.SimpleStmt)
	if !ok {
		SyntaxError()
	}
	return st
}

func checkLetStmt(root ast.AstDecl) *ast.LetStmt {
	st, ok := root.(*ast.PackageDecl).List[0].(*ast.LetStmt)
	if !ok {
		SyntaxError()
	}
	return st
}

func checkIfStmt(root ast.AstDecl) *ast.IfStmt {
	st, ok := root.(*ast.PackageDecl).List[0].(*ast.IfStmt)
	if !ok {
		SyntaxError()
	}
	return st
}

func checkForStmt(root ast.AstDecl) *ast.ForStmt {
	st, ok := root.(*ast.PackageDecl).List[0].(*ast.ForStmt)
	if !ok {
		SyntaxError()
	}
	return st
}

func checkFunCallExpr(v *ast.SimpleStmt, name string) *ast.FuncCallExpr {
	st, ok := v.Expr.(*ast.FuncCallExpr)
	if !ok {
		SyntaxError()
	}
	if st.Name != name {
		SyntaxError()
	}
	return st
}

func checkBinaryExpr(v *ast.SimpleStmt, opt ast.TokenKind) *ast.BinaryExpr {
	st, ok := v.Expr.(*ast.BinaryExpr)
	if !ok {
		SyntaxError()
	}
	if st.Opt != opt {
		SyntaxError()
	}
	return st
}

func checkUnaryExpr(v *ast.SimpleStmt, opt ast.TokenKind) *ast.UnaryExpr {
	st, ok := v.Expr.(*ast.UnaryExpr)
	if !ok {
		SyntaxError()
	}
	if st.Opt != opt {
		SyntaxError()
	}
	return st
}

func checkConditionalExpr(v *ast.SimpleStmt) *ast.ConditionalExpr {
	st, ok := v.Expr.(*ast.ConditionalExpr)
	if !ok {
		SyntaxError()
	}
	return st
}

func checkAssignExpr(v *ast.SimpleStmt, opt ast.TokenKind) {
	st, ok := v.Expr.(*ast.AssignExpr)
	if !ok {
		SyntaxError()
	}
	if st.Opt != opt {
		SyntaxError()
	}
}

func checkBlockDecl(node ast.AstNode) *ast.BlockDecl {
	st, ok := node.(*ast.BlockDecl)
	if !ok {
		SyntaxError()
	}
	return st
}

func TestParser0(t *testing.T) {
	root := ast.ParseText("//")
	utils.Assert(len(root.List) == 0, "unexpect")
	utils.Assert(len(root.Func) == 0, "unexpect")

	root = ast.ParseText("////")
	utils.Assert(len(root.List) == 0, "unexpect")
	utils.Assert(len(root.Func) == 0, "unexpect")

	root = ast.ParseText(`//
	//
	//`)
	utils.Assert(len(root.List) == 0, "unexpect")
	utils.Assert(len(root.Func) == 0, "unexpect")

	root = ast.ParseText(`//line1
	//line2
	//line2`)
	utils.Assert(len(root.List) == 0, "unexpect")
	utils.Assert(len(root.Func) == 0, "unexpect")

	root = ast.ParseText(`//		tab		
			//line2
	    //line2	`)
	utils.Assert(len(root.List) == 0, "unexpect")
	utils.Assert(len(root.Func) == 0, "unexpect")

	root = ast.ParseText(`//line1
	//line2
	//line2
	func main() {
	}`)
	utils.Assert(len(root.List) == 0, "unexpect")
}

func TestParser1(t *testing.T) {
	root := ast.ParseText("3+2")
	a := checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_PLUS)

	root = ast.ParseText("3-2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_MINUS)

	root = ast.ParseText("3*2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_TIMES)

	root = ast.ParseText("3/2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_DIV)

	root = ast.ParseText("3%2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_MOD)

	root = ast.ParseText("3&2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_BITAND)

	root = ast.ParseText("3|2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_BITOR)

	root = ast.ParseText("3^2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_BITXOR)

	root = ast.ParseText("~32")
	a = checkSimpleStmt(root)
	checkUnaryExpr(a, ast.TK_BITNOT)

	root = ast.ParseText("3>2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_GT)

	root = ast.ParseText("3>=2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_GE)

	root = ast.ParseText("3<2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_LT)

	root = ast.ParseText("3<=2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_LE)

	root = ast.ParseText("3==2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_EQ)

	root = ast.ParseText("3!=2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_NE)

	root = ast.ParseText("3<<2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_LSHIFT)

	root = ast.ParseText("3>>2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_RSHIFT)

	root = ast.ParseText("3&&2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_LOGAND)

	root = ast.ParseText("3||2")
	a = checkSimpleStmt(root)
	checkBinaryExpr(a, ast.TK_LOGOR)

	root = ast.ParseText("!3")
	a = checkSimpleStmt(root)
	checkUnaryExpr(a, ast.TK_LOGNOT)
}

func TestParser2(t *testing.T) {
	root := ast.ParseText("a+=3")
	a := checkSimpleStmt(root)
	checkAssignExpr(a, ast.TK_PLUS_AGN)

	root = ast.ParseText("a-=3")
	a = checkSimpleStmt(root)
	checkAssignExpr(a, ast.TK_MINUS_AGN)

	root = ast.ParseText("a%=3")
	a = checkSimpleStmt(root)
	checkAssignExpr(a, ast.TK_MOD_AGN)

	root = ast.ParseText("a*=3")
	a = checkSimpleStmt(root)
	checkAssignExpr(a, ast.TK_TIMES_AGN)

	root = ast.ParseText("a/=3")
	a = checkSimpleStmt(root)
	checkAssignExpr(a, ast.TK_DIV_AGN)

	root = ast.ParseText("foo()")
	a = checkSimpleStmt(root)
	checkFunCallExpr(a, "foo")
}

func TestParser3(t *testing.T) {
	root := ast.ParseText("let p =3")
	a := checkLetStmt(root)
	MustNotNull(a.Init)
	MustNotNull(a.Var)

	root = ast.ParseText("let p int =3")
	a = checkLetStmt(root)
	MustNotNull(a.Init)
	MustNotNull(a.Var.GetType())
	MustNotNull(a.Var)

	root = ast.ParseText(`if p==3 {} else {}`)
	a1 := checkIfStmt(root)
	MustNotNull(a1)

	root = ast.ParseText(`if p==3 {} else if {}`)
	a1 = checkIfStmt(root)
	MustNotNull(a1)

	root = ast.ParseText(`if p==3 {} else if {} else {}`)
	a1 = checkIfStmt(root)
	MustNotNull(a1)

	root = ast.ParseText(`if p==3 {} else while true {}`)
	a1 = checkIfStmt(root)
	MustNotNull(a1)

	root = ast.ParseText(`if p==3 { if p==4{} } else { if p==5{}else{ if p==6{}else{}} }`)
	a1 = checkIfStmt(root)
	MustNotNull(a1)

	root = ast.ParseText(`if v==1 {
        r=1
    } else {
        if v==2 {
            r=2
        } else {
            if v==3 {
                r=3
            } else {
                if v==4 {
                    r=4
                } else {
                    if v==5 {
                        r=5
                    } else {
                        r=6
                    }
                }
            }
        }
    }`)
	a1 = checkIfStmt(root)
	MustNotNull(a1)

	root = ast.ParseText(`if true {} else {}`)
	a1 = checkIfStmt(root)
	MustNotNull(a1)

	root = ast.ParseText(`if true {}`)
	a1 = checkIfStmt(root)
	MustNotNull(a1)

	root = ast.ParseText(`for i=0;i<10;i+=1{sum += i }`)
	a2 := checkForStmt(root)
	MustNotNull(a2)
	MustNotNull(a2.Init)
	MustNotNull(a2.Cond)
	MustNotNull(a2.Post)
	MustBe(len(checkBlockDecl(a2.Body).Stmts) > 0)

	root = ast.ParseText(`for i=0;i<10;i+=1{}`)
	a2 = checkForStmt(root)
	MustNotNull(a2)
	MustNotNull(a2.Init)
	MustNotNull(a2.Cond)
	MustNotNull(a2.Post)
	MustBe(len(checkBlockDecl(a2.Body).Stmts) == 0)

	root = ast.ParseText(`for i=1;i<=100;i+=1{sum += i }`)
	a2 = checkForStmt(root)
	MustNotNull(a2)
	MustNotNull(a2.Init)
	MustNotNull(a2.Cond)
	MustNotNull(a2.Post)
	MustBe(len(checkBlockDecl(a2.Body).Stmts) > 0)

	root = ast.ParseText(`for i=1;i<=100;i+=1{ for k=0;k<10;k+=1{cprint(i+j)} }`)
	a2 = checkForStmt(root)
	MustNotNull(a2)
	MustNotNull(a2.Init)
	MustNotNull(a2.Cond)
	MustNotNull(a2.Post)
	MustBe(len(checkBlockDecl(a2.Body).Stmts) > 0)
}

func mustBinaryExpr(v ast.AstNode) *ast.BinaryExpr {
	st, ok := v.(*ast.BinaryExpr)
	if !ok {
		SyntaxError()
	}
	return st
}

// Test for associativity and precedence
func TestParser4(t *testing.T) {
	root := ast.ParseText("3+2-2")
	// It should be (3+2)-2
	a := checkSimpleStmt(root)
	e := mustBinaryExpr(a.Expr)
	MustBe(e.Opt == ast.TK_MINUS)
	e = mustBinaryExpr(mustBinaryExpr(a.Expr).Left)
	MustBe(e.Opt == ast.TK_PLUS)

	root = ast.ParseText("-3+2-5")
	// It should be (-3+2)-2
	a = checkSimpleStmt(root)
	e = mustBinaryExpr(a.Expr)
	MustBe(e.Opt == ast.TK_MINUS)
	e = mustBinaryExpr(mustBinaryExpr(a.Expr).Left)
	MustBe(e.Opt == ast.TK_PLUS)

	root = ast.ParseText("3+(2-2)")
	// It should be 3+(2-2)
	a = checkSimpleStmt(root)
	e = mustBinaryExpr(a.Expr)
	MustBe(e.Opt == ast.TK_PLUS)
	e = mustBinaryExpr(mustBinaryExpr(a.Expr).Right)
	MustBe(e.Opt == ast.TK_MINUS)

	root = ast.ParseText("3+2*2")
	// It should be 3+(2*2)
	a = checkSimpleStmt(root)
	e = mustBinaryExpr(a.Expr)
	MustBe(e.Opt == ast.TK_PLUS)
	e = mustBinaryExpr(mustBinaryExpr(a.Expr).Right)
	MustBe(e.Opt == ast.TK_TIMES)

	root = ast.ParseText("3/2*2")
	// It should be (3/2)*2
	a = checkSimpleStmt(root)
	e = mustBinaryExpr(a.Expr)
	MustBe(e.Opt == ast.TK_TIMES)
	e = mustBinaryExpr(mustBinaryExpr(a.Expr).Left)
	MustBe(e.Opt == ast.TK_DIV)

	// TODO: Verify all precedence and associativity
}

func TestParser5(t *testing.T) {
	root := ast.ParseText("true?false:true")
	a := checkSimpleStmt(root)
	checkConditionalExpr(a)

	root = ast.ParseText("a==b?foo():bar[index()]")
	a = checkSimpleStmt(root)
	checkConditionalExpr(a)

	root = ast.ParseText("foo()?():()")
	a = checkSimpleStmt(root)
	checkConditionalExpr(a)
}

func TestParseLoop(t *testing.T) {
	root := ast.ParseText(`for ;true;{}`)
	a2 := checkForStmt(root)
	MustNotNull(a2)
	MustNull(a2.Init)
	MustNotNull(a2.Cond)
	MustNull(a2.Post)
	MustBe(len(checkBlockDecl(a2.Body).Stmts) == 0)
}
