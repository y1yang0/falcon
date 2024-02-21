package parser;

import ast.Token;
import ast.TokenKind;

import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;

import static ast.TokenKind.*;


public class Lexer {
    private String fileName;
    private BufferedReader reader;
    private int line;
    private int column;

    public Lexer(String fileName) {
        init(fileName);
    }

    public String getFileName() {
        return fileName;
    }

    private void init(String fileName) {
        try {
            reader = new BufferedReader(new FileReader(fileName));
            this.fileName = fileName;
            line = 1;
            column = 0;
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    private int next() {
        try {
            int b = reader.read();
            if (b == -1) {
                return -1;
            }
            return b;
        } catch (IOException e) {
            e.printStackTrace();
            return -1;
        }
    }

    private int peek() {
        try {
            reader.mark(1);
            int b = reader.read();
            reader.reset();
            if (b == -1) {
                return -1;
            }
            return b;
        } catch (IOException e) {
            e.printStackTrace();
            return -1;
        }
    }

    public Token nextToken() {
        final int EOF = -1;
        int c = next();
        if (c == EOF) {
            return new Token(TK_EOF, "");
        }
        if (Character.isWhitespace(c)) {
            while (Character.isWhitespace(c)) {
                if (c == '\n') {
                    line++;
                    column = 0;
                }
                c = next();
            }
            if (c == EOF) {
                return new Token(TK_EOF, "");
            }
        }
        if (c == '/') {
            while (peek() == '/') {
                while (c != '\n' && c != EOF) {
                    c = next();
                }

                while (c == '\n' || c == ' ' || c == '\t' || c == '\r') {
                    line++;
                    column = 0;
                    c = next();
                }
                if (c == '/') {
                    continue;
                }
                if (c == EOF) {
                    return new Token(TK_EOF, "");
                }
            }
        }
        if (c >= '0' && c <= '9') {
            StringBuilder lexeme = new StringBuilder();
            boolean isDouble = false;
            int cn = peek();
            while ((cn >= '0' && cn <= '9') || (!isDouble && cn == '.')) {
                if (c == '.') {
                    isDouble = true;
                }
                c = next();
                cn = peek();
                lexeme.append((char) c);
            }
            if (!isDouble) {
                return new Token(LIT_INT, lexeme.toString());
            } else {
                return new Token(LIT_DOUBLE, lexeme.toString());
            }
        }
        if (Character.isLetter(c) || c == '_') {
            StringBuilder lexeme = new StringBuilder();
            int cn = peek();
            while (Character.isLetter(cn) || Character.isDigit(cn) || cn == '_') {
                c = next();
                lexeme.append((char) c);
                cn = peek();
            }
            String lexemeStr = lexeme.toString();
            if (TokenKind.isKeyword(lexemeStr)) {
                return new Token(TokenKind.get(lexemeStr), lexemeStr);
            }
            return new Token(TK_IDENT, lexemeStr);
        }
        switch (c) {
            case '\'':
                StringBuilder lexeme = new StringBuilder();
                int nextChar = next();
                if (nextChar != '\'') {
                    lexeme.append((char) nextChar);
                    if (peek() != '\'') {
                        throw new RuntimeException("a character literal should surround with single-quote");
                    }
                    c = next();
                }
                return new Token(LIT_CHAR, lexeme.toString());
            case '"':
                StringBuilder lexemeStr = new StringBuilder();
                int cn = peek();
                while (cn != '"') {
                    c = next();
                    lexemeStr.append((char) c);
                    cn = peek();
                }
                c = next();
                return new Token(LIT_STR, lexemeStr.toString());
            case '[':
                return new Token(TK_LBRACKET, "[");
            case ']':
                return new Token(TK_RBRACKET, "]");
            case '{':
                return new Token(TK_LBRACE, "{");
            case '}':
                return new Token(TK_RBRACE, "}");
            case '(':
                return new Token(TK_LPAREN, "(");
            case ')':
                return new Token(TK_RPAREN, ")");
            case ',':
                return new Token(TK_COMMA, ",");
            case ';':
                return new Token(TK_SEMICOLON, ";");
            case ':':
                return new Token(TK_COLON, ":");
            case '+':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_PLUS_AGN, "+=");
                }
                return new Token(TK_PLUS, "+");
            case '-':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_MINUS_AGN, "-=");
                }
                return new Token(TK_MINUS, "-");
            case '*':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_TIMES_AGN, "*=");
                }
                return new Token(TK_TIMES, "*");
            case '/':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_DIV_AGN, "/=");
                }
                return new Token(TK_DIV, "/");
            case '%':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_MOD_AGN, "%=");
                }
                return new Token(TK_MOD, "%");
            case '~':
                return new Token(TK_BITNOT, "~");
            case '.':
                return new Token(TK_DOT, ".");
            case '?':
                return new Token(TK_QUESTION, "?");
            case '=':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_EQ, "==");
                } else if (peek() == '>') {
                    c = next();
                    return new Token(TK_MATCH, "=>");
                }
                return new Token(TK_ASSIGN, "=");
            case '!':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_NE, "!=");
                }
                return new Token(TK_LOGNOT, "!");
            case '|':
                if (peek() == '|') {
                    c = next();
                    return new Token(TK_LOGOR, "||");
                } else if (peek() == '=') {
                    c = next();
                    return new Token(TK_BITOR_AGN, "|=");
                }
                return new Token(TK_BITOR, "|");
            case '&':
                if (peek() == '&') {
                    c = next();
                    return new Token(TK_LOGAND, "&&");
                } else if (peek() == '=') {
                    c = next();
                    return new Token(TK_BITAND_AGN, "&=");
                }
                return new Token(TK_BITAND, "&");
            case '^':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_BITXOR_AGN, "^=");
                }
                return new Token(TK_BITXOR, "^");
            case '>':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_GE, ">=");
                } else if (peek() == '>') {
                    c = next();
                    if (peek() == '=') {
                        c = next();
                        return new Token(TK_RSHIFT_AGN, ">>=");
                    }
                    return new Token(TK_RSHIFT, ">>");
                }
                return new Token(TK_GT, ">");
            case '<':
                if (peek() == '=') {
                    c = next();
                    return new Token(TK_LE, "<=");
                } else if (peek() == '<') {
                    c = next();
                    if (peek() == '=') {
                        c = next();
                        return new Token(TK_LSHIFT_AGN, "<<=");
                    }
                    return new Token(TK_LSHIFT, "<<");
                }
                return new Token(TK_LT, "<");
            default:
                throw new RuntimeException("unknown token '" + (char) c + "'");
        }
    }

    public void printTokenized(String fileName) {
        try {
            BufferedReader fileReader = new BufferedReader(new FileReader(fileName));
            Lexer lexer = new Lexer(fileName);
            Token token;
            while ((token = lexer.nextToken()).getKind() != TK_EOF) {
                System.out.printf("[%s, \"%s\"]\n", token.getKind(), token.getLexeme());
            }
            fileReader.close();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}


