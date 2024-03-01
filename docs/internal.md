# Internals

For those who are interested in the internals of our compiler, here is a brief
overview of the pipeline:

```
Source->Lexer->Token->Parser->AST->Infer->TypeCheck->TypedAst->Compiler->HIR->LIR->Asm->Linker->Executable
```

We start from the Falcon source code, which is then tokenized by the lexer. The tokens are then parsed into an abstract syntax tree (AST).

<img src="ast_object.png " width="400">

Type inference is then performed on the AST, because we encourage omitting type annotations in the source code. Consequently, the AST is transformed into a typed AST. The typed AST is then type-checked and transformed into a high-level intermediate representation (HIR).

<img src="ssa_bubbleSort.png " width="400">

After HIR construction, several classical optimizations are performed on the HIR, such as dead code elimination, CFG simplification, phi simplification, and local value numbering.

The HIR is then lowered into a low-level intermediate representation (LIR), which is then translated into assembly code. Finally, the assembly code is linked into an executable by using GCC toolchain.


- [GCC Assembly Syntax1](https://www.felixcloutier.com/documents/gcc-asm.html)
- [GCC Assembly Syntax2](https://ftp.gnu.org/old-gnu/Manuals/gas-2.9.1/html_chapter/as_toc.html)