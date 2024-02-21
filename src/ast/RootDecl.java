package ast;

import java.util.List;

public class RootDecl extends AstDecl {
    private String source;
    private List<AstDecl> func;
    private List<AstNode> list;

    public String getSource() {
        return source;
    }

    public void setSource(String source) {
        this.source = source;
    }

    public List<AstDecl> getFunc() {
        return func;
    }

    public void setFunc(List<AstDecl> func) {
        this.func = func;
    }

    public List<AstNode> getList() {
        return list;
    }

    public void setList(List<AstNode> list) {
        this.list = list;
    }

    public String toString() {
        return "RootDecl";
    }
}
