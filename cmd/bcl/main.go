package main

import (
	"fmt"
	"io"
	"os"

	"github.com/wkhere/bcl"
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

	res, err := bcl.Interpret(buf)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", res)
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
