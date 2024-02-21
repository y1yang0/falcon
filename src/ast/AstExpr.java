package ast;

import type.Type;

public abstract class AstExpr implements AstNode {
    private Type type;

    public Type getType() {
        return type;
    }

    public void setType(Type type) {
        this.type = type;
    }
}