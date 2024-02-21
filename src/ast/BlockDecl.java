package ast;

import java.util.List;

public class BlockDecl extends AstDecl {
    private String name;
    private List<AstStmt> stmts;

    public BlockDecl(String name) {
        this.name = name;
    }

    public BlockDecl(String nativeBody, List<AstStmt> simpleStmts) {
        this.name = nativeBody;
        this.stmts = simpleStmts;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public List<AstStmt> getStmts() {
        return stmts;
    }

    public void setStmts(List<AstStmt> stmts) {
        this.stmts = stmts;
    }

    public String toString() {
        return String.format("BlockDecl{%s}", name);
    }
}
