package ast;

public class VarExpr extends AstExpr {
    private String name;

    public VarExpr(String paramName) {
        this.name = paramName;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String toString() {
        return String.format("VarExpr{%s}", name);
    }
}
