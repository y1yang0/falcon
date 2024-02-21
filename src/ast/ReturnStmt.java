package ast;

public class ReturnStmt extends AstStmt {
    private AstExpr expr;

    public AstExpr getExpr() {
        return expr;
    }

    public void setExpr(AstExpr expr) {
        this.expr = expr;
    }

    public String toString() {
        return "ReturnStmt";
    }
}

