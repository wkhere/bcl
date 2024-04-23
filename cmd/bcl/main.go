package main

import (
	"fmt"
	"io"
	"os"

	"github.com/wkhere/bcl"
)

func readBuffer(file string) ([]byte, error) {
	if file == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(file)
}

func run(a *parsedArgs) (err error) {

	buf, err := readBuffer(a.file)
	if err != nil {
		return err
	}

	res, err := bcl.Interpret(
		buf,
		bcl.OptDisasm(a.disasm),
		bcl.OptTrace(a.trace),
	)
	if err != nil {
		return err
	}
	if !a.mute {
		fmt.Printf("%+v\n", res)
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
