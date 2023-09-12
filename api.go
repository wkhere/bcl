package bcl

import (
	"fmt"
	"reflect"
	"strings"
)

type Block struct {
	Type, Name string
	Fields     map[string]any
}

func InterpAndSave(dest any, input []byte) error {
	res, err := Interp(input)
	if err != nil {
		return err
	}
	return Save(dest, res)
}

func Interp(input []byte) ([]Block, error) {
	top, err := parse(input)
	if err != nil {
		return nil, err

	}
	return eval(&top)
}

// Save saves the blocks into the dest, which needs to be a pointer
// to a slice of structs.
func Save(dest any, blocks []Block) error {
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
		err := save1(newSlice.Index(i), &block)
		if err != nil {
			return err
		}
	}

	destSlice.Set(reflect.AppendSlice(destSlice, newSlice))
	return nil
}

func save1(v reflect.Value, block *Block) error {
	t := v.Type()
	if st, bt := t.Name(), block.Type; !snakeEq(st, bt) {
		return StructErr(
			fmt.Sprintf("mismatch: struct type %s, block type %s", st, bt),
		)
	}

	setField := func(name string, x any) error {
		f, ok := t.FieldByNameFunc(snakeMatcher(name))
		if !ok {
			return StructErr(
				fmt.Sprintf(
					"field mapping for %q not found in struct", name,
				),
			)
		}
		vx := reflect.ValueOf(x)
		if st, bt := f.Type, vx.Type(); st != bt {
			return StructErr(
				fmt.Sprintf(
					"type mismatch for the mapped field: "+
						"struct.%s has %s, block.%s has %s",
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

func snakeMatcher(snake string) func(string) bool {
	u := strings.ReplaceAll(snake, "_", "")
	return func(s string) bool {
		return strings.EqualFold(s, u)
	}
}

func snakeEq(orig, snake string) bool {
	return snakeMatcher(snake)(orig)
}

type TypeErr string
type StructErr string

func (e TypeErr) Error() string   { return string(e) }
func (e StructErr) Error() string { return string(e) }
