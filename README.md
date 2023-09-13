BCL
===

[![Build Status](https://github.com/wkhere/bcl/actions/workflows/go.yml/badge.svg)](https://github.com/wkhere/bcl/actions/workflows/go.yml)
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
* _O(N)_ lexer and parser
* deserialization aka unmarshalling to static Go structs

### Example:
BCL:
```hcl
var domain = "internal.acme.com"
var default_port    = 8400
var local_port_base = default_port + 1000

tunnel "myservice-prod" {
	host = "prod" + "." + domain
	local_port  = local_port_base + 1
	remote_port = default_port
	enabled = true
}
```
Go:
```Go
type Tunnel struct {
	Host                  string
	LocalPort, RemotePort int
	Enabled               bool
}
var config []Tunnel

err := bcl.Unmarshal(fileContent, &config)
// handle err
fmt.Printf("%+v\n", config)

```
Output:
```
[{Name:"myservice-prod" LocalPort:9401 RemotePort:8400 Host:prod.internal.acme.com Enabled:true}]
```
### Expressions, data conversions

..to be documented..

### TODO

* [ ] data types:
  - [ ] lists
  - [ ] floats
  - [ ] nested blocks?
* [ ] more operators
* [ ] string interpolation
* [ ] `getenv()` builtin

* [ ] unmarshalling options:
  - [ ] allow fields to be missing in the target struct

* [ ] port to more programming languages


[strange limitations]: https://stackoverflow.com/a/73745980/229154