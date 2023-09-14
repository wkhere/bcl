package bcl

import (
	"fmt"
	"reflect"
	"strings"
)

// Unmarshal interprets the BCL input, and stores the result in dest,
// which should be a slice of structs.
// See [AppendBlocks] for a struct format.
func Unmarshal(input []byte, dest any) error {
	res, err := Interpret(input)
	if err != nil {
		return err
	}
	return AppendBlocks(dest, res)
}

// Block is a dynamic result of BCL interpretation. It can be put into
// a static structure via AppendBlocks.
type Block struct {
	Type, Name string
	Fields     map[string]any
}

// Interpret parses and evaluates the input, creating Blocks.
func Interpret(input []byte) ([]Block, error) {
	top, err := parse(input)
	if err != nil {
		return nil, err

	}
	return eval(&top)
}

// AppendBlocks adds the blocks to the dest, which needs to be a pointer
// to a slice of structs.
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
// case-insentitive.
//
// The lack of corresponding fields in the Go struct is reported as error.
// So is type mismatch of the fields.
func AppendBlocks(dest any, blocks []Block) error {
	destPtr := reflect.ValueOf(dest)
	if destPtr.Kind() != reflect.Pointer {
		return TypeErr("expected pointer to a slice of structs")
	}

	destSlice := destPtr.Elem()
	if destSlice.Kind() != reflect.Slice {
		return TypeErr("expected pointer to a slice of structs")
	}

	newSlice := reflect.MakeSlice(destSlice.Type(), len(blocks), len(blocks))
	for i, block := range blocks {
		err := copyBlock(newSlice.Index(i), &block)
		if err != nil {
			return err
		}
	}

	destSlice.Set(reflect.AppendSlice(destSlice, newSlice))
	return nil
}

func copyBlock(v reflect.Value, block *Block) error {
	t := v.Type()
	if st, bt := t.Name(), block.Type; !unsnakeEq(st, bt) {
		return StructErr(
			fmt.Sprintf("mismatch: struct type %s, block type %s", st, bt),
		)
	}

	tagged := map[string]int{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if tagv := f.Tag.Get("bcl"); tagv != "" {
			tagged[tagv] = i
		}
	}

	setField := func(name string, x any) error {
		var f reflect.StructField
		var ok bool
		if len(tagged) > 0 {
			var i int
			i, ok = tagged[name]
			if ok {
				f = t.Field(i)
			}
		}
		if !ok {
			f, ok = t.FieldByNameFunc(unsnakeMatcher(name))
		}
		if !ok {
			return StructErr(
				fmt.Sprintf("field mapping for %q not found in struct", name),
			)
		}
		if !f.IsExported() {
			return StructErr(fmt.Sprintf("found field %q but is unexported", f.Name))
		}

		vx := reflect.ValueOf(x)
		if st, bt := f.Type, vx.Type(); st != bt {
			return StructErr(
				fmt.Sprintf("type mismatch for the mapped field: struct.%s has %s, block.%s has %s",
					f.Name, st, name, bt,
				),
			)
		}
		namei := f.Index[0]
		v.Field(namei).Set(vx)
		return nil
	}

	err := setField("Name", block.Name)
	if err != nil {
		return err
	}
	for fkey, fval := range block.Fields {
		err = setField(fkey, fval)
		if err != nil {
			return err
		}
	}
	return nil
}

func unsnakeMatcher(snake string) func(string) bool {
	u := strings.ReplaceAll(snake, "_", "")
	return func(s string) bool {
		return strings.EqualFold(s, u)
	}
}

func unsnakeEq(orig, snake string) bool {
	return unsnakeMatcher(snake)(orig)
}

type TypeErr string
type StructErr string

func (e TypeErr) Error() string   { return string(e) }
func (e StructErr) Error() string { return string(e) }
