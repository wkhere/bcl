### v0.8.4
* generate api example tests from test.py
* more coverage

### v0.8.3
* use variable-length ints in the bytecode where it's possible
* parse & execution stats, avail. via cmd/bcl -s,--stats flag
* cmd/bcl -r,--result flag; can intermix flags and file arg
* disasm & trace improvements
* lexer: cleanup unused fields, return closed channel info
* improvements of error messages, makefile & docs
* Forth-like stack comments in the vm code as the extra doc

### v0.8.2
* return error for non-empty stack on prog end
* fix stack handling for in-block exprs
* more docs

### v0.8.1
* allow unmarshaling to struct without name field if BCL block didn't have name
* parser: fix handling missing var refs at the toplevel
* improve docs

### v0.8.0
* switch to vm-based parser
* variables behavior changed a bit (global -> local, can shadow outer scope)

### v0.7.3
* fix exp dependency

### v0.7.1
* fix lexing a line with a single comment char

### v0.7.0
* language change: `def` keyword for blocks 

### v0.6.6
* yeah, the day of fixin the docs

### v0.6.5
* fix docs again

### v0.6.4
* fix docs

### v0.6.3
* bool algebra with and, or, not
* improved tests

Note: it's great to have _both_ changelog and version tag annotations,
the latter were missing since v0.6.0, now I am filling them again.

### v0.6.2
* comparison operators
* improved tests & docs

### v0.6.1
* get line info into eval errors
  <br> Btw, in the implementation,
  position is carried over from lexer to ast and eval,
  so it's possible for errors to report even more.

### v0.6.0
* API change: CopyBlocks, not AppendBlocks, so it copies, not appends
* lex: record item's position not line number,
  calculate line numbers only when errors are printed;
  this fixes wrong reporting of line numbers for some parse errors
* cmd/bcl: read a file as well as stdin
* introduce this changelog

### v0.5.1
* improved tests & docs

### v0.5.0
* handle floats in the language: lexing, parsing, evaluation, operators
* improved syntax errors near eof

### v0.4.6
* handle division by zero - return error instead of panic
* allow saving to a slice of anonymous structs
* improved tests & docs

### v0.4.5
* token constants private now, no need to keep them exported

### v0.4.4
* fix: block name should be unquoted
* improved tests & docs

### v0.4.3
* fix yacc-generated file so it doesn't mess with the doc comments
* fix doc links

### v0.4.2
* improved tests & docs

### v0.4.1
* improved tests & docs
* CI stuff

### v0.4.0
* first API cleanup (obsolete now, anyway)

### v0.3.0
* save to []struct via reflection

### v0.2.0
* int arithmetics; bool still has only 'not' operator

### v0.1.0
* initial version: parse and eval to dict-like blocks, with a restricted grammar
