package ast;

public class ShortExpr extends AstExpr {
    private short value;

    public short getValue() {
        return value;
    }

    public void setValue(short value) {
        this.value = value;
    }

    public String toString() {
        return String.format("ShortExpr{%s}", value);
    }
}


