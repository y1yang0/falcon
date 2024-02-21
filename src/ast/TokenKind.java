package ast;

import java.util.Set;

public enum TokenKind {
    INVALID("<invalid>"),
    TK_IDENT("<identifier>"),
    TK_EOF("<eof>"),
    LIT_INT("<integer>"),
    LIT_LONG("<long>"),
    LIT_SHORT("<short>"),
    LIT_DOUBLE("<decimal>"),
    LIT_FLOAT("<float>"),
    LIT_CHAR("<character>"),
    LIT_BOOL("<bool>"),
    LIT_BYTE("<byte>"),
    LIT_STR("<string>"),

    TK_BITAND("&"),
    TK_BITOR("|"),
    TK_BITNOT("~"),
    TK_BITXOR("^"),
    TK_LOGAND("&&"),
    TK_LOGOR("||"),
    TK_LOGNOT("!"),
    TK_PLUS("+"),
    TK_MINUS("-"),
    TK_TIMES("*"),
    TK_DIV("/"),
    TK_MOD("%"),
    TK_EQ("=="),
    TK_NE("!="),
    TK_GT(">"),
    TK_GE(">="),
    TK_LT("<"),
    TK_LE("<="),
    TK_RSHIFT(">>"),
    TK_LSHIFT("<<"),

    TK_ASSIGN("="),
    TK_PLUS_AGN("+="),
    TK_MINUS_AGN("-="),
    TK_TIMES_AGN("*="),
    TK_DIV_AGN("/="),
    TK_MOD_AGN("%="),
    TK_RSHIFT_AGN(">>="),
    TK_LSHIFT_AGN("<<="),
    TK_BITXOR_AGN("^="),
    TK_BITAND_AGN("&="),
    TK_BITOR_AGN("|="),
    TK_MATCH("=>"),
    TK_COMMA(","),
    TK_LPAREN("("),
    TK_RPAREN(")"),
    TK_LBRACE("{"),
    TK_RBRACE("}"),
    TK_LBRACKET("["),
    TK_RBRACKET("]"),
    TK_SEMICOLON(";"),
    TK_COLON(":"),
    TK_DOT("."),
    TK_QUESTION("?"),

    KW_IF("if"),
    KW_ELSE("else"),
    KW_TRUE("true"),
    KW_FALSE("false"),
    KW_WHILE("while"),
    KW_FOR("for"),
    KW_NULL("null"),
    KW_FUNC("func"),
    KW_RETURN("return"),
    KW_BREAK("break"),
    KW_CONTINUE("continue"),
    KW_VAR("var"),

    KW_TYPE_INT("int"),
    KW_TYPE_LONG("long"),
    KW_TYPE_SHORT("short"),
    KW_TYPE_DOUBLE("double"),
    KW_TYPE_FLOAT("float"),
    KW_TYPE_CHAR("char"),
    KW_TYPE_BOOL("bool"),
    KW_TYPE_BYTE("byte"),
    KW_TYPE_VOID("void"),
    KW_TYPE_STR("string");

    private static final Set<TokenKind> keywords = Set.of(
            KW_IF, KW_ELSE, KW_TRUE, KW_FALSE, KW_WHILE, KW_FOR, KW_NULL, KW_FUNC, KW_RETURN, KW_BREAK, KW_CONTINUE, KW_VAR,
            KW_TYPE_INT, KW_TYPE_LONG, KW_TYPE_SHORT, KW_TYPE_DOUBLE, KW_TYPE_FLOAT, KW_TYPE_CHAR, KW_TYPE_BOOL, KW_TYPE_BYTE, KW_TYPE_VOID, KW_TYPE_STR
    );
    private final String name;

    TokenKind(String name) {
        this.name = name;
    }

    public static boolean isKeyword(String lexeme) {
        for (TokenKind kind : keywords) {
            if (kind.name.equals(lexeme)) {
                return true;
            }
        }
        return false;
    }

    public static TokenKind get(String lexeme) {
        for (TokenKind kind : values()) {
            if (kind.name.equals(lexeme)) {
                return kind;
            }
        }
        return TK_IDENT;
    }

    public String toString() {
        return name;
    }

}