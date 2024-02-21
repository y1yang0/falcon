package compile.ssa;

import type.Type;

import java.util.ArrayList;
import java.util.List;

public class Block {
    private static int globalBlockId = 0;
    private final Func func;
    private final int id;
    private final BlockKind kind;
    private final List<Value> values;
    private final List<Block> succs;
    private final List<Block> preds;
    private Value ctrl;
    private final BlockHint hint;

    public Block(Func func, BlockKind kind) {
        this.func = func;
        this.id = globalBlockId++;
        this.kind = kind;
        this.values = new ArrayList<>();
        this.succs = new ArrayList<>();
        this.preds = new ArrayList<>();
        this.ctrl = null;
        this.hint = BlockHint.None;
    }

    public List<Value> values() {
        return values;
    }

    public void wireTo(Block to) {
        this.succs.add(to);
        to.preds.add(this);
    }

    public Value newValue(Op op, Type type, Value... args) {
        Value val = new Value(op);
        val.setBlock(this);
        val.setType(type);
        for (Value arg : args) {
            val.addArg(arg);
        }
        if (op == Op.Phi) {
            this.values.add(0, val);
        } else {
            this.values.add(val);
        }
        return val;
    }

    public void removeValue(Value val) {
        this.values.remove(val);
        for (Value def : val.args()) {
            def.removeUse(val);
        }
    }

    public boolean removeSucc(Block succ) {
        return this.succs.remove(succ);
    }

    public boolean removePred(Block pred) {
        return this.preds.remove(pred);
    }

    public void setCtrl(Value ctrl) {
        this.ctrl = ctrl;
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        if (!this.preds.isEmpty()) {
            sb.append("b").append(this.id).append(": [");
            for (int i = 0; i < this.preds.size(); i++) {
                if (i == this.preds.size() - 1) {
                    sb.append("b").append(this.preds.get(i).id);
                } else {
                    sb.append("b").append(this.preds.get(i).id).append(" ");
                }
            }
            sb.append("]\n");
        } else {
            sb.append("b").append(this.id).append(":\n");
        }
        for (Value val : this.values) {
            sb.append(" ").append(val).append("\n");
        }
        if (this.ctrl != null) {
            sb.append(" ").append(this.kind).append(" v").append(this.ctrl.id()).append(" ");
        } else {
            sb.append(" ").append(this.kind).append(" ");
        }
        if (!this.succs.isEmpty()) {
            sb.append("[");
            for (int i = 0; i < this.succs.size(); i++) {
                if (i == this.succs.size() - 1) {
                    sb.append("b").append(this.succs.get(i).id);
                } else {
                    sb.append("b").append(this.succs.get(i).id).append(" ");
                }
            }
            sb.append("]");
        }
        return sb.toString();
    }
}
