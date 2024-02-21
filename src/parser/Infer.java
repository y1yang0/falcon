package parser;

import ast.*;
import type.Type;

import java.util.HashMap;
import java.util.Map;
import java.util.Stack;

public class Infer {
    private Stack<Map<String, Type>> varScopes;
    private Map<String, Type> funcScopes;

    public Infer() {
        varScopes = new Stack<>();
        funcScopes = new HashMap<>();
    }

    public int numScopes() {
        return varScopes.size();
    }

    public Map<String, Type> enterScope() {
        Map<String, Type> names = new HashMap<>();
        varScopes.push(names);
        return names;
    }

    public void exitScope() {
        varScopes.pop();
    }

    public Type getVarType(String name) {
        for (int i = varScopes.size() - 1; i >= 0; i--) {
            if (varScopes.get(i) instanceof Map) {
                Map<String, Type> names = (Map<String, Type>) varScopes.get(i);
                if (names.containsKey(name)) {
                    return names.get(name);
                }
            }
        }
        return null;
    }

    public void setVarType(String name, Type t) {
        for (int i = varScopes.size() - 1; i >= 0; i--) {
            if (varScopes.get(i) instanceof Map) {
                Map<String, Type> names = (Map<String, Type>) varScopes.get(i);
                if (names.containsKey(name)) {
                    names.put(name, t);
                    return;
                }
            }
        }
        Map<String, Type> names = varScopes.peek();
        names.put(name, t);
    }

    public Object inferPre(AstNode node, AstNode _, int depth) {
        if (node instanceof FuncDecl) {
            Map<String, Type> names = enterScope();
            FuncDecl funcDecl = (FuncDecl) node;
            for (AstNode param : funcDecl.Params) {
                names.put(((VarExpr) param).Name, param.getType());
            }
        } else if (node instanceof BlockDecl) {
            enterScope();
        } else if (node instanceof ForStmt || node instanceof IfStmt || node instanceof WhileStmt) {
            enterScope();
        }
        return null;
    }

    public Object inferPost(AstNode node, AstNode _, int depth) {
        if (node instanceof FuncDecl) {
            exitScope();
        } else if (node instanceof BlockDecl) {
            exitScope();
        } else if (node instanceof ForStmt || node instanceof IfStmt || node instanceof WhileStmt) {
            exitScope();
        }
        return null;
    }

    public Type resolveType(TokenKind opt, Object left, Object right) {
        Type lt = null;
        Type rt = null;
        if (left != null) {
            lt = (Type) left;
        }
        if (right != null) {
            rt = (Type) right;
        }
        if (lt == BasicTypes[TypeDouble] || rt == BasicTypes[TypeDouble]) {
            return BasicTypes[TypeDouble];
        }
        return rt;
    }

    public Object infer(AstNode node, AstNode _, int depth) {
        if (node instanceof IntExpr || node instanceof LongExpr || node instanceof DoubleExpr
                || node instanceof FloatExpr || node instanceof CharExpr
                || node instanceof BoolExpr || node instanceof ByteExpr
                || node instanceof StrExpr || node instanceof ArrayExpr) {
            AstExpr e = (AstExpr) node;
            if (e.getType() == null) {
                syntaxError("literal must hold the type info");
            }
            return e.getType();
        } else if (node instanceof IndexExpr) {
            IndexExpr e = (IndexExpr) node;
            Type arrType = getVarType(e.Name);
            Type elemType = arrType.ElemType;
            e.setType(elemType);
            return elemType;
        } else if (node instanceof TernaryExpr) {
            TernaryExpr e = (TernaryExpr) node;
            AstNode thenExpr = e.Then;
            Type thenType = (Type) infer(thenExpr, e, depth + 1);
            e.setType(thenType);
        } else if (node instanceof FuncCallExpr) {
            FuncCallExpr e = (FuncCallExpr) node;
            Type retType = funcScopes.get(e.Name);
            e.setType(retType);
            return retType;
        } else if (node instanceof UnaryExpr) {
            UnaryExpr e = (UnaryExpr) node;
            Object leftType = infer(e.Left, e, depth + 1);
            if (leftType != null) {
                e.setType((Type) leftType);
            }
        } else if (node instanceof BinaryExpr) {
            BinaryExpr e = (BinaryExpr) node;
            if (e.Opt.IsCmpOp() || e.Opt.IsShortCircuitOp()) {
                e.setType(BasicTypes[TypeBool]);
                return BasicTypes[TypeBool];
            }
            Object leftType = infer(e.Left, e, depth + 1);
            Object rightType = infer(e.Right, e, depth + 1);
            Type finalType = resolveType(e.Opt, leftType, rightType);
            if (finalType != null) {
                e.setType(finalType);
            }
            return rightType;
        } else if (node instanceof AssignExpr) {
            AssignExpr e = (AssignExpr) node;
            Object rightType = infer(e.Right, e, depth + 1);
            if (rightType != null) {
                e.setType((Type) rightType);
                AstExpr left = e.Left;
                if (left instanceof VarExpr) {
                    VarExpr v = (VarExpr) left;
                    setVarType(v.Name, (Type) rightType);
                } else if (left instanceof IndexExpr) {

                } else {
                    Utils.Unimplement();
                }
            }
            return rightType;
        } else if (node instanceof VarExpr) {
            VarExpr v = (VarExpr) node;
            if (v.getType() != null) {
                setVarType(v.Name, v.getType());
                return v.getType();
            }
            Type vt = getVarType(v.Name);
            if (vt != null) {
                setVarType(v.Name, vt);
                v.setType(vt);
            }
            return vt;
        } else if (node instanceof DefVarStmt) {
            DefVarStmt s = (DefVarStmt) node;
            Object right = infer(s.Init, s, depth + 1);
            if (right != null) {
                if (s.Name instanceof VarExpr) {
                    String varName = ((VarExpr) s.Name).Name;
                    setVarType(varName, (Type) right);
                } else if (s.Name instanceof IndexExpr) {

                }
            }
        }
        return null;
    }

    public void verifyTypedAst(AstDecl root) {
        AstWalker walker = new AstWalker(root, this::verifyTyped, null, null);
        walker.WalkAst(root, root, 0);
    }

    public void InferTypes(boolean debug, RootDecl... roots) {
        Infer infer = new Infer();

        infer.funcScopes = new HashMap<>();
        for (RootDecl root : roots) {
            for (AstDecl funcDecl : root.Func) {
                FuncDecl funcDecl = (FuncDecl) funcDecl;
                infer.funcScopes.put(funcDecl.Name, funcDecl.RetType);
                if (funcDecl.Builtin) {
                    FuncCallExpr call = (FuncCallExpr) ((BlockDecl) funcDecl.Block).Stmts.get(0).Expr;
                    String callName = call.Name;
                    infer.funcScopes.put(callName, funcDecl.RetType);
                }
            }
        }

        for (RootDecl root : roots) {
            infer.varScopes = new Stack<>();
            AstWalker walker = new AstWalker(root, infer::infer, infer::inferPre, infer::inferPost);
            walker.WalkAst(root, root, 0);
            if (infer.numScopes() != 0) {
                syntaxError("scope is unbalanced after type inference");
            }
            if (debug) {
                System.out.printf("== Type Inference(%s) ==\n", root.Source);
                PrintAst(root, true);
            }

            verifyTypedAst(root);
        }
    }
}

