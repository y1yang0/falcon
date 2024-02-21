package ast;

public class SimpleStmt extends AstStmt {
    private AstExpr expr;

    public SimpleStmt(AstExpr expr) {
        this.expr = expr;
    }

    public AstExpr getExpr() {
        return expr;
    }

    public void setExpr(AstExpr expr) {
        this.expr = expr;
    }

    public String toString() {
        return "SimpleStmt";
    }
}

