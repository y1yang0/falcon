package ast;

public class BoolExpr extends AstExpr {
    private boolean value;

    public boolean getValue() {
        return value;
    }

    public void setValue(boolean value) {
        this.value = value;
    }

    public String toString() {
        return String.format("BoolExpr{%s}", value);
    }
}