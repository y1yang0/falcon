package ast;

import type.Type;

import java.util.List;

public class FuncDecl extends AstDecl {
    private String name;
    private List<AstExpr> params;
    private AstDecl block;
    private Type retType;
    private boolean builtin;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public List<AstExpr> getParams() {
        return params;
    }

    public void setParams(List<AstExpr> params) {
        this.params = params;
    }

    public AstDecl getBlock() {
        return block;
    }

    public void setBlock(AstDecl block) {
        this.block = block;
    }

    public Type getRetType() {
        return retType;
    }

    public void setRetType(Type retType) {
        this.retType = retType;
    }

    public boolean isBuiltin() {
        return builtin;
    }

    public void setBuiltin(boolean builtin) {
        this.builtin = builtin;
    }

    public String toString() {
        if (builtin) {
            return String.format("FuncDecl{%s@builtin}", name);
        } else {
            return String.format("FuncDecl{%s}", name);
        }
    }
}