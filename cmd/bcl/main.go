package main

import (
	"fmt"
	"io"
	"os"

	"github.com/wkhere/bcl"
)

type parsedArgs struct {
	file string
	help func()
}

func parseArgs(args []string) (a parsedArgs, _ error) {
	const usage = "usage: bcl [FILE|-]"
	switch {
	case len(args) == 0:
		a.file = "-"

	case args[0] == "-h":
		a.help = func() {
			fmt.Println(usage)
		}
		return a, nil

	case len(args) == 1:
		a.file = args[0]

	default:
		return a, fmt.Errorf(usage)
	}
	return a, nil
}

func readBuffer(file string) ([]byte, error) {
	if file == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(file)
}

func run(a *parsedArgs) error {
	buf, err := readBuffer(a.file)
	if err != nil {
		return err
	}

	res, err := bcl.Interpret(buf)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", res)
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
