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

	var prog *bcl.Prog

	if a.bload {
		prog, err = bcl.LoadProg(
			f, a.file,
			bcl.OptDisasm(a.disasm),
		)
		f.Close()
	} else {
		prog, err = bcl.ParseFile(
			f,
			bcl.OptDisasm(a.disasm),
			bcl.OptStats(a.stats),
		)
	}
	if err != nil {
		return err
	}

	if a.bdump {
		bf, err := os.Create(a.bdumpFile)
		if err != nil {
			return fmt.Errorf("dump: %w", err)
		}

		err = prog.Dump(bf)
		safeClose(bf, &err)
		if err != nil {
			return fmt.Errorf("dump: %w", err)
		}
	}

	res, binding, err := bcl.Execute(
		prog,
		bcl.OptTrace(a.trace),
		bcl.OptStats(a.stats),
	)
	if err != nil {
		return err
	}
	if a.result {
		fmt.Printf("result:  %+v\n", res)
		fmt.Printf("binding: %+v\n", binding)
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

func safeClose(f *os.File, errp *error) {
	cerr := f.Close()
	if cerr != nil && *errp == nil {
		*errp = cerr
	}
}
