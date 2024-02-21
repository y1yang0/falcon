package ast;

public class CharExpr extends AstExpr {
    private int value;

    public int getValue() {
        return value;
    }

    public void setValue(int value) {
        this.value = value;
    }

    public String toString() {
        return String.format("CharExpr{%s}", value);
    }
}

