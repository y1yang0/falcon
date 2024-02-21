package ast;

public interface AstFunc {
    void apply(AstNode node, AstNode prev, int depth);
}
