## Vars, fields and scopes

Toplevel:

- vars are defined with `var`
- vars are local, it is simply the toplevel scope, accessible from everywhere
- no forward declarations

Block:

- vars are defined with `var`
- vars are local
- vars referenced are searched in the current scope, then in the outer scopes
  up to the toplevel
- no forward declarations
- assignments to non-var identifiers are updates of block fields 

Last point requires some deeper digging. It might be that what looks like
block field update will be actually an inner var update. It will be better
to provide some syntax to make it clear.

Similar with print vs eval: now the var update can happen inside print, not only
eval; it can actually happen anywhere expr is expected, because assignments are
also expressions. This might be surprising for some users, while others can
possibly embrace it. So, restricting when assignment can happen is to be thought
about carefully.


Note the differences from yacc-based bcl, where vars were global and it was
possible to have forward declarations, with some special detection of cycles.


## Reading multiple files

To be decided: many progs, or one concatenated prog?

Many progs:

* each input file is a separate compilation unit and produces a prog
* what about visibility of the vars and blocks? need either 'export' syntax in
  the referred file, or 'extern' syntax in the referring file, or a form of both
* more elegant but more complex and more changes to the API

Concatenated prog:

* `include` syntax
* there is one compilation unit, with tokens coming from multiple input files
* lexer needs to handle multiple inputs (done)
* token.pos needs to be extended with a file name (tbd)
* same for prog positions and linePos, actually the prog needs a filename table
  and an index to that table with each pos (tbd)
* simpler and less changes to the API


## Reflection revamp

Extra syntax for binding parsed blocks to various structs via reflection:

`bind <block_type>:<block_selector> -> <target_type>`

Examples:
```
bind tunnel:1 -> struct   # must be just 1 block of type tunnel; same as:
bind tunnel   -> struct

bind tunnel:last  -> struct
bind tunnel:first -> struct
bind tunnel:all   -> slice
bind tunnel:"name" -> struct
bind tunnel:"name1","name2" -> slice
```

Rules for bind targets based on a target type:

- `struct`: target is a ptr-to-struct of a type matching the block type
- `slice`:  target is a ptr-to-slice of structs of a type matching the block type

Bind rule can be defined only at the toplevel and correspond to blocks defined
before. (VM instruction impl should use already evaluated `[]Block`).

Subsequent bind rule ovverrides previous one and should generate a warning.
