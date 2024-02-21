package ast;

public class ForStmt extends AstStmt {
    private AstExpr init;
    private AstExpr cond;
    private AstExpr post;
    private AstDecl body;


    public void setInit(AstExpr init) {
        this.init = init;
    }

    public AstExpr getCond() {
        return cond;
    }

    public void setCond(AstExpr cond) {
        this.cond = cond;
    }

    public AstExpr getPost() {
        return post;
    }

    public void setPost(AstExpr post) {
        this.post = post;
    }

    public AstDecl getBody() {
        return body;
    }

    public void setBody(AstDecl body) {
        this.body = body;
    }

    public String toString() {
        return "ForStmt";
    }
}

