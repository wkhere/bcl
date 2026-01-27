default: bcl test

generated = tokentype_string.go opcode_string.go typecode_string.go testapi_test.go

src: stringer $(generated) go.mod go.sum *.go cmd/bcl/*.go

bcl: src
	go build ./cmd/bcl

install: src
	go install -ldflags=-s ./cmd/bcl

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

test: src
	go test -cover -run=$(sel) .

test-py:
	./test.py

test-race: src
	go test -race .

test-full: test test-py test-race

bench: src
	go test -bench=$(sel) -count=$(cnt) -benchmem .

cov: src
	go test -coverprofile=cov -run=$(sel) .
	go tool cover -html=cov -o cov.html

stringer: $(shell go env GOPATH)/bin/stringer
$(shell go env GOPATH)/bin/stringer:
	go install -ldflags=-s golang.org/x/tools/cmd/stringer@latest

.PHONY: default generated src install clean test test-py test-race test-full bench cov stringer
