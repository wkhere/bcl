// Package bcl provides interpreting of the Basic Configuration Language (BCL)
// and storing the evaluated result in dynamic Blocks or static structs.
//
//   - [Interpret] or [InterpretFile] parses and executes definitions
//     from a BCL file, then creates Blocks
//   - [CopyBlocks] takes Blocks and saves the content in static Go structs
//   - [Unmarshal] = [Interpret] + [CopyBlocks]
//   - [UnmarshalFile] = [InterpretFile] + [CopyBlocks]
//
// It is also possible to first [Parse], creating [Prog], and then [Execute] it.
//   - [Interpret] = [Parse] + [Execute]
//   - [InterpretFile] = [ParseFile] + [Execute]
//
// [Prog] can be dumped to a Writer with Dump and loaded with Load,
// there is also wrapper function [LoadProg], to load previously dumped Prog
// instead of using Parse on the BCL input.
package bcl

import (
	"bufio"
	"io"
)

// FileInput abstracts the input that is read from a file.
// It is going to be closed as soon as it's read.
// The only information needed from a file besides reading/closing
// is that it has a name.
type FileInput interface {
	io.ReadCloser
	Name() string
}

// Block is a dynamic result of running BCL [Interpret].
// It can be put into a static structure via [CopyBlocks].
type Block struct {
	Type, Name string
	Fields     map[string]any
}

// Parse parses the input data, producing executable Prog.
func Parse(input []byte, name string, opts ...Option) (*Prog, error) {
	c := make(chan string, 1)
	c <- string(input)
	close(c)
	return parseWithOpts(c, name, opts)
}

// ParseFile reads and parses the input from a BCL file,
// producing executable Prog.
// The input file will be closed as soon as possible.
func ParseFile(f FileInput, opts ...Option) (prog *Prog, _ error) {
	inpc := make(chan string)
	errc := make(chan error, 1)

	go func() {
		b := bufio.NewReader(f)
		for {
			line, err := b.ReadSlice('\n')
			if err != nil && err != io.EOF {
				errc <- err
				break
			}
			if err == io.EOF && len(line) == 0 {
				break
			}
			inpc <- string(line)
		}
		f.Close()
		close(inpc)
	}()

	go func() {
		p, err := parseWithOpts(inpc, f.Name(), opts)
		prog = p
		errc <- err
	}()

	err := <-errc
	return prog, err
}

func parseWithOpts(inputs <-chan string, name string, opts []Option) (*Prog, error) {
	cf := makeConfig(opts)

	prog, pstats, err := parse(inputs, name, parseConfig{cf.output, cf.logw})
	if err == nil && cf.disasm {
		prog.disasm()
	}
	if cf.stats {
		printPStats(cf.output, pstats)
	}
	return prog, err
}

func LoadProg(r io.Reader, name string, opts ...Option) (*Prog, error) {
	cf := makeConfig(opts)

	prog := newProg(name, cf.output)
	err := prog.Load(r)
	if err == nil && cf.disasm {
		prog.disasm()
	}
	return prog, err
}

// Execute executes the Prog, creating Blocks.
func Execute(prog *Prog, opts ...Option) (result []Block, err error) {
	cf := makeConfig(opts)

	result, xstats, err := execute(prog, vmConfig{cf.trace})
	if cf.stats {
		printXStats(cf.output, xstats)
	}
	return result, err
}

// Interpret parses and executes the BCL input, creating Blocks.
func Interpret(input []byte, opts ...Option) ([]Block, error) {
	p, err := Parse(input, "input", opts...)
	if err != nil {
		return nil, err
	}
	return Execute(p, opts...)
}

// InterpretFile reads, parses and executes the input from a BCL file.
// The file will be closed as soon as possible.
func InterpretFile(f FileInput, opts ...Option) ([]Block, error) {
	p, err := ParseFile(f, opts...)
	if err != nil {
		return nil, err
	}
	return Execute(p, opts...)
}

// CopyBlocks copies the blocks to the dest,
// which needs to be a pointer to a slice of structs.
//
// The requirements for the struct are:
//   - struct type name should correspond to the BCL block type
//   - struct needs the Name string field
//   - for each block field, struct needs a corresponding field, of type as
//     the evaluated value (currently supporting int, string and bool)
//
// The mentioned name correspondence is similar to handling json:
// as BCL is expected to use snake case, and Go struct - capitalized camel case,
// the snake underscores are simply removed and then the strings are compared,
// case-insensitive.
//
// The name corresponcence can be also set explicitly,
// by typing a tag keyed `bcl` at the struct field:
//
//	type Record struct {
//		Status string `bcl:"my_status"`
//	}
//
// The lack of corresponding fields in the Go struct is reported as error.
// So is type mismatch of the fields.
//
// If the slice pointed by dest contained any elements, they are overwritten.
func CopyBlocks(dest any, blocks []Block) error {
	return copyBlocks(dest, blocks)
}

// Unmarshal interprets the BCL input, and stores the result in dest,
// which should be a slice of structs.
// See [CopyBlocks] for a struct format.
func Unmarshal(input []byte, dest any) error {
	res, err := Interpret(input)
	if err != nil {
		return err
	}
	return CopyBlocks(dest, res)
}

// UnmarshalFile interprets the BCL file and stores the result in dest,
// which should be a slice of structs.
// See [CopyBlocks] for a struct format.
func UnmarshalFile(f FileInput, dest any) error {
	res, err := InterpretFile(f)
	if err != nil {
		return err
	}
	return CopyBlocks(dest, res)
}
