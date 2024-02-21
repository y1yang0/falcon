package ast;

public class FloatExpr extends AstExpr {
    private float value;

    public float getValue() {
        return value;
    }

    public void setValue(float value) {
        this.value = value;
    }

    public String toString() {
        return String.format("FloatExpr{%s}", value);
    }
}
