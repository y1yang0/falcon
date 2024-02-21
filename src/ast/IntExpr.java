package ast;

public class IntExpr extends AstExpr {
    private int value;

    public int getValue() {
        return value;
    }

    public void setValue(int value) {
        this.value = value;
    }

    public String toString() {
        return String.format("IntExpr{%s}", value);
    }
}

