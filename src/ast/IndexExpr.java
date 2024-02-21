package ast;

public class IndexExpr extends AstExpr {
    private AstExpr index;
    private String name;

    public AstExpr getIndex() {
        return index;
    }

    public void setIndex(AstExpr index) {
        this.index = index;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String toString() {
        return String.format("IndexExpr{%s}", name);
    }
}
