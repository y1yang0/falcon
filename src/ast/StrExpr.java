package ast;

public class StrExpr extends AstExpr {
    private String value;

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    public String toString() {
        return String.format("StrExpr{%s}", value);
    }
}