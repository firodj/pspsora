# pspsora

psp disasm / trace parser from PPSSPP sora brach.

## development

```
$ go run main.go -- testDisasm
```

## todos

1. [❎] Using LLVM IR ? Not using! many tools can lift LLVM IR to be converted into C later,
   but need learnig curve.
2. [ ] Check PPSSPP IR
3. [ ] After BB trace, check function that have visited one of its return.
4. [ ] we may jump to AST (smartdec called it IR).

Decompile in General:
1. binary -> disasm
2. disasm -> pseudo/IR
   a. to String
   b. to AST
   c. to Idioms (xor eax,eax => mov eax,0)
3. expression propagation
4. dataflow anaylsis, temporary register
5. type analysis, struct
6. while, if/then/else restrcuture
7. highlevel code

## notes

* https://github.com/jdek/jim-psp
