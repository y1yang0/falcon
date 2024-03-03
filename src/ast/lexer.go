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
	"bufio"
	"falcon/utils"
	"fmt"
	"io"
	"os"
)

type Lexer struct {
	fileName string
	reader   *bufio.Reader
	line     int32
	column   int32
}

func NewLexer(file *os.File) *Lexer {
	lexer := new(Lexer)
	lexer.reader = bufio.NewReader(file)
	lexer.fileName = file.Name()
	lexer.line = 1
	lexer.column = 0
	return lexer
}

func (lexer *Lexer) next() int32 {
	b, readErr := lexer.reader.ReadByte()
	if readErr != nil {
		if readErr != io.EOF {
			panic(readErr)
		}
		return -1
	}
	return int32(b)
}

func (lexer *Lexer) peek() int32 {
	b, readErr := lexer.reader.Peek(1)
	if readErr != nil {
		if readErr != io.EOF {
			panic(readErr)
		}
		return -1
	}
	return int32(b[0])
}

func (lexer *Lexer) throwSyntaxError(format string, args ...interface{}) {
	finalFormat := "\033[1;31;40m%s:%d:%d %s\033[0m\n"
	fmt.Errorf(finalFormat, lexer.fileName, lexer.line, lexer.column, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (lexer *Lexer) NextToken() (TokenKind, string) {
	const EOF = -1
	c := lexer.next()

	if c == EOF {
		return TK_EOF, ""
	}
	// blank or whitespace
	if utils.Any(c, ' ', '\n', '\r', '\t') {
		for utils.Any(c, ' ', '\n', '\r', '\t') {
			if c == '\n' {
				lexer.line++
				lexer.column = 0
			}
			c = lexer.next()
		}
		if c == EOF {
			return TK_EOF, ""
		}
	}

	// comment
	if c == '/' {
	anotherComment:
		if lexer.peek() == '/' {
			for c != '\n' && c != EOF {
				c = lexer.next()
			}
			// consume newlines or whitespace
			for c == '\n' || c == ' ' || c == '\t' || c == '\r' {
				lexer.line++
				lexer.column = 0
				c = lexer.next()
			}
			if c == '/' {
				goto anotherComment
			}
			if c == EOF {
				return TK_EOF, ""
			}
		}
	}
	// number
	if c >= '0' && c <= '9' {
		lexeme := string(c)
		isDouble := false
		cn := lexer.peek()
		for (cn >= '0' && cn <= '9') || (!isDouble && cn == '.') {
			if c == '.' {
				isDouble = true
			}
			c = lexer.next()
			cn = lexer.peek()
			lexeme += string(c)
		}
		// if number is followed by a letter, it's a numeric suffix
		const LongSuffix = 'L'
		const FloatSuffix = 'F'
		const ShortSuffix = 'S'
		const ByteSuffix = 'B'
		switch cn {
		case LongSuffix:
			// long literal, e.g. 123L
			lexer.next()
			return LIT_LONG, lexeme
		case FloatSuffix:
			// float literal, e.g. 123.0F
			lexer.next()
			return LIT_FLOAT, lexeme
		case ShortSuffix:
			// short literal, e.g. 123S
			lexer.next()
			return LIT_SHORT, lexeme
		case ByteSuffix:
			// byte literal, e.g. 123B
			lexer.next()
			return LIT_BYTE, lexeme
		}
		if !isDouble {
			return LIT_INT, lexeme
		} else {
			return LIT_DOUBLE, lexeme
		}
	}
	// identifier
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
		lexeme := string(c)
		cn := lexer.peek()
		for (cn >= 'a' && cn <= 'z') || (cn >= 'A' && cn <= 'Z') ||
			(cn >= '0' && cn <= '9') || cn == '_' {
			c = lexer.next()
			lexeme += string(c)
			cn = lexer.peek()
		}
		if tk, exist := Keywords[lexeme]; exist {
			return tk, lexeme
		}
		return TK_IDENT, lexeme
	}

	// ok, it's a symbol, let's check it then
	switch c {
	case '\'':
		lexeme := ""
		nextChar := lexer.next()
		if nextChar != '\'' {
			lexeme += string(nextChar)
			if lexer.peek() != '\'' {
				lexer.throwSyntaxError("a character literal should surround with single-quote")
			}
			lexer.next()
		}
		return LIT_CHAR, lexeme
	case '"':
		lexeme := ""
		cn := lexer.peek()
		for cn != '"' {
			c = lexer.next()
			lexeme += string(c)
			cn = lexer.peek()
		}
		lexer.next()
		return LIT_STR, lexeme
	case '[':
		return TK_LBRACKET, "["
	case ']':
		return TK_RBRACKET, "]"
	case '{':
		return TK_LBRACE, "{"
	case '}':
		return TK_RBRACE, "}"
	case '(':
		return TK_LPAREN, "("
	case ')':
		return TK_RPAREN, ")"
	case ',':
		return TK_COMMA, ","
	case ';':
		return TK_SEMICOLON, ";"
	case ':':
		return TK_COLON, ":"
	case '+':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_PLUS_AGN, "+="
		} else if lexer.peek() == '+' {
			lexer.next()
			return TK_INCREMENT, "++"
		}
		if lexer.peek() == '+' {
			c = lexer.next()
			return TK_INCREMENT, "++"
		}
		return TK_PLUS, "+"
	case '-':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_MINUS_AGN, "-="
		} else if lexer.peek() == '-' {
			lexer.next()
			return TK_DECREMENT, "--"
		}
		if lexer.peek() == '-' {
			c = lexer.next()
			return TK_DECREMENT, "--"
		}
		return TK_MINUS, "-"
	case '*':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_TIMES_AGN, "*="
		}
		return TK_TIMES, "*"
	case '/':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_DIV_AGN, "/="
		}
		return TK_DIV, "/"
	case '%':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_MOD_AGN, "%="
		}
		return TK_MOD, "%"
	case '~':
		{
			return TK_BITNOT, "~"
		}
	case '.':
		return TK_DOT, "."
	case '?':
		return TK_QUESTION, "?"
	case '=':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_EQ, "=="
		} else if lexer.peek() == '>' {
			lexer.next()
			return TK_MATCH, "=>"
		}
		return TK_ASSIGN, "="
	case '!':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_NE, "!="
		}
		return TK_LOGNOT, "!"
	case '|':
		if lexer.peek() == '|' {
			lexer.next()
			return TK_LOGOR, "||"
		} else if lexer.peek() == '=' {
			lexer.next()
			return TK_BITOR_AGN, "|="
		}
		return TK_BITOR, "|"
	case '&':
		if lexer.peek() == '&' {
			lexer.next()
			return TK_LOGAND, "&&"
		} else if lexer.peek() == '=' {
			lexer.next()
			return TK_BITAND_AGN, "&="
		}
		return TK_BITAND, "&"
	case '^':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_BITXOR_AGN, "^="
		}
		return TK_BITXOR, "^"
	case '>':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_GE, ">="
		} else if lexer.peek() == '>' {
			lexer.next()
			if lexer.peek() == '=' {
				lexer.next()
				return TK_RSHIFT_AGN, ">>="
			}
			return TK_RSHIFT, ">>"
		}
		return TK_GT, ">"
	case '<':
		if lexer.peek() == '=' {
			lexer.next()
			return TK_LE, "<="
		} else if lexer.peek() == '<' {
			lexer.next()
			if lexer.peek() == '=' {
				lexer.next()
				return TK_LSHIFT_AGN, "<<="
			}
			return TK_LSHIFT, "<<"
		}
		return TK_LT, "<"
	default:
		lexer.throwSyntaxError("Bad character: %c", c)
	}
	return TK_EOF, ""
}

func PrintTokenized(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	lexer := NewLexer(file)
	for c, l := lexer.NextToken(); c != TK_EOF; c, l = lexer.NextToken() {
		fmt.Printf("[%v, \"%v\"]\n", c, l)
	}
}
