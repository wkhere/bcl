default: test

bcl: *.go y.go cmd/bcl/*.go
	go build ./cmd/bcl

y.go: lang.y
	goyacc -l lang.y
	sed -i 1d y.go # remove 'Code generated by goyacc' as it messes w/godoc
clean:
	rm -f bcl

generated: y.go

test: generated
	go test ./...

sel=.
cnt=5

bench: generated
	 go test -bench=$(sel) -count=$(cnt) -benchmem .

cov:
	go test -coverprofile=cov -run=$(sel) .
	go tool cover -html=cov -o cov.html

.PHONY: default generated test bench cov clean
