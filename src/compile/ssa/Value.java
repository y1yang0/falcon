package compile.ssa;

import type.Type;

import java.util.ArrayList;
import java.util.List;

public class Value {
    private static int globalValueId = 0;
    private final int id;
    private final Op op;
    private final List<Value> args;
    private final Object sym;
    private Block block;
    private final List<Value> uses;
    private final List<Block> useBlock;
    private Type type;

    public Value(Op op) {
        this.id = globalValueId++;
        this.op = op;
        this.args = new ArrayList<>();
        this.sym = null;
        this.block = null;
        this.uses = new ArrayList<>();
        this.useBlock = new ArrayList<>();
        this.type = null;
    }

    public void setBlock(Block block) {
        this.block = block;
    }

    public void setType(Type type) {
        this.type = type;
    }

    public List<Value> args() {
        return args;
    }

    public int id() {
        return id;
    }

    public void addArg(Value... args) {
        for (Value arg : args) {
            this.args.add(arg);
            arg.uses.add(this);
        }
    }

    public void addUseBlock(Block block) {
        this.useBlock.add(block);
        block.setCtrl(this);
    }

    public void removeUseBlock(Block block) {
        this.useBlock.remove(block);
        block.setCtrl(null);
    }

    public void removeUse(Value value) {
        this.uses.remove(value);
    }

    public void replaceUses(Value value) {
        for (int i = 0; i < this.uses.size(); i++) {
            Value use = this.uses.get(i);
            for (int j = 0; j < use.args.size(); j++) {
                if (use.args.get(j) == this) {
                    use.args.set(j, value);
                    this.uses.set(i, null);
                    value.uses.add(use);
                    break;
                }
            }
        }
        this.uses.removeIf(u -> u == null);
        if (!this.uses.isEmpty()) {
            value.useBlock.addAll(this.useBlock);
            for (Block ub : value.useBlock) {
                ub.setCtrl(value);
            }
            this.useBlock.clear();
        }
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        sb.append("v").append(this.id).append(" = ").append(this.op);
        if (this.type != null) {
            sb.append("<").append(this.type).append(">");
        }
        for (Value arg : this.args) {
            sb.append(" v").append(arg.id);
        }
        if (this.sym != null) {
            sb.append(" @").append(this.sym);
        }
        return sb.toString();
    }
}
