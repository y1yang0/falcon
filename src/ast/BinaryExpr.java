package ast;

public class BinaryExpr extends AstExpr {
    private AstExpr left;
    private AstExpr right;
    private TokenKind opt;

    public AstExpr getLeft() {
        return left;
    }

    public void setLeft(AstExpr left) {
        this.left = left;
    }

    public AstExpr getRight() {
        return right;
    }

    public void setRight(AstExpr right) {
        this.right = right;
    }

    public TokenKind getOpt() {
        return opt;
    }

    public void setOpt(TokenKind opt) {
        this.opt = opt;
    }

    public String toString() {
        return String.format("BinaryExpr{%s}", opt);
    }
}