package ast;

public class IfStmt extends AstStmt {
    private AstExpr cond;
    private AstDecl thenB;
    private AstDecl elseB;

    public AstExpr getCond() {
        return cond;
    }

    public void setCond(AstExpr cond) {
        this.cond = cond;
    }

    public AstDecl getThen() {
        return thenB;
    }

    public void setThen(AstDecl then) {
        this.thenB = then;
    }

    public AstDecl getElse() {
        return elseB;
    }

    public void setElse(AstDecl elseB) {
        this.elseB = elseB;
    }

    public String toString() {
        return "IfStmt";
    }
}
