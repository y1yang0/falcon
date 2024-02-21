package ast;

public class ByteExpr extends AstExpr {
    private byte value;

    public byte getValue() {
        return value;
    }

    public void setValue(byte value) {
        this.value = value;
    }

    public String toString() {
        return String.format("ByteExpr{%s}", value);
    }
}

