### v0.6.1
* get line info to eval errors: in the impl, position is carried over from lexer
  to ast and is used in eval, so it's possible in the future to have
  more sophisticated error reporting than a line number

### v0.6.0
* API change: CopyBlocks, not AppendBlocks, so it copies, not appends
* lex: record item's position not line number,
  calculate line numbers only when errors are printed;
  this fixes wrong reporting of line numbers for some parse errors
* cmd/bcl: read a file as well as stdin
* introduce this changelog
  (future version tags will not have annotations, description of changes
   should go here)

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
