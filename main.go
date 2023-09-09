package main

import (
	"fmt"
	"io"
	"os"
)

type args struct {
	//...
}

func parseArgs(aa []string) (*args, error) {
	//...
	return &args{}, nil
}

func run(r io.Reader) error {
	buf, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	top, err := parse(buf)
	if err != nil {
		return err
	}
	fmt.Println("vars:")
	fmt.Printf("\t%v\n", top.vars)
	fmt.Println("blocks:")
	for _, x := range top.blocks {
		fmt.Printf("\t%+v\n", x)
	}
	err = eval(&top)
	if err != nil {
		return err
	}

	fmt.Println()
	return nil
}

func main() {
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		die(2, err)
	}

	_ = args //tmp
	err = run(os.Stdin)
	if err != nil {
		die(1, err)
	}

}

func die(exitcode int, err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(exitcode)
}
