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

Variation:

* prog can be concatenated, but chunks need to have 1-1 correspondence to functions,
  like in the book; so, what is in prog will become a chunk tied to a single file -
  no complications with linePos then



## Reflection revamp

Extra syntax for binding parsed blocks to various structs via reflection:

`bind <block_type>:<block_selector>`

Examples:
```HCL
bind tunnel:1  # there must be just 1 tunnel block; same as:
bind tunnel

bind tunnel:first
bind tunnel:last
bind tunnel:"name"
bind tunnel:all              # bind to a slice
bind tunnel:"name1","name2"  # bind to a slice
bind tunnel:"name",          # trailing comma also means: bind to a slice
```

The Go target is supposed to be a struct or a slice based on cardinality of
the chosen selector:

* keywords `1` `first` `last` denote cardinality of one,
* block selected by a name in doublequotes also denote cardinality of one,
* keyword `all` denotes cardinality of many,
* multiple blocks selected by doublequoted names, separated by a comma, denote
cardinality of many,
* a single named block but with a trailing comma is also about cardinality of many.

When binding to a struct, the Go target should be a pointer to a struct of a type
matching the block type.
Similarly, for a slice, the Go target should be a pointer to a slice of structs
matching the block type.

Often there is a need to bind different kinds of block in a single API invocation.
For that, another form of `bind` can be used - it is called "umbrella bind":

```HCL
bind {
  wormhole:"galaxy42"
  probe:all
  setup:1
}
```
This should correspond on Go side to an "umbrella struct" - can be anonymous struct:
```Go
var x struct { W Wormhole; PP []Probe; Setup Setup }
err := bcl.UnmarshalFile(input, &x)
```
Inside such an umbrella same rules apply as for the single `bind` in terms of expecting
a block-corresponding struct or a slice of them.

Few extra rules: 

* Bind statement can be defined only at the toplevel and corresponds to blocks defined
before.
* Subsequent bind statement ovverrides previous one and prints a warning.
