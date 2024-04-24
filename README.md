BCL
===

[![Build Status](https://github.com/wkhere/bcl/actions/workflows/go.yml/badge.svg)](https://github.com/wkhere/bcl/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/wkhere/bcl/badge.svg?branch=master&kill_cache=1)](https://coveralls.io/github/wkhere/bcl?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/wkhere/bcl)](https://goreportcard.com/report/github.com/wkhere/bcl)
[![Go Reference](https://pkg.go.dev/badge/github.com/wkhere/bcl.svg)](https://pkg.go.dev/github.com/wkhere/bcl)

Basic Configuration Language.

__BCL__ is like HCL (the language driving Terraform, Packer and friends),
but more basic:

* no dollar-referenced variables; variable name can be used in an expression
  as it is
* block starts with `def` keyword; seems to be more general
  compared to Terraform's `resource` keyword
* dead-simple API: variables get evaluated automatically and fill the fields of
  the output structure;
  no [strange limitations] of where variables can be defined
* variables with lexical scope
* nested definitions
* one-pass lexer, parser and VM executor
* deserialization aka unmarshalling to static Go structs (possibly nested)

While being basic, BCL also has features reaching beyond HCL:

* rich expressions operating on strings, int & float numbers, and booleans
* planned: make the outside world accessible via environment variables, or via
  running a command and catching its output

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
var config []Tunnel

err := bcl.Unmarshal(fileContent, &config)
// handle err
fmt.Println(strings.ReplaceAll(fmt.Sprintf("%+v", config), " ", "\n  "))
```
Output:
```
[{Name:myservice-prod
  Host:prod.acme.com
  LocalPort:9401
  RemotePort:8400
  Enabled:true
  Extras:{MaxLatency:8.5}}]
```
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

There are boolean operators `and`, `or`, `not` behaving like in Python and Ruby:
they are short-cirtuit and retain the type of an operand 
(`1==1 and 42` will return 42). Non-boolean types can be a boolean operand;
for this, there is a definition what is considered "falsey": `false`,
empty string, and `nil`. This may at resemble Python, but note that the zero number
is not considered falsey -- a behavior that may change in the future.

Boolean constants are `true` and `false`.
Another constant is `nil`, whose meaning is that
a declared but uninitialized variable (`var a`) has this value.
You can test equality to nil and it can be a part of a boolean expression,
acting as false.


### Note on the parser

Versions up to v0.7.x used goyacc, from v0.8.0 there is a top-down Pratt parser
with bytecode VM.


[strange limitations]: https://stackoverflow.com/a/73745980/229154
