package ast;

import java.util.List;

public class ArrayExpr extends AstExpr {
    private List<AstExpr> elems;

    public List<AstExpr> getElems() {
        return elems;
    }

    public void setElems(List<AstExpr> elems) {
        this.elems = elems;
    }

    public String toString() {
        return "ArrayExpr";
    }
}
