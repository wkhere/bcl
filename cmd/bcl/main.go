package main

import (
	"fmt"
	"os"

	"github.com/wkhere/bcl"
)

func open(file string) (*os.File, error) {
	if file == "-" {
		return os.Stdin, nil
	}
	return os.Open(file)
}

func run(a *parsedArgs) (err error) {

	f, err := open(a.file)
	if err != nil {
		return err
	}

	res, err := bcl.InterpretFile(
		f,
		bcl.OptDisasm(a.disasm),
		bcl.OptTrace(a.trace),
		bcl.OptStats(a.stats),
	)
	if err != nil {
		return err
	}
	if a.result {
		fmt.Printf("result: %+v\n", res)
	}
	return nil
}

func main() {
	a, err := parseArgs(os.Args[1:])
	if err != nil {
		die(2, err)
	}
	if a.help != nil {
		a.help()
		os.Exit(0)
	}

	err = run(&a)
	if err != nil {
		die(1, err)
	}
}

func die(exitcode int, err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(exitcode)
}
