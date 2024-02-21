package ast;

public class AstDotWriter implements AstFunc {
    private final StringBuilder sb;

    public AstDotWriter() {
        sb = new StringBuilder();
        sb.append("digraph G {\n");
        sb.append("  graph [ dpi = 500 ];\n");
    }

    public void apply(AstNode node, AstNode prev, int depth) {
        if (node == null) {
            return;
        }
        String edge = String.format("  %s_%s -> %s_%s\n",
                prev.getClass().getSimpleName(),
                prev.hashCode(),
                node.getClass().getSimpleName(),
                node.hashCode());
        String n = String.format("  %s_%s [label=\"%s\"]\n",
                node.getClass().getSimpleName(),
                node.hashCode(),
                node);
        System.out.println("===" + edge);
        sb.append(edge);
        sb.append(n + "\n");
    }

    public String toString() {
        sb.append("}\n");
        return sb.toString();
    }
}
