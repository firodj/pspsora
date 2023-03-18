# pspsora

psp disasm / trace parser from PPSSPP sora brach.

## development

```
$ go run main.go -- testDisasm
```

## todos

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