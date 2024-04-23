default: test

generated = tokentype_string.go opcode_string.go

bcl: stringer $(generated) *.go cmd/bcl/*.go
	go build ./cmd/bcl

tokentype_string.go: token.go
	go generate

opcode_string.go: opcode.go
	go generate

clean:
	rm -f bcl

test: bcl test.py
	go test .
	./test.py

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
