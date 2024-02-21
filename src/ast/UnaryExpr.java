package ast;

public class UnaryExpr extends AstExpr {
    private AstExpr left;
    private TokenKind opt;

    public AstExpr getLeft() {
        return left;
    }

    public void setLeft(AstExpr left) {
        this.left = left;
    }

    public TokenKind getOpt() {
        return opt;
    }

    public void setOpt(TokenKind opt) {
        this.opt = opt;
    }

    public String toString() {
        return String.format("UnaryExpr{%s}", opt);
    }
}