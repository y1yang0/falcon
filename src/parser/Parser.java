package parser;

import ast.*;
import type.Type;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;

import static type.Type.BasicTypes;
import static type.TypeKind.*;

public class Parser {
    private TokenKind token;
    private String lexeme;
    private Lexer lexer;

    private void syntaxError(String fmt, Object... args) {
        System.err.printf(fmt, args);
        System.exit(1);
    }

    private void guarantee(boolean cond, String fmt, Object... args) {
        if (!cond) {
            syntaxError("SyntaxError: " + fmt, args);
        }
    }

    private void consume() {
        Token token = lexer.nextToken();
        this.token = token.getKind();
        this.lexeme = token.getLexeme();
    }

    private List<AstExpr> parseParams() {
        List<AstExpr> params = new ArrayList<>();
        consume();

        if (token == TokenKind.TK_RPAREN) {
            consume();
            return params;
        }

        while (token != TokenKind.TK_RPAREN) {
            if (token == TokenKind.TK_IDENT) {
                String paramName = lexeme;
                consume();
                Type paramType = parseType();
                VarExpr param = new VarExpr(paramName);
                param.setType(paramType);
                params.add(param);
            } else {
                guarantee(token == TokenKind.TK_COMMA, "Expected ,");
                consume();
            }
        }
        guarantee(token == TokenKind.TK_RPAREN, "Expected ')'");
        consume();
        return params;
    }

    private FuncDecl parseFuncDecl() {
        guarantee(token == TokenKind.KW_FUNC, "Expected function definition");
        consume();
        FuncDecl fn = new FuncDecl();
        fn.setName(lexeme);
        consume();
        guarantee(token == TokenKind.TK_LPAREN, "Expected '('");
        fn.setParams(parseParams());
        Type retType = parseType();
        if (retType != null) {
            fn.setRetType(retType);
        } else {
            fn.setRetType(BasicTypes.get(TypeVoid));
        }
        if (token == TokenKind.TK_LBRACE) {
            fn.setBlock(parseBlockDecl("body"));
        } else {
            final String RUNTIME_PREFIX = "rt_";
            fn.setBuiltin(true);
            fn.setBlock(new BlockDecl("nativeBody", List.of(new SimpleStmt(new FuncCallExpr(RUNTIME_PREFIX + fn.getName(), fn.getParams())))));
        }
        return fn;
    }

    private BlockDecl parseBlockDecl(String name) {
        guarantee(token == TokenKind.TK_LBRACE, "Expected '{'");
        consume();
        BlockDecl elem = new BlockDecl(name);
        elem.setStmts(parseStatementList());
        guarantee(token == TokenKind.TK_RBRACE, "Expected '}'");
        consume();
        return elem;
    }

    private List<AstStmt> parseStatementList() {
        List<AstStmt> stmts = new ArrayList<>();
        AstStmt elem;
        while ((elem = parseStatement()) != null) {
            stmts.add(elem);
        }
        return stmts;
    }

    private AstStmt parseStatement() {
        switch (token) {
            case KW_RETURN:
                return parseReturnStmt();
            case KW_VAR:
                return parseDefVarStmt();
            case KW_IF:
                return parseIfStmt();
            case KW_FOR:
                return parseForStmt();
            case KW_WHILE:
                return parseWhileStmt();
            case KW_BREAK:
                return parseBreakStmt();
            case KW_CONTINUE:
                return parseContinueStmt();
            default:
                return parseSimpleStmt();
        }
    }

    private AstStmt parseForStmt() {
        guarantee(token == TokenKind.KW_FOR, "Expected for");
        consume();
        ForStmt elem = new ForStmt();
        AstExpr expr = parseExpression();
        elem.setInit(expr);
        consume();
        elem.setCond(parseExpression());
        guarantee(token == TokenKind.TK_SEMICOLON, "Expected ;");
        consume();
        elem.setPost(parseExpression());
        elem.setBody(parseBlockDecl("forBody"));
        return elem;
    }

    private AstStmt parseWhileStmt() {
        guarantee(token == TokenKind.KW_WHILE, "Expected while");
        consume();
        WhileStmt elem = new WhileStmt();
        elem.setCond(parseExpression());
        elem.setBody(parseBlockDecl("whileBody"));
        return elem;
    }

    private AstStmt parseReturnStmt() {
        guarantee(token == TokenKind.KW_RETURN, "Expected return");
        consume();
        ReturnStmt elem = new ReturnStmt();
        elem.setExpr(parseExpression());
        return elem;
    }

    private AstStmt parseSimpleStmt() {
        AstExpr elem = parseExpression();
        if (elem != null) {
            return new SimpleStmt(elem);
        }
        return null;
    }

    private AstStmt parseDefVarStmt() {
        guarantee(token == TokenKind.KW_VAR, "Expected var");
        consume();
        DefVarStmt elem = new DefVarStmt();
        elem.setName((VarExpr) parsePrimaryExpr());
        if (token == TokenKind.TK_ASSIGN) {
            guarantee(token == TokenKind.TK_ASSIGN, "Expected =");
            consume();
            elem.setInit(parseExpression());
        }
        return elem;
    }

    private AstStmt parseIfStmt() {
        guarantee(token == TokenKind.KW_IF, "Expected if");
        consume();
        IfStmt elem = new IfStmt();
        elem.setCond(parseExpression());
        elem.setThen(parseBlockDecl("then"));
        if (token == TokenKind.KW_ELSE) {
            consume();
            if (token == TokenKind.KW_IF) {
                BlockDecl b = new BlockDecl("else");
                b.setStmts(List.of(parseIfStmt()));
                elem.setElse(b);
            } else if (token == TokenKind.TK_LBRACE) {
                elem.setElse(parseBlockDecl("else"));
            } else {
                syntaxError("Expected else or {");
            }
        }
        return elem;
    }

    private AstStmt parseBreakStmt() {
        guarantee(token == TokenKind.KW_BREAK, "Expected break");
        consume();
        return new BreakStmt();
    }

    private AstStmt parseContinueStmt() {
        guarantee(token == TokenKind.KW_CONTINUE, "Expected continue");
        consume();
        return new ContinueStmt();
    }

    private AstExpr parsePrimaryExpr() {
        switch (token) {
            case TK_IDENT: {
                String ident = lexeme;
                consume();
                switch (token) {
                    case TK_LPAREN: {
                        consume();
                        FuncCallExpr val = new FuncCallExpr(ident);
                        while (token != TokenKind.TK_RPAREN) {
                            val.getArgs().add(parseExpression());
                            if (token == TokenKind.TK_COMMA) {
                                consume();
                            }
                        }
                        guarantee(token == TokenKind.TK_RPAREN, "Expected )");
                        consume();
                        return val;
                    }
                    case TK_LBRACKET: {
                        consume();
                        IndexExpr val = new IndexExpr();
                        val.setName(ident);
                        val.setIndex(parseExpression());
                        val.setType(null);
                        guarantee(token == TokenKind.TK_RBRACKET, "Expected ']'");
                        consume();
                        return val;
                    }
                    default: {
                        VarExpr val = new VarExpr(ident);
                        val.setType(parseType());
                        return val;
                    }
                }
            }
            case TK_LBRACKET:{
                consume();
                ArrayExpr val = new ArrayExpr();
                while (token != TokenKind.TK_RBRACKET) {
                    val.getElems().add(parseExpression());
                    if (token == TokenKind.TK_COMMA) {
                        consume();
                    }
                }
                val.setType(new Type(TypeArray, val.getElems().get(0).getType()));
                guarantee(token == TokenKind.TK_RBRACKET, "Expected ']'");
                consume();
                return val;}
            case KW_FUNC:
                break;
            case LIT_INT:{
                IntExpr elem = new IntExpr();
                elem.setType(BasicTypes.get(TypeInt));
                elem.setValue(Integer.parseInt(lexeme));
                consume();
                return elem;}
            case LIT_LONG:
            case LIT_SHORT:
            case LIT_BOOL:
            case LIT_FLOAT:
            case LIT_BYTE:
                break;
            case LIT_DOUBLE:{
                DoubleExpr elem = new DoubleExpr();
                elem.setType(BasicTypes.get(TypeDouble));
                elem.setValue(Double.parseDouble(lexeme));
                consume();
                return elem;}
            case LIT_STR:{
                StrExpr elem = new StrExpr();
                elem.setType(BasicTypes.get(TypeString));
                elem.setValue(lexeme);
                consume();
                return elem;}
            case LIT_CHAR:{
                CharExpr elem = new CharExpr();
                elem.setType(BasicTypes.get(TypeChar));
                elem.setValue(lexeme.charAt(0));
                consume();
                return elem;}
            case KW_TRUE:{
                BoolExpr elem = new BoolExpr();
                elem.setType(BasicTypes.get(TypeBool));
                elem.setValue(true);
                consume();
                return elem;}
            case KW_FALSE:{
                BoolExpr elem = new BoolExpr();
                elem.setType(BasicTypes.get(TypeBool));
                elem.setValue(false);
                consume();
                return elem;}
            case KW_NULL:
                return new NullExpr();
            case TK_LPAREN:
                consume();
                AstExpr expr = parseExpression();
                guarantee(token == TokenKind.TK_RPAREN, "Expected )");
                consume();
                return expr;
        }
        return null;
    }

    private AstExpr parseUnaryExpr() {
        if (token == TokenKind.TK_MINUS || token == TokenKind.TK_LOGNOT || token == TokenKind.TK_BITNOT) {
            UnaryExpr val = new UnaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(parseUnaryExpr());
            return val;
        } else if (token == TokenKind.LIT_DOUBLE || token == TokenKind.LIT_INT || token == TokenKind.LIT_STR ||
                token == TokenKind.LIT_CHAR || token == TokenKind.TK_IDENT || token == TokenKind.TK_LPAREN ||
                token == TokenKind.TK_LBRACKET || token == TokenKind.KW_TRUE || token == TokenKind.KW_FALSE ||
                token == TokenKind.KW_NULL || token == TokenKind.KW_FUNC) {
            AstExpr prim = parsePrimaryExpr();
            if (token == TokenKind.TK_DOT) {

            } else {
                return prim;
            }
        }
        return null;
    }

    private AstExpr parseMulExpr() {
        AstExpr left = parseUnaryExpr();
        while (token == TokenKind.TK_TIMES || token == TokenKind.TK_DIV || token == TokenKind.TK_MOD) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseUnaryExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseAddExpr() {
        AstExpr left = parseMulExpr();
        while (token == TokenKind.TK_PLUS || token == TokenKind.TK_MINUS) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseMulExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseBitshiftExpr() {
        AstExpr left = parseAddExpr();
        while (token == TokenKind.TK_LSHIFT || token == TokenKind.TK_RSHIFT) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseAddExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseRelationalExpr() {
        AstExpr left = parseBitshiftExpr();
        while (token == TokenKind.TK_GT || token == TokenKind.TK_GE || token == TokenKind.TK_LT || token == TokenKind.TK_LE) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseBitshiftExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseEqualityExpr() {
        AstExpr left = parseRelationalExpr();
        while (token == TokenKind.TK_EQ || token == TokenKind.TK_NE) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseRelationalExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseBitandExpr() {
        AstExpr left = parseEqualityExpr();
        while (token == TokenKind.TK_BITAND) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseEqualityExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseBitxorExpr() {
        AstExpr left = parseBitandExpr();
        while (token == TokenKind.TK_BITXOR) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseBitandExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseBitorExpr() {
        AstExpr left = parseBitxorExpr();
        while (token == TokenKind.TK_BITOR) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseBitxorExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseLogicalAndExpr() {
        AstExpr left = parseBitorExpr();
        while (token == TokenKind.TK_LOGAND) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseBitorExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseLogicalOrExpr() {
        AstExpr left = parseLogicalAndExpr();
        while (token == TokenKind.TK_LOGOR) {
            BinaryExpr val = new BinaryExpr();
            val.setOpt(token);
            consume();
            val.setLeft(left);
            val.setRight(parseLogicalAndExpr());
            left = val;
        }
        return left;
    }

    private AstExpr parseTenaryExpr() {
        AstExpr left = parseLogicalOrExpr();
        if (token == TokenKind.TK_QUESTION) {
            TernaryExpr val = new TernaryExpr();
            val.setCond(left);
            consume();
            val.setThen(parseExpression());
            guarantee(token == TokenKind.TK_COLON, "Expected :");
            consume();
            val.setElse(parseTenaryExpr());
            return val;
        }
        return left;
    }

    private AstExpr parseExpression() {
        AstExpr left = parseTenaryExpr();
        if (token == TokenKind.TK_ASSIGN || token == TokenKind.TK_PLUS_AGN || token == TokenKind.TK_MINUS_AGN ||
                token == TokenKind.TK_TIMES_AGN || token == TokenKind.TK_DIV_AGN || token == TokenKind.TK_MOD_AGN ||
                token == TokenKind.TK_BITAND_AGN || token == TokenKind.TK_BITOR_AGN || token == TokenKind.TK_BITXOR_AGN) {
            AssignExpr val = new AssignExpr();
            val.setOpt(token);
            val.setLeft(left);
            consume();
            val.setRight(parseExpression());
            return val;
        }
        return left;
    }

    private Type parseType() {
        switch (token) {
            case KW_TYPE_INT:
                consume();
                return BasicTypes.get(TypeInt);
            case KW_TYPE_LONG:
                consume();
                return BasicTypes.get(TypeLong);
            case KW_TYPE_SHORT:
                consume();
                return BasicTypes.get(TypeShort);
            case KW_TYPE_CHAR:
                consume();
                return BasicTypes.get(TypeChar);
            case KW_TYPE_FLOAT:
                consume();
                return BasicTypes.get(TypeFloat);
            case KW_TYPE_DOUBLE:
                consume();
                return BasicTypes.get(TypeDouble);
            case KW_TYPE_STR:
                consume();
                return BasicTypes.get(TypeString);
            case KW_TYPE_BYTE:
                consume();
                return BasicTypes.get(TypeByte);
            case KW_TYPE_BOOL:
                consume();
                return BasicTypes.get(TypeBool);
            case TK_LBRACKET:
                consume();
                guarantee(token == TokenKind.TK_RBRACKET, "Expected ']'");
                consume();
                return new Type(TypeArray, parseType());
        }
        return null;
    }

    public void init(File file) {
        lexer = new Lexer(file.getPath());
    }

    public AstNode parse() {
        RootDecl root = new RootDecl();
        root.setSource(lexer.getFileName());
        consume();
        if (token == TokenKind.TK_EOF) {
            return root;
        }
        while (token != TokenKind.TK_EOF) {
            if (token == TokenKind.KW_FUNC) {
                root.getFunc().add(parseFuncDecl());
            } else {
                AstStmt stmt = parseStatement();
                root.getList().add(stmt);
            }
        }
        return root;
    }

    public static RootDecl parseFile(String fileName) {
        File file = new File(fileName);
        Parser parser = new Parser();
        parser.init(file);
        return (RootDecl) parser.parse();
    }

    public static RootDecl parseText(String text) {
        try {
            File tmpFile = File.createTempFile("falcon", null);
            Files.write(Paths.get(tmpFile.getPath()), text.getBytes());
            Parser parser = new Parser();
            parser.init(tmpFile);
            return (RootDecl) parser.parse();
        } catch (IOException e) {
            e.printStackTrace();
        }
        return null;
    }
}


