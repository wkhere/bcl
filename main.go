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
	fmt.Println("vars:")
	fmt.Printf("\t%v\n", top.vars)
	fmt.Println("tunnels:")
	for _, x := range top.tunnels {
		fmt.Printf("\t%+v\n", x)
	}
	fmt.Println()
	return err
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
