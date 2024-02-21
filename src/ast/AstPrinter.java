package ast;

import type.Type;

public class AstPrinter implements AstFunc {
    public void apply(AstNode node, AstNode prev, int depth) {
        if (node == null) {
            return;
        }
        for (int i = 0; i < depth; i++) {
            System.out.print("..");
        }
        String str = node.toString();
        if (node instanceof AstExpr) {
            Type type = ((AstExpr) node).getType();
            if (type != null) {
                str += String.format(" :: %s", type);
            }
        }
        System.out.println(str);
    }
}