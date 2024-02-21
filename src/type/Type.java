package type;

import java.util.HashMap;
import java.util.Map;

public class Type {
    public static final Map<TypeKind, Type> BasicTypes = new HashMap<>() {{
        put(TypeKind.TypeInt, new Type(TypeKind.TypeInt));
        put(TypeKind.TypeLong, new Type(TypeKind.TypeLong));
        put(TypeKind.TypeShort, new Type(TypeKind.TypeShort));
        put(TypeKind.TypeDouble, new Type(TypeKind.TypeDouble));
        put(TypeKind.TypeFloat, new Type(TypeKind.TypeFloat));
        put(TypeKind.TypeChar, new Type(TypeKind.TypeChar));
        put(TypeKind.TypeBool, new Type(TypeKind.TypeBool));
        put(TypeKind.TypeByte, new Type(TypeKind.TypeByte));
        put(TypeKind.TypeVoid, new Type(TypeKind.TypeVoid));
        put(TypeKind.TypeString, new Type(TypeKind.TypeString));
    }};
    private final TypeKind kind;
    private Type elemType;

    public Type(TypeKind kind) {
        this.kind = kind;
    }

    public Type(TypeKind kind, Type elemType) {
        this.kind = kind;
        this.elemType = elemType;
    }

    @Override
    public String toString() {
        switch (this.kind) {
            case TypeInt:
                return "int";
            case TypeLong:
                return "long";
            case TypeShort:
                return "short";
            case TypeDouble:
                return "double";
            case TypeFloat:
                return "float";
            case TypeChar:
                return "char";
            case TypeBool:
                return "boolean";
            case TypeByte:
                return "byte";
            case TypeVoid:
                return "void";
            case TypeString:
                return "String";
            case TypeArray:
                return this.elemType + "[]";
            default:
                throw new UnsupportedOperationException("Unimplemented type");
        }
    }

}


