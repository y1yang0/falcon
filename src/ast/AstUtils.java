package ast;

public class AstUtils {
    public void printAst(RootDecl root, boolean showTypes) {
        AstPrinter printer = new AstPrinter();
        AstWalker walker = new AstWalker(root, printer);
        walker.walkAst(root, root, 0);
    }

    public void dumpAstToDotFile(String name, RootDecl root) {
        AstDotWriter writer = new AstDotWriter();
        AstWalker walker = new AstWalker(root, writer);
        walker.walkAst(root, root, 0);
        String dotFile = String.format("ast_%s.dot", name);
        String pngFile = String.format("ast_%s.png", name);
        try {
            java.io.FileWriter fw = new java.io.FileWriter(dotFile);
            fw.write(writer.toString());
            fw.close();
            Runtime.getRuntime().exec(String.format("dot -Tpng %s -o %s", dotFile, pngFile));
        } catch (java.io.IOException e) {
            e.printStackTrace();
        }
    }
}