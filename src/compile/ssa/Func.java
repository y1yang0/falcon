package compile.ssa;

import java.util.ArrayList;
import java.util.List;

public class Func {
    private final String name;
    private final Block entry;
    private final List<Block> blocks;

    public Func(String name) {
        this.name = name;
        this.entry = null;
        this.blocks = new ArrayList<>();
    }

    public Block newBlock(BlockKind kind) {
        Block block = new Block(this, kind);
        this.blocks.add(block);
        return block;
    }

    public void removeBlock(Block block) {
        this.blocks.remove(block);
        for (Value val : block.values()) {
            block.removeValue(val);
        }
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        for (Block block : this.blocks) {
            sb.append(block).append("\n");
        }
        return sb.toString();
    }
}