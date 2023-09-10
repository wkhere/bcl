bcl: *.go y.go cmd/bcl/*.go
	go build ./cmd/bcl

y.go: lang.y
	goyacc lang.y

clean:
	rm -f bcl

generated: y.go

test: generated
	go test .

bench: generated
	 go test -bench=. -count=$(cnt) -benchmem .
cnt=5

.PHONY: generated test bench clean
