// Package bcl provides interpreting of the Basic Configuration Language (BCL)
// and storing the evaluated result in dynamic Blocks or static structs.
//
//   - [Interpret] or [InterpretFile] parses and executes definitions
//     from a BCL file, then creates blocks and their binding
//   - [Bind] takes blocks binding and saves the content in static Go structs
//   - [Unmarshal] = [Interpret] + [Bind]
//   - [UnmarshalFile] = [InterpretFile] + [Bind]
//
// It is also possible to first [Parse], creating [Prog], and then [Execute] it.
//   - [Interpret] = [Parse] + [Execute]
//   - [InterpretFile] = [ParseFile] + [Execute]
//
// [Prog] can be dumped to a Writer with Dump and loaded with Load,
// there is also wrapper function [LoadProg], to load previously dumped Prog
// instead of using Parse on the BCL input.
package bcl

import "io"

// FileInput abstracts the input that is read from a file.
// It is going to be closed as soon as it's read.
// The only information needed from a file besides reading/closing
// is that it has a name.
type FileInput interface {
	io.ReadCloser
	Name() string
}

// Block is a dynamic result of running BCL [Interpret].
// It can be put into a static structure via [Bind].
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
	rerr := make(chan error)
	perr := make(chan error)
	done := make(chan struct{})

	go func() {
		defer f.Close()
		var b [4096]byte

		for {
			n, err := f.Read(b[:])
			if err != nil && err != io.EOF {
				rerr <- err
				break
			}
			if err == io.EOF && n == 0 {
				rerr <- nil
				break
			}
			select {
			case inpc <- string(b[:n]):
				continue
			case <-done:
				rerr <- nil
				return
			}
		}
		close(inpc)
	}()

	go func() {
		p, err := parseWithOpts(inpc, f.Name(), opts)
		if err != nil {
			close(done)
		}
		prog = p
		perr <- err
	}()

	err, err2 := <-rerr, <-perr
	if err == nil {
		err = err2
	}
	return prog, err
}

func parseWithOpts(inputs <-chan string, name string, opts []Option) (*Prog, error) {
	cf := makeConfig(opts)

	prog, pstats, err := parse(inputs, name, writers{cf.output, cf.logw})
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

	prog := newProg(name, writers{cf.output, cf.logw})
	err := prog.Load(r)
	if err == nil && cf.disasm {
		prog.disasm()
	}
	return prog, err
}

// Execute executes the Prog.
func Execute(prog *Prog, opts ...Option) (result []Block, binding Binding, err error) {
	cf := makeConfig(opts)

	result, binding, xstats, err := execute(prog, vmConfig{cf.trace})
	if cf.stats {
		printXStats(cf.output, xstats)
	}
	return result, binding, err
}

// Interpret parses and executes the BCL input.
func Interpret(input []byte, opts ...Option) ([]Block, Binding, error) {
	p, err := Parse(input, "input", opts...)
	if err != nil {
		return nil, nil, err
	}
	return Execute(p, opts...)
}

// InterpretFile reads, parses and executes the input from a BCL file.
// The file will be closed as soon as possible.
func InterpretFile(f FileInput, opts ...Option) ([]Block, Binding, error) {
	p, err := ParseFile(f, opts...)
	if err != nil {
		return nil, nil, err
	}
	return Execute(p, opts...)
}

// Bind binds the blocks selection (defined via binding) to the target.
// Target can be actually a struct or a slice of structs, and must correspond
// to a concrete Binding implementation, right now: [StructBinding] or [SliceBinding].
//
// When inside the struct (or the slice), the requirements are:
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
// If the binding type is a slice, and a slice pointed by target
// contained any elements, they are overwritten.
func Bind(target any, binding Binding) error {
	return copyBlocks(target, binding)
}

// Unmarshal interprets the BCL input and stores the blocks selected via 'bind' statement
// in the target.
// See [Bind] for details.
func Unmarshal(input []byte, target any, opts ...Option) error {
	_, binding, err := Interpret(input, opts...)
	if err != nil {
		return err
	}
	return Bind(target, binding)
}

// UnmarshalFile interprets the BCL file and stores the blocks selected via 'bind' statement
// in the target.
// See [Bind] for details.
func UnmarshalFile(f FileInput, target any, opts ...Option) error {
	_, binding, err := InterpretFile(f, opts...)
	if err != nil {
		return err
	}
	return Bind(target, binding)
}
