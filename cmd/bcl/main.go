package main

import (
	"fmt"
	"os"

	"github.com/wkhere/bcl"
)

func openInput(file string) (f *os.File, st fileStat, err error) {
	if file == "-" {
		f = os.Stdin
		st, err = statFile(f)
		return f, st, err
	}

	f, err = os.Open(file)
	if err != nil {
		return nil, st, err
	}
	st, err = statFile(f)
	if err != nil {
		f.Close()
		return nil, st, err
	}
	return f, st, nil
}

// openOutput opens the file for writing, refusing to clobber the input.
// Truncation happens only after the same-file check, so that a symlink
// or hardlink pointing to the input can't destroy it.
func openOutput(file string, force bool, inputSt fileStat) (*os.File, error) {
	flag := os.O_WRONLY | os.O_CREATE
	if !force {
		flag |= os.O_EXCL
	}
	f, err := os.OpenFile(file, flag, 0644)
	if err != nil {
		return nil, err
	}
	st, err := statFile(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	if st.sameFile(inputSt) {
		f.Close()
		return nil, fmt.Errorf("%s: same file as the input", file)
	}
	if st.isRegular() {
		err = f.Truncate(0)
		if err != nil {
			f.Close()
			return nil, err
		}
	}
	return f, nil
}

func run(a *parsedArgs) (err error) {

	f, st, err := openInput(a.file)
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
		bf, err := openOutput(a.bdumpFile, a.force, st)
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
