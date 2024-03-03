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
	. "falcon/utils"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type Parser struct {
	token      TokenKind
	lexeme     string
	nextToken  TokenKind
	nextLexeme string
	lexer      *Lexer
}

func syntaxError(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}

func (p *Parser) guarantee(cond bool, fmt string, args ...interface{}) {
	if !cond {
		syntaxError("SyntaxError: "+fmt, args...)
	}
}

func (p *Parser) lookNext() {
	p.nextToken, p.nextLexeme = p.lexer.NextToken()
}

func (p *Parser) consume() {
	if p.nextToken != INVALID {
		p.token, p.lexeme = p.nextToken, p.nextLexeme
		p.nextToken = INVALID
		p.nextLexeme = ""

	} else {
		p.token, p.lexeme = p.lexer.NextToken()
	}
}

func (p *Parser) parseParams() []AstExpr {
	params := make([]AstExpr, 0)
	p.consume()
	// Empty params
	if p.token == TK_RPAREN {
		p.consume()
		return params
	}

	// Non-empty params
	for p.token != TK_RPAREN {
		if p.token == TK_IDENT {
			paramName := p.lexeme
			p.consume()
			paramType := p.parseType()
			param := &VarExpr{Name: paramName}
			param.Type = paramType
			params = append(params, param)
		} else {
			p.guarantee(p.token == TK_COMMA, "Expected ,")
			p.consume()
		}
	}
	p.guarantee(p.token == TK_RPAREN, "Expected ')'")
	p.consume()
	return params
}

func (p *Parser) parseFuncDecl() *FuncDecl {
	p.guarantee(p.token == KW_FUNC, "Expected function definition")
	p.consume()

	fn := &FuncDecl{}
	fn.Name = p.lexeme
	p.consume()
	p.guarantee(p.token == TK_LPAREN, "Expected '('")
	fn.Params = p.parseParams()
	if retType := p.parseType(); retType != nil {
		fn.RetType = retType
	} else {
		fn.RetType = TVoid
	}

	if p.token == TK_LBRACE {
		// Function has {} body
		fn.Block = p.parseBlockDecl("body")
	} else {
		// Function does not have {} body, must be
		// a builtin or external function
		const RUNTIME_PREFIX = "rt_"
		fn.Builtin = true
		fn.Block = &BlockDecl{
			Name: "nativeBody",
			Stmts: []AstStmt{&SimpleStmt{
				Expr: &FuncCallExpr{
					Name: RUNTIME_PREFIX + fn.Name,
					Args: fn.Params,
				},
			}},
		}
	}
	return fn
}

func (p *Parser) parseBlockDecl(name string) *BlockDecl {
	p.guarantee(p.token == TK_LBRACE, "Expected '{'")
	elem := &BlockDecl{}
	elem.Name = name
	p.consume()
	elem.Stmts = p.parseStatementList()
	p.guarantee(p.token == TK_RBRACE, "Expected '}'")
	p.consume()
	return elem
}

func (p *Parser) parseStatementList() []AstStmt {
	stmts := make([]AstStmt, 0)
	for elem := p.parseStatement(); elem != nil; elem = p.parseStatement() {
		stmts = append(stmts, elem)
	}
	return stmts
}

func (p *Parser) parseControlStatement() AstStmt {
	switch p.token {
	case KW_IF:
		return p.parseIfStmt()
	case KW_FOR:
		return p.parseForStmt()
	case KW_DO:
		return p.parseDoWhileStmt()
	case KW_WHILE:
		return p.parseWhileStmt()
	default:
		return nil
	}
}

// The statement BNF is as follows:
//
//	statement = return_stmt
//				| let_stmt
//				| if_stmt
//				| for_stmt
//				| do_while_stmt
//				| while_stmt
//				| break_stmt
//				| continue_stmt
//				| simple_stmt

func (p *Parser) parseStatement() AstStmt {
	switch p.token {
	case KW_RETURN:
		return p.parseReturnStmt()
	case KW_LET:
		return p.parseLetStmt()
	case KW_IF:
		return p.parseIfStmt()
	case KW_FOR:
		return p.parseForStmt()
	case KW_DO:
		return p.parseDoWhileStmt()
	case KW_WHILE:
		return p.parseWhileStmt()
	case KW_BREAK:
		return p.parseBreakStmt()
	case KW_CONTINUE:
		return p.parseContinueStmt()
	case KW_PACKAGE:
		return p.parsePackageStmt()
	default:
		if elem := p.parseIncDecStmt(); elem != nil {
			return elem
		}
		return p.parseSimpleStmt()
	}
}

func (p *Parser) parseForStmt() AstStmt {
	p.guarantee(p.token == KW_FOR, "Expected for")
	p.consume()
	elem := &ForStmt{}

	expr := p.parseExpression()
	// for a;b;c {} form
	elem.Init = expr
	p.consume()
	elem.Cond = p.parseExpression()
	p.guarantee(p.token == TK_SEMICOLON, "Expected ;")
	p.consume()
	elem.Post = p.parseExpression()
	elem.Body = p.parseBlockDecl("forBody")
	return elem
}

func (p *Parser) parseDoWhileStmt() AstStmt {
	p.guarantee(p.token == KW_DO, "Expected do")
	p.consume()
	// do {...} while cond form
	elem := &DoWhileStmt{}
	elem.Body = p.parseBlockDecl("doWhileBody")
	p.guarantee(p.token == KW_WHILE, "Expected while")
	p.consume()
	elem.Cond = p.parseExpression()
	return elem
}

func (p *Parser) parseWhileStmt() AstStmt {
	p.guarantee(p.token == KW_WHILE, "Expected while")
	p.consume()
	// while cond {} form
	elem := &WhileStmt{}
	elem.Cond = p.parseExpression()
	elem.Body = p.parseBlockDecl("whileBody")
	return elem

}

func (p *Parser) parseReturnStmt() AstStmt {
	p.guarantee(p.token == KW_RETURN, "Expected return")
	p.consume()
	elem := &ReturnStmt{}
	elem.Expr = p.parseExpression()
	return elem
}

func (p *Parser) parseIncDecStmt() AstStmt {
	if p.token == TK_IDENT {
		p.lookNext()
		if p.nextToken == TK_INCREMENT || p.nextToken == TK_DECREMENT {
			val := &VarExpr{Name: p.lexeme}
			elem := &IncDecStmt{Var: val, Opt: p.nextToken}
			p.consume() // consume ident
			p.consume() // consume ++
			return elem
		}
	}
	return nil
}

func (p *Parser) parseSimpleStmt() AstStmt {
	if elem := p.parseExpression(); elem != nil {
		return &SimpleStmt{Expr: elem}
	}
	return nil
}

func (p *Parser) parseLetStmt() AstStmt {
	p.guarantee(p.token == KW_LET, "Expected let")
	p.consume()
	elem := &LetStmt{}
	elem.Var = p.parsePrimaryExpr().(*VarExpr)
	if p.token == TK_ASSIGN {
		// Let with init
		p.guarantee(p.token == TK_ASSIGN, "Expected =")
		p.consume()
		elem.Init = p.parseExpression()
	} else {
		// Let without init, set default value based on type
		t := elem.Var.Type
		Assert(t != nil, "Type of let variable must be specified")
		var defaultValue AstExpr
		switch t.Kind {
		case TypeInt:
			defaultValue = &IntExpr{Value: 0}
		case TypeLong:
			defaultValue = &LongExpr{Value: 0}
		case TypeShort:
			defaultValue = &ShortExpr{Value: 0}
		case TypeChar:
			defaultValue = &CharExpr{Value: 0}
		case TypeFloat:
			defaultValue = &FloatExpr{Value: 0}
		case TypeDouble:
			defaultValue = &DoubleExpr{Value: 0}
		case TypeString:
			defaultValue = &StrExpr{Value: ""}
		case TypeBool:
			defaultValue = &BoolExpr{Value: false}
		case TypeByte:
			defaultValue = &ByteExpr{Value: 0}
		default:
			syntaxError("Unsupported type %v for let variable", t)
		}
		defaultValue.SetType(t)
		elem.Init = defaultValue
	}
	return elem
}

func (p *Parser) parseIfStmt() AstStmt {
	p.guarantee(p.token == KW_IF, "Expected if")
	p.consume()
	elem := &IfStmt{}
	elem.Cond = p.parseExpression()
	elem.Then = p.parseBlockDecl("then")
	if p.token == KW_ELSE {
		p.consume()
		if p.token == TK_LBRACE {
			// if cond {} else {} form
			elem.Else = p.parseBlockDecl("else")
		} else if ctrl := p.parseControlStatement(); ctrl != nil {
			// if cond {} else if {} form
			// if cond {} else while{} also work
			elem.Else = ctrl
		} else {
			syntaxError("Expected if or { or other control statement after else")
		}
	}
	return elem
}

func (p *Parser) parseBreakStmt() AstStmt {
	p.guarantee(p.token == KW_BREAK, "Expected break")
	p.consume()
	return &BreakStmt{}
}

func (p *Parser) parseContinueStmt() AstStmt {
	p.guarantee(p.token == KW_CONTINUE, "Expected continue")
	p.consume()
	return &ContinueStmt{}
}

func (p *Parser) parsePackageStmt() AstStmt {
	p.guarantee(p.token == KW_PACKAGE, "Expected package")
	p.consume()
	elem := &PackageStmt{}
	elem.Name = p.lexeme
	p.consume()
	return elem
}

// The expression BNF is as follows:
//
//	expression = tenary_expr
//				| expression ( = | += | -= | *= | /= | %= ) tenary_expr
//	tenary_expr = logical_or_expr
//				| logical_or_expr ? expression : tenary_expr
//	logical_or_expr = logical_and_expr
//				| logical_and_expr || logical_or_expr
//	logical_and_expr = bit_or_expr
//				| bit_or_expr && logical_and_expr
//	bit_or_expr = bit_xor_expr
//				| bit_xor_expr | bit_or_expr
//	bit_xor_expr = bit_and_expr
//				| bit_and_expr ^ bit_xor_expr
//	bit_and_expr = equality_expr
//				| equality_expr & bit_and_expr
//	equality_expr = relational_expr
//				| relational_expr (== | !=) equality_expr
//	relational_expr = bitshift_expr
//				| bitshift_expr (< | <= | > | >=) relational_expr
//	bitshift_expr = add_expr
//				| add_expr (<< | >>) bitshift_expr
//	add_expr = mul_expr
//				| mul_expr (+ | -) add_expr
//	mul_expr = unary_expr
//				| unary_expr (* | / | %) mul_expr
//	unary_expr = (! | - | ~) unary_expr
//				| primary_expr
//	primary_expr = (3.14 | 32 | "foo" | 'c' | name | [] | true | null | func(){} | name.foo())
//				| (expression)
func (p *Parser) parsePrimaryExpr() AstExpr {
	switch p.token {
	case TK_IDENT:
		ident := p.lexeme
		p.consume()
		switch p.token {
		case TK_LPAREN: // foo()
			p.consume()
			val := &FuncCallExpr{}
			val.Name = ident
			for p.token != TK_RPAREN {
				val.Args = append(val.Args, p.parseExpression())
				if p.token == TK_COMMA {
					p.consume()
				}
			}
			p.guarantee(p.token == TK_RPAREN, "Expected )")
			p.consume()
			return val
		case TK_LBRACKET: //foo[bar]
			p.consume()
			val := &IndexExpr{}
			val.Name = ident
			val.Index = p.parseExpression()
			val.Type = nil
			p.guarantee(p.token == TK_RBRACKET, "Expected ']'")
			p.consume()
			return val
		default: // foo int
			val := &VarExpr{}
			val.Name = ident
			val.Type = p.parseType()
			return val
		}
	case TK_LBRACKET:
		p.consume()
		val := &ArrayExpr{}
		for p.token != TK_RBRACKET {
			val.Elems = append(val.Elems, p.parseExpression())
			if p.token == TK_COMMA {
				p.consume()
			}
		}
		// FIXME: elements may empty, we should rely on type inference to deduce
		// element type of array literal
		val.Type = &Type{TypeArray, val.Elems[0].GetType()}
		p.guarantee(p.token == TK_RBRACKET, "Expected ']'")
		p.consume()
		return val
	case KW_FUNC:
	case LIT_INT:
		elem := &IntExpr{}
		elem.Type = TInt
		var err error
		elem.Value, err = strconv.Atoi(p.lexeme)
		if err != nil {
			syntaxError("Failed to parse int literal %v", p.lexeme)
		}
		p.consume()
		return elem
	case LIT_LONG:
		elem := &LongExpr{}
		elem.Type = TLong
		var err error
		elem.Value, err = strconv.ParseInt(p.lexeme, 10, 64)
		if err != nil {
			syntaxError("Failed to parse long literal %v", p.lexeme)
		}
		p.consume()
		return elem
	case LIT_SHORT:
		elem := &ShortExpr{}
		elem.Type = TShort
		var err error
		val, err := strconv.ParseInt(p.lexeme, 10, 16)
		elem.Value = int16(val)
		if err != nil {
			syntaxError("Failed to parse short literal %v", p.lexeme)
		}
		p.consume()
		return elem
	case LIT_BYTE:
		elem := &ByteExpr{}
		elem.Type = TByte
		var err error
		val, err := strconv.ParseInt(p.lexeme, 10, 8)
		elem.Value = byte(val)
		if err != nil {
			syntaxError("Failed to parse byte literal %v", p.lexeme)
		}
		p.consume()
		return elem
	case LIT_FLOAT:
		// TODO: Complete me!
		utils.Unimplement()
	case LIT_DOUBLE:
		elem := &DoubleExpr{}
		elem.Type = TDouble
		var err error
		elem.Value, err = strconv.ParseFloat(p.lexeme, 64)
		if err != nil {
			syntaxError("Failed to parse double literal %v", p.lexeme)
		}
		p.consume()
		return elem
	case LIT_STR:
		elem := &StrExpr{}
		elem.Type = TString
		elem.Value = p.lexeme
		p.consume()
		return elem
	case LIT_CHAR:
		elem := &CharExpr{}
		elem.Type = TChar
		elem.Value = int8(p.lexeme[0])
		p.consume()
		return elem
	case KW_TRUE:
		elem := &BoolExpr{}
		elem.Type = TBool
		elem.Value = true
		p.consume()
		return elem
	case KW_FALSE:
		elem := &BoolExpr{}
		elem.Type = TBool
		elem.Value = false
		p.consume()
		return elem
	case KW_NULL:
		elem := &NullExpr{}
		p.consume()
		return elem
	case TK_LPAREN:
		p.consume()
		expr := p.parseExpression()
		p.guarantee(p.token == TK_RPAREN, "Expected )")
		p.consume()
		return expr
	}
	return nil
}

func (p *Parser) parseUnaryExpr() AstExpr {
	// !expr,-3.14,~1
	if Any(p.token, TK_MINUS, TK_LOGNOT, TK_BITNOT) {
		val := &UnaryExpr{}
		val.Opt = p.token
		p.consume()
		val.Left = p.parseUnaryExpr()
		return val
	} else if Any(p.token,
		LIT_DOUBLE, LIT_INT, LIT_LONG, LIT_SHORT, LIT_BYTE,
		LIT_CHAR, LIT_FLOAT, LIT_STR,
		TK_IDENT, TK_LPAREN, TK_LBRACKET, KW_TRUE, KW_FALSE,
		KW_NULL, KW_FUNC) {
		// 3.14,32,"foo",'c',name,[],true,null,func(){},name.foo()
		prim := p.parsePrimaryExpr()
		if p.token == TK_DOT {
			// todo
		} else {
			return prim
		}
	}
	return nil
}

func (p *Parser) parseMulExpr() AstExpr {
	left := p.parseUnaryExpr()
	for Any(p.token, TK_TIMES, TK_DIV, TK_MOD) {
		val := &BinaryExpr{Opt: p.token}
		p.consume()
		val.Left = left
		val.Right = p.parseUnaryExpr()
		left = val
	}
	return left
}

func (p *Parser) parseAddExpr() AstExpr {
	left := p.parseMulExpr()
	for Any(p.token, TK_PLUS, TK_MINUS) {
		val := &BinaryExpr{Opt: p.token}
		p.consume()
		val.Left = left
		val.Right = p.parseMulExpr()
		left = val
	}
	return left

}

func (p *Parser) parseBitshiftExpr() AstExpr {
	left := p.parseAddExpr()
	for Any(p.token, TK_LSHIFT, TK_RSHIFT) {
		val := &BinaryExpr{Opt: p.token}
		p.consume()
		val.Left = left
		val.Right = p.parseAddExpr()
		left = val
	}
	return left
}

func (p *Parser) parseRelationalExpr() AstExpr {
	left := p.parseBitshiftExpr()
	for Any(p.token, TK_GT, TK_GE, TK_LT, TK_LE) {
		val := &BinaryExpr{}
		val.Opt = p.token
		p.consume()
		val.Left = left
		val.Right = p.parseBitshiftExpr()
		left = val
	}
	return left
}
func (p *Parser) parseEqualityExpr() AstExpr {
	left := p.parseRelationalExpr()
	for p.token == TK_EQ || p.token == TK_NE {
		val := &BinaryExpr{}
		val.Opt = p.token
		p.consume()
		val.Left = left
		val.Right = p.parseRelationalExpr()
		left = val
	}
	return left
}

func (p *Parser) parseBitandExpr() AstExpr {
	left := p.parseEqualityExpr()
	for p.token == TK_BITAND {
		val := &BinaryExpr{}
		val.Opt = p.token
		p.consume()
		val.Left = left
		val.Right = p.parseEqualityExpr()
		left = val
	}
	return left
}

func (p *Parser) parseBitxorExpr() AstExpr {
	left := p.parseBitandExpr()
	for p.token == TK_BITXOR {
		val := &BinaryExpr{}
		val.Opt = p.token
		p.consume()
		val.Left = left
		val.Right = p.parseBitandExpr()
		left = val
	}
	return left

}

func (p *Parser) parseBitorExpr() AstExpr {
	left := p.parseBitxorExpr()
	for p.token == TK_BITOR {
		val := &BinaryExpr{}
		val.Opt = p.token
		p.consume()
		val.Left = left
		val.Right = p.parseBitxorExpr()
		left = val
	}
	return left
}

func (p *Parser) parseLogicalAndExpr() AstExpr {
	left := p.parseBitorExpr()
	for p.token == TK_LOGAND {
		val := &BinaryExpr{}
		val.Opt = p.token
		p.consume()
		val.Left = left
		val.Right = p.parseBitorExpr()
		left = val
	}
	return left
}
func (p *Parser) parseLogicalOrExpr() AstExpr {
	left := p.parseLogicalAndExpr()
	for p.token == TK_LOGOR {
		val := &BinaryExpr{}
		val.Opt = p.token
		p.consume()
		val.Left = left
		val.Right = p.parseLogicalAndExpr()
		left = val
	}
	return left
}

func (p *Parser) parseConditionalExpr() AstExpr {
	left := p.parseLogicalOrExpr()
	if p.token == TK_QUESTION {
		val := &ConditionalExpr{}
		val.Cond = left
		p.consume()
		val.Then = p.parseExpression()
		p.guarantee(p.token == TK_COLON, "Expected :")
		p.consume()
		val.Else = p.parseConditionalExpr()
		return val
	}
	return left

}

func (p *Parser) parseExpression() AstExpr {
	left := p.parseConditionalExpr()
	if Any(p.token, TK_ASSIGN, TK_PLUS_AGN, TK_MINUS_AGN, TK_TIMES_AGN,
		TK_DIV_AGN, TK_MOD_AGN, TK_BITAND_AGN, TK_BITOR_AGN, TK_BITXOR_AGN,
		TK_LSHIFT_AGN, TK_RSHIFT_AGN) {
		val := &AssignExpr{}
		val.Opt = p.token
		val.Left = left
		p.consume()
		val.Right = p.parseExpression()
		return val

	}
	return left
}

func (p *Parser) parseType() *Type {
	switch p.token {
	case KW_TYPE_INT:
		p.consume()
		return TInt
	case KW_TYPE_LONG:
		p.consume()
		return TLong
	case KW_TYPE_SHORT:
		p.consume()
		return TShort
	case KW_TYPE_CHAR:
		p.consume()
		return TChar
	case KW_TYPE_FLOAT:
		p.consume()
		return TFloat
	case KW_TYPE_DOUBLE:
		p.consume()
		return TDouble
	case KW_TYPE_STR:
		p.consume()
		return TString
	case KW_TYPE_BYTE:
		p.consume()
		return TByte
	case KW_TYPE_BOOL:
		p.consume()
		return TBool
	case TK_LBRACKET:
		p.consume()
		p.guarantee(p.token == TK_RBRACKET, "Expected ']'")
		p.consume()
		return &Type{TypeArray, p.parseType()}
	}
	return nil
}
func NewParser(file *os.File) *Parser {
	p := new(Parser)
	p.lexer = NewLexer(file)
	return p
}

func (p *Parser) Parse() AstNode {
	root := &PackageDecl{}
	root.Source = p.lexer.fileName
	p.consume()
	if p.token == TK_EOF {
		return root
	}
	for p.token != TK_EOF {
		if p.token == KW_FUNC {
			// Parse function
			root.Func = append(root.Func, p.parseFuncDecl())
		} else {
			// Parse statement
			stmt := p.parseStatement()
			root.List = append(root.List, stmt)
		}
	}
	return root
}

func ParseFile(fileName string) *PackageDecl {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	parser := NewParser(file)
	return parser.Parse().(*PackageDecl)
}

func ParseText(text string) *PackageDecl {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "falcon")
	if err != nil {
		fmt.Errorf("Cannot create temporary file %v", err)
		os.Exit(1)
	}
	_, err = tmpFile.WriteString(text)
	if err != nil {
		fmt.Errorf("Failed to write to temporary file %v", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name())
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		panic(err)
	}
	defer file.Close()
	parser := NewParser(file)
	return parser.Parse().(*PackageDecl)
}
