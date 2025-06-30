BCL
===

[![Build Status](https://github.com/wkhere/bcl/actions/workflows/go.yml/badge.svg)](https://github.com/wkhere/bcl/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/wkhere/bcl/badge.svg?branch=master)](https://coveralls.io/github/wkhere/bcl?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/wkhere/bcl)](https://goreportcard.com/report/github.com/wkhere/bcl)
[![Go Reference](https://pkg.go.dev/badge/github.com/wkhere/bcl.svg)](https://pkg.go.dev/github.com/wkhere/bcl)

Basic Configuration Language.

__BCL__ is like HCL,
but instead if being tied to specific HashiCorp products, it brings to the table more features from a typical programming language:

* rich expressions: full numerical arithmetics, operations on strings and booleans
* no dollar-referenced variables; just use the variable name
* variables with lexical scope, nested definitions; no [strange limitations] of where variables can be defined
* one-pass lexer, parser and VM executor
* deserialization aka unmarshalling to static Go structs (possibly nested) via
  [bind](NOTES.md#reflection-revamp) statement
* optimized to parse very large input if needed
* planned: make the outside world accessible via environment variables and via catching the command output

### Example:
BCL:
```hcl
var domain = "acme.com"
var default_port    = 8400
var local_port_base = default_port + 1000

def tunnel "myservice-prod" {
	host = "prod" + "." + domain
	local_port  = local_port_base + 1
	remote_port = default_port
	enabled = true

	def extras {
		max_latency = 8.5 # [ms]
	}
}

bind tunnel
```
Go:
```Go
type Tunnel struct {
	Name       string
	Host       string
	LocalPort  int
	RemotePort int
	Enabled    bool
	Extras     struct {
		MaxLatency float64
	}
}
var config Tunnel

err := bcl.UnmarshalFile(file, &config)
fmt.Println(strings.ReplaceAll(fmt.Sprintf("%+v", config), " ", "\n  "))
```
Output:
```
{Name:myservice-prod
  Host:prod.acme.com
  LocalPort:9401
  RemotePort:8400
  Enabled:true
  Extras:{MaxLatency:8.5}}
```
### Syntax

BCL has statements and expressions.

A basic statement is `def block_type [block_name] {...}` to define a block with
`field = value` expressions inside.
Such block after running [Interpret] will be available as 
a [Block] with a map of fields,
and can be put into a static Go struct via [Bind] or [Unmarshal].
Blocks can be nested.

Both toplevel scope and a block can have variables created with 
the `var x = expr` statement, or just `var x` which leaves it uninitialized.
Variables do not count when produding result Block structures, but they are
taking part of the state flow.

Variables have lexical scope. Any block has access to the varables declared
at the toplevel and also nested block have access to their parent's variables.
There are no forward declarations.

Variables are mutable and can be mutated with the `eval x = expr` statement.
This statement solely exists to allow evaluation of expressions in the context
expecting stamenents, that is at the toplevel. Please note that inside the block
the raw statements are allowed, for example `field = value` is actually
an assignment expression. So, when in block, it's good to remember whether 
we are operating on fields or variables. This may be made more explicit in the future.

The last stament in the clan is `print expr` which is useful for debugging.

More on expressions below.

### Expressions, data conversions

There are three basic types: numbers (int and float), strings and booleans.

Values in expressions know their types, although they are not enforced
in the language; certain operations can cause runtime error.

Number arithmetics use int or float operations depending on the values
involved; if any of the operands is float, then the int part is transparently
converted to float. Complex numbers are not supported atm.

Strings can be concatenated with the plus `+`. 
If the right side of such plus is a number, it will be transparently
coverted to string. However, the number plus string is an error.

Another string operator borrowed from numbers is asterisk `*`, this time
the left side must be a string and right side just an int; the result is
repeating the string given times.

Equality comparisons `==`, `!=` are allowed between all types, including mixing them.
Obviously values of different non-number types are not equal.

Order comparisons `<`, `>`, , `<=`, `>=` are allowed between numbers and between strings,
but not between mixed types.

There are boolean operators `and`, `or`, `not` behaving like in Python,
or like `&&`, `||`, `!` in Ruby [1]:
they are short-cirtuit and retain the type of an operand 
(`1==1 and 42` will return 42). Non-boolean types can be a boolean operand;
for this, there is a definition what is considered "falsey": `false`, `nil`,
empty string, and zero, like in Python.

Boolean constants are `true` and `false`.
Another constant is `nil`, value of an uninitialized variable (`var a`).

[1] Note that in Ruby `!` has surprising priority, though.


### Note on the parser

Versions up to v0.7.x used goyacc, since v0.8.0 there is a top-down Pratt parser
with bytecode VM.


### Cool stuff

Internals can be peeked in many ways, here is bytecode disassembly,
execution trace with stack content, plus some stats:
```
./bcl -dts <<<'var x=1; def block{eval x=x+2; field=x}'
== /dev/stdin ==
0000    1:8  ONE
0001   1:20  DEFBLOCK      0 'block'         1 ''
0004   1:28  GETLOCAL      0
0006   1:30  CONST         2 '2'
0008      |  ADD
0009      |  SETLOCAL      0
0011      |  POP
0012   1:39  GETLOCAL      0
0014      |  SETFIELD      3 'field'
0016      |  POP
0017   1:40  ENDBLOCK
0018    2:1  POP
0019      |  RET
pstats.tokens:        20
pstats.localMax:       1
pstats.depthMax:       1
pstats.constants:      4
pstats.opsCreated:    13
pstats.codeBytes:     20
             0: 
0000    1:8  ONE
             1: [ 1 ]
0001   1:20  DEFBLOCK      0 'block'         1 ''
             1: [ 1 ]
0004   1:28  GETLOCAL      0
             2: [ 1 ][ 1 ]
0006   1:30  CONST         2 '2'
             3: [ 1 ][ 1 ][ 2 ]
0008      |  ADD
             2: [ 1 ][ 3 ]
0009      |  SETLOCAL      0
             2: [ 3 ][ 3 ]
0011      |  POP
             1: [ 3 ]
0012   1:39  GETLOCAL      0
             2: [ 3 ][ 3 ]
0014      |  SETFIELD      3 'field'
             2: [ 3 ][ 3 ]
0016      |  POP
             1: [ 3 ]
0017   1:40  ENDBLOCK
             1: [ 3 ]
0018    2:1  POP
             0: 
0019      |  RET
xstats.tosMax:         3
xstats.blockTosMax:    1
xstats.opsRead:       13
xstats.pcFinal:       20
```


[strange limitations]: https://stackoverflow.com/a/73745980/229154
[Block]: https://pkg.go.dev/github.com/wkhere/bcl#Block
[Interpret]:  https://pkg.go.dev/github.com/wkhere/bcl#Interpret
[Bind]:       https://pkg.go.dev/github.com/wkhere/bcl#Bind
[Unmarshal]:  https://pkg.go.dev/github.com/wkhere/bcl#Unmarshal
