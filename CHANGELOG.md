### v0.13.5
* license: anti-mining changes
* readme polishing, notes update

### v0.13.4
* reflection: implicit BCL int -> Go float64 conversion
* parser: block names are interned, producing less constants
* cmd/bcl: --bdump improvements
* cmd/bcl: one-letter flags optimization
* improved docs, notes, comments, tests

### v0.13.3
* update readme & API docs

### v0.13.2
* fix error messages in the new reflection code
* update notes

### v0.13.1
Nothing really changes, just fixed some typos in the changelog itself.

### v0.13.0 "lucky thirteen"
This version brings refreshed `bind` syntax, allowing for an "umbrella bind"
to get different block types into Go struct/slices with a single API call.
Also, the `-> struct|slice` syntax has been removed, in favor of simple inference
based on block(s) selector. See [notes](NOTES.md#reflection-revamp).

More changes:
* bytecode version 2.0
* show mismatched bytecode version when reading older bytecode dump
* improved warnings, error messages, benchmarks & docs
* snake: spare one in-loop call to unicode func

### v0.12.2
* bind multiple named blocks
* allow binding single named block to a 1-element slice
  (same as with :first,:last,:1)
* bytecode version 1.3 (bind instruction carrying multiple block names)

### v0.12.1
The main highlight of this version is allowing the bind of a named block
(a single block, for now).

More changes:

* bytecode version 1.2
  (new bind selector requires an instruction carrying a block name)
* better notes on alternatives for multiple input files
* notes on scope issues and possible evolution 
* note on reflection bug with inner blocks
* fix cmd: missing Close after bload
* snake-case goodies, not fully used yet
* minor optimisations, source layout

### v0.12.0 "bind polishing"
* API change: Unmarshal/UnmarshalFile with variadic options
* API change: remove TypeErr, StructErr
* bind improvements: errors, warnings, tests, docs
* silence expected errors in reflect tests / benchmarks
* encoding fix: always write typeNIL

### v0.11.0
* bind syntax
* API change: Execute/Interpret return Binding as well
* API change: CopyBlocks -> Bind
* bytecode version 1.1

This new mechanism gives a simple and robust way to bind parsed
blocks to Go structs (single struct or a slice of structs).

### v0.10.12
* shorten license extra text (pkg.go.dev policy)

### v0.10.11
* test.py: allow backticks in error messages
* lex: return syntax errors for "sticky tokens"
* cmd/bcl: better error for --bdump flag

### v0.10.10
* fix TestEarlyParseErr where intentional error was printed
* improve test.py
* simplify encoding code
* simplify handling of 2-,1-rune tokens
* cosmetic changes

### v0.10.9
* small fix in reflect

### v0.10.8
* nameless block is keyed by its type

### v0.10.7
* fix fake-eof and related behavior in lex/parse

### v0.10.6
* simplify test dep
* parse: code reuse, small fixes

### v0.10.5
* ParseFile: cancel reads on lex errors; fix more data races
* fix rare case of token leftover at end

This version solves a number of nontrivial concurrency and buffering issues.

### v0.10.4
* ParseFile: fix data race
* makefile & CI improvements

### v0.10.3
* ParseFile by pages - even better
* fix 'token leftover bug' (hard to show when parsing by lines, now important)
* fix unary plus disasm
* test/benchmark/doc improvements

### v0.10.2
* ParseFile by lines - fixes handling big files
* trace: nicer stack printout
* test/benchmark improvements

### v0.10.1
* fix: disasm when no errors occured

### v0.10.0
* cmd/bcl: --bdump, --bload flags
* serialize Prog via Dump/Load; new API: LoadProg, Execute 
* cmd/bcl: break flags processing on --
* build/test/benchmark improvements
* BROKEN disasm - use v0.10.1

### v0.9.0
* new API: InterpretFile, UnmarshalFile
* use input file name in disasm/trace output
* simplify linecalc
* arch-dependent tests for maxint
* various notes

### v0.8.10
* lang: make numerical zero falsey

### v0.8.9 - prep for prog serialization
* test gen: separate pkg
* cleanup emit functions
* safety fix: proper buffer for encoding a big varint
* switch u16 encoding to big endian, to match varint
* store linepos struct instead of a closure
* various notes

### v0.8.8
* fix subtle bug in tests generator from v0.8.5
* tests generator for error cases
* cover a lot of error cases from both parser & vm runtime
* lang change: allow string+nil op
* lang fix: unary plus works only for numbers
* fix division by int zero

### v0.8.7
* CI & makefile changes
* refine license

### v0.8.6
* fix indentation of generated tests

### v0.8.5
* generate api tests from test.py, fix
  Note: this adds API option to redirect the output

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
* require Go 1.21

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
