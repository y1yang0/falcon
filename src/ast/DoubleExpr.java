package ast;

public class DoubleExpr extends AstExpr {
    private double value;

    public double getValue() {
        return value;
    }

    public void setValue(double value) {
        this.value = value;
    }

    public String toString() {
        return String.format("DoubleExpr{%s}", value);
    }
}

