package bcl

import (
	"fmt"
	"reflect"
	"strings"
)

func copyBlocks(target any, binding Binding) error {
	if binding == nil {
		return fmt.Errorf("no binding")
	}

	targetPtr := reflect.ValueOf(target)
	if targetPtr.Kind() != reflect.Pointer {
		return fmt.Errorf("expected pointer")
	}

	switch b := binding.(type) {

	case StructBinding:
		targetStruct := targetPtr.Elem()
		if targetStruct.Kind() != reflect.Struct {
			return fmt.Errorf("expected pointer to a struct")
		}

		return copyBlock(targetStruct, b.Value)

	case SliceBinding:
		targetSlice := targetPtr.Elem()
		if targetSlice.Kind() != reflect.Slice {
			return fmt.Errorf("expected pointer to a slice of structs")
		}

		blocks := b.Value
		newSlice := reflect.MakeSlice(targetSlice.Type(), len(blocks), len(blocks))
		for i, block := range blocks {
			err := copyBlock(newSlice.Index(i), block)
			if err != nil {
				return err
			}
		}

		targetSlice.Set(newSlice)

	default:
		return fmt.Errorf("unknown binding type %T", binding)
	}

	return nil
}

func copyBlock(v reflect.Value, block Block) error {
	t := v.Type()
	if st, bt := t.Name(), block.Type; st != "" && !unsnakeEq(st, bt) {
		return fmt.Errorf("mismatch: struct type %s, block type %s", st, bt)
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
			name, _, _ := strings.Cut(name, ".")
			f, ok = t.FieldByNameFunc(unsnakeMatcher(name))
		}
		if !ok {
			return fieldMappingErr(
				fmt.Sprintf("field mapping for %q not found in struct", name),
			)
		}
		if !f.IsExported() {
			return fmt.Errorf("found field %q but is unexported", f.Name)
		}

		namei := f.Index[0]
		vx := reflect.ValueOf(x)

		if vx.Type().AssignableTo(blockType) {
			return copyBlock(v.Field(namei), x.(Block))
		}

		if st, bt := f.Type, vx.Type(); !bt.AssignableTo(st) {
			return fmt.Errorf(
				"type mismatch for the mapped field: struct.%s has %s, block.%s has %s",
				f.Name, st, name, bt,
			)
		}

		v.Field(namei).Set(vx)
		return nil
	}

	err := setField("Name", block.Name)
	if err != nil {
		if _, ok := err.(fieldMappingErr); block.Name == "" && ok {
			goto fields
		}
		return err
	}
fields:
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

type fieldMappingErr string

func (e fieldMappingErr) Error() string { return string(e) }

var blockType = reflect.TypeOf(Block{})
