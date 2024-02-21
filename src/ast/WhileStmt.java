package ast;

public class WhileStmt extends AstStmt {
    private AstExpr cond;
    private AstDecl body;

    public AstExpr getCond() {
        return cond;
    }

    public void setCond(AstExpr cond) {
        this.cond = cond;
    }

    public AstDecl getBody() {
        return body;
    }

    public void setBody(AstDecl body) {
        this.body = body;
    }

    public String toString() {
        return "WhileStmt";
    }
}

