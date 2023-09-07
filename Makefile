tmon: *.go y.go
	go build

y.go: lang.y
	goyacc lang.y

clean:
	rm tmon

generated: y.go

test: generated
	go test

bench: generated
	 go test -bench=. -count=$(cnt) -benchmem .
cnt=5

.PHONY: generated test bench clean
