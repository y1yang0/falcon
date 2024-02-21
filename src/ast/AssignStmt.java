package ast;

public class AssignStmt extends AstStmt {
    private AstExpr left;
    private AstExpr right;

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

    public String toString() {
        return "AssignStmt";
    }
}

