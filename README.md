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
* block starts with just a resource type, then comes a name, then the fields 
  inside the brackets; no need for an artificial `resource` keyword
* dead-simple API: variables get evaluated automatically and fill the fields of
  the output structure;
  no [strange limitations] of where variables can be defined
* detection of variable reference cycles
* _O(N)_ lexer, parser and evaluator
* deserialization aka unmarshalling to static Go structs

While being basic, BCL also has features reaching beyond HCL:

* rich expressions operating on strings, int & float numbers, and booleans
* while variables are defined within the file, the outside world is accessible
  via environment variables, or via running a command and catching its output
  (both features planned)

### Example:
BCL:
```hcl
var domain = "acme.com"
var default_port    = 8400
var local_port_base = default_port + 1000

tunnel "myservice-prod" {
	host = "prod" + "." + domain
	local_port  = local_port_base + 1
	remote_port = default_port
	enabled = true
	max_latency = 8.5 # [ms]
}
```
Go:
```Go
type Tunnel struct {
	Name                  string
	Host                  string
	LocalPort, RemotePort int
	Enabled               bool
	MaxLatency            float64
}
var config []Tunnel

err := bcl.Unmarshal(fileContent, &config)
// handle err
fmt.Printf("%+v\n", config)

```
Output:
```
[{Name:myservice-prod Host:prod.acme.com LocalPort:9401 RemotePort:8400 Enabled:true MaxLatency:8.5}]
```
### Expressions, data conversions

..to be documented..

### Note on the parser

I am using goyacc at the early stage and I find it very useful especially when
the grammar is still being changed to some extent. The often-discussed issues
with yacc, like resolving conflicts, are non-existent in my case - it seems
pretty trivial to get it right for a sensible grammar. The error reporting is
just fine with some minimal hacks, though I don't plan at this moment to do
a serious parser sync to avoid a cascade of errors
-- in this early implementation just the first error is reported.

Though it may see inconvenient, BCL will have top-down parser
and a small intermediate VM. It is exciting prospect to code it.


### TODO

* [x] get line info into eval errors
* [ ] more data types
  * [x] floats
  * [ ] lists
  * [ ] nested blocks?
* [x] more operators
  * [x] comparison
  * [x] boolean algebra
* [ ] string interpolation
* [ ] `getenv()` builtin
* [ ] unmarshaling options
  * [ ] allow fields to be missing in the target struct
  * [ ] allow struct type to be named differently than block type
        (at runtime, without struct tags)
* [ ] port to more programming languages


[strange limitations]: https://stackoverflow.com/a/73745980/229154
