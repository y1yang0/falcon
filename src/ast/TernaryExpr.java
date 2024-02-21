package ast;

public class TernaryExpr extends AstExpr {
    private AstExpr cond;
    private AstExpr thenB;
    private AstExpr elseB;

    public AstExpr getCond() {
        return cond;
    }

    public void setCond(AstExpr cond) {
        this.cond = cond;
    }

    public AstExpr getThen() {
        return thenB;
    }

    public void setThen(AstExpr then) {
        this.thenB = then;
    }

    public AstExpr getElse() {
        return elseB;
    }

    public void setElse(AstExpr elseB) {
        this.elseB = elseB;
    }

    public String toString() {
        return "TernaryExpr";
    }
}
