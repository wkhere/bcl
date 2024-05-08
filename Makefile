default: test

generated = tokentype_string.go opcode_string.go typecode_string.go testapi_test.go

bcl: stringer $(generated) go.mod go.sum *.go cmd/bcl/*.go
	go build ./cmd/bcl

tokentype_string.go: token.go
	go generate token.go

opcode_string.go: opcode.go
	go generate opcode.go

typecode_string.go: typecode.go
	go generate typecode.go

testapi_test.go: test.py api_test.go
	go generate api_test.go && go fmt testapi_test.go

clean:
	rm -f bcl

sel=.
cnt=6

test: bcl test.py
	go test -cover -run=$(sel) .

bench: bcl
	go test -bench=$(sel) -count=$(cnt) -benchmem .

cov:
	go test -coverprofile=cov -run=$(sel) .
	go tool cover -html=cov -o cov.html

stringer: $(shell go env GOPATH)/bin/stringer
$(shell go env GOPATH)/bin/stringer:
	go install golang.org/x/tools/cmd/stringer@latest

.PHONY: default generated clean test bench cov stringer
