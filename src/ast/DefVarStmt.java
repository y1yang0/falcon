package ast;

public class DefVarStmt extends AstStmt {
    private AstExpr name;
    private AstExpr init;

    public AstExpr getName() {
        return name;
    }

    public void setName(AstExpr name) {
        this.name = name;
    }

    public AstExpr getInit() {
        return init;
    }

    public void setInit(AstExpr init) {
        this.init = init;
    }

    public String toString() {
        return "DefVarStmt";
    }
}
