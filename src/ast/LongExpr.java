package ast;

public class LongExpr extends AstExpr {
    private long value;

    public long getValue() {
        return value;
    }

    public void setValue(long value) {
        this.value = value;
    }

    public String toString() {
        return String.format("LongExpr{%s}", value);
    }
}


