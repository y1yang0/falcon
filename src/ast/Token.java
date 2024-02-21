package ast;

public class Token {
    private final TokenKind kind;
    private final String lexeme;

    public Token(TokenKind kind, String lexeme) {
        this.kind = kind;
        this.lexeme = lexeme;
    }

    public TokenKind getKind() {
        return kind;
    }

    public String getLexeme() {
        return lexeme;
    }
}
