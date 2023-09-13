default: test

bcl: *.go y.go cmd/bcl/*.go
	go build ./cmd/bcl

y.go: lang.y
	goyacc -l lang.y

clean:
	rm -f bcl

generated: y.go

test: generated
	go test -v ./...

bench: generated
	 go test -bench=. -count=$(cnt) -benchmem .
cnt=5

cover:
	go test -coverprofile=cov .
	go tool cover -html=cov -o cov.html && browse cov.html

.PHONY: default generated test bench cover clean
