package ast;

public class AstWalker {
    private final AstNode root;
    private AstFunc funcPre;
    private final AstFunc func;
    private AstFunc funcPost;

    public AstWalker(AstNode root, AstFunc func) {
        this.root = root;
        this.func = func;
    }

    public AstWalker(AstNode root, AstFunc func, AstFunc funcPost) {
        this.root = root;
        this.func = func;
        this.funcPost = funcPost;
    }

    public AstWalker(AstNode root, AstFunc funcPre, AstFunc func, AstFunc funcPost) {
        this.root = root;
        this.funcPre = funcPre;
        this.func = func;
        this.funcPost = funcPost;
    }

    public void walkAst(AstNode node, AstNode prev, int depth) {
        if (node == null) {
            return;
        }
        if (funcPre != null) {
            funcPre.apply(node, prev, depth);
        }
        func.apply(node, prev, depth);
        if (node instanceof BreakStmt || node instanceof ContinueStmt) {
            return;
        }
        if (node instanceof SimpleStmt) {
            walkAst(((SimpleStmt) node).getExpr(), node, depth + 1);
        } else if (node instanceof AssignStmt) {
            walkAst(((AssignStmt) node).getLeft(), node, depth + 1);
            walkAst(((AssignStmt) node).getRight(), node, depth + 1);
        } else if (node instanceof ReturnStmt) {
            walkAst(((ReturnStmt) node).getExpr(), node, depth + 1);
        } else if (node instanceof DefVarStmt) {
            walkAst(((DefVarStmt) node).getName(), node, depth + 1);
            walkAst(((DefVarStmt) node).getInit(), node, depth + 1);
        } else if (node instanceof IfStmt) {
            walkAst(((IfStmt) node).getCond(), node, depth + 1);
            walkAst(((IfStmt) node).getThen(), node, depth + 1);
            walkAst(((IfStmt) node).getElse(), node, depth + 1);
        } else if (node instanceof ForStmt) {
            walkAst(((ForStmt) node).getInit(), node, depth + 1);
            walkAst(((ForStmt) node).getCond(), node, depth + 1);
            walkAst(((ForStmt) node).getPost(), node, depth + 1);
            walkAst(((ForStmt) node).getBody(), node, depth + 1);
        } else if (node instanceof WhileStmt) {
            walkAst(((WhileStmt) node).getCond(), node, depth + 1);
            walkAst(((WhileStmt) node).getBody(), node, depth + 1);
        } else if (node instanceof UnaryExpr) {
            walkAst(((UnaryExpr) node).getLeft(), node, depth + 1);
        } else if (node instanceof BinaryExpr) {
            walkAst(((BinaryExpr) node).getLeft(), node, depth + 1);
            walkAst(((BinaryExpr) node).getRight(), node, depth + 1);
        } else if (node instanceof TernaryExpr) {
            walkAst(((TernaryExpr) node).getCond(), node, depth + 1);
            walkAst(((TernaryExpr) node).getThen(), node, depth + 1);
            walkAst(((TernaryExpr) node).getElse(), node, depth + 1);
        } else if (node instanceof AssignExpr) {
            walkAst(((AssignExpr) node).getLeft(), node, depth + 1);
            walkAst(((AssignExpr) node).getRight(), node, depth + 1);
        } else if (node instanceof IntExpr || node instanceof LongExpr || node instanceof ShortExpr ||
                node instanceof DoubleExpr || node instanceof FloatExpr || node instanceof CharExpr ||
                node instanceof BoolExpr || node instanceof ByteExpr || node instanceof VoidExpr ||
                node instanceof NullExpr || node instanceof StrExpr) {
        } else if (node instanceof ArrayExpr) {
            for (AstExpr elem : ((ArrayExpr) node).getElems()) {
                walkAst(elem, node, depth + 1);
            }
        } else if (node instanceof VarExpr) {
        } else if (node instanceof FuncCallExpr) {
            for (AstExpr elem : ((FuncCallExpr) node).getArgs()) {
                walkAst(elem, node, depth + 1);
            }
        } else if (node instanceof IndexExpr) {
            walkAst(((IndexExpr) node).getIndex(), node, depth + 1);
        } else if (node instanceof RootDecl) {
            for (AstNode elem : ((RootDecl) node).getList()) {
                walkAst(elem, node, depth + 1);
            }
            for (AstDecl elem : ((RootDecl) node).getFunc()) {
                walkAst(elem, node, depth + 1);
            }
        } else if (node instanceof FuncDecl) {
            for (AstExpr elem : ((FuncDecl) node).getParams()) {
                walkAst(elem, node, depth + 1);
            }
            walkAst(((FuncDecl) node).getBlock(), node, depth + 1);
        } else if (node instanceof BlockDecl) {
            for (AstStmt elem : ((BlockDecl) node).getStmts()) {
                walkAst(elem, node, depth + 1);
            }
        } else {
            throw new UnsupportedOperationException();
        }
        if (funcPost != null) {
            funcPost.apply(node, prev, depth);
        }
    }
}