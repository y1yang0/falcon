package ast;

import java.util.List;

public class FuncCallExpr extends AstExpr {
    private String name;
    private List<AstExpr> args;

    public FuncCallExpr(String s, List<AstExpr> params) {
        this.name = s;
        this.args = params;
    }

    public FuncCallExpr(String ident) {
        this.name = ident;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public List<AstExpr> getArgs() {
        return args;
    }

    public void setArgs(List<AstExpr> args) {
        this.args = args;
    }

    public String toString() {
        return String.format("FuncCallExpr{%s}", name);
    }
}


