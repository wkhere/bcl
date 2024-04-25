default: test

generated = tokentype_string.go opcode_string.go apigen_test.go

bcl: stringer $(generated) *.go go.mod cmd/bcl/*.go
	go build ./cmd/bcl

tokentype_string.go: token.go
	go generate

opcode_string.go: opcode.go
	go generate

apigen_test.go: test.py gen_test.go
	go generate

clean:
	rm -f bcl

test: bcl test.py
	go test -cover .

bench: bcl
	go test -bench=. -count=$(cnt) -benchmem .

cov:
	go test -coverprofile=cov -run=$(sel) .
	go tool cover -html=cov -o cov.html
sel=.
cnt=6

stringer: $(shell go env GOPATH)/bin/stringer
$(shell go env GOPATH)/bin/stringer:
	go install golang.org/x/tools/cmd/stringer@latest

.PHONY: default generated clean test bench cov stringer
