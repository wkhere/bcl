package bcl

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

func copyBlocksToPtr(target any, binding Binding) error {
	if binding == nil {
		return fmt.Errorf("no binding")
	}

	targetPtr := reflect.ValueOf(target)
	if k := targetPtr.Kind(); k != reflect.Pointer {
		return fmt.Errorf("expected pointer, have: %s", k)
	}

	err := copyBlocks(targetPtr.Elem(), binding)
	if err != nil {
		return err
	}
	return nil
}

func copyBlocks(targetValue reflect.Value, binding Binding) error {
	switch b := binding.(type) {

	case StructBinding:
		targetStruct := targetValue
		if k := targetStruct.Kind(); k != reflect.Struct {
			return fmt.Errorf("expected struct, have: %s", k)
		}

		return copyBlock(targetStruct, b.Value)

	case SliceBinding:
		targetSlice := targetValue
		if k := targetSlice.Kind(); k != reflect.Slice {
			return fmt.Errorf("expected slice, have: %s", k)
		}
		if k := targetSlice.Type().Elem().Kind(); k != reflect.Struct {
			return fmt.Errorf("slice element deref: expected struct, have: %s", k)
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

	case *UmbrellaBinding:
		targetUmbrella := targetValue
		if k := targetUmbrella.Kind(); k != reflect.Struct {
			return fmt.Errorf("expected umbrella struct, have: %s", k)
		}
		t := targetUmbrella.Type()
		for i, part := range b.Parts {
			err := copyBlocks(targetUmbrella.Field(i), part)
			if err != nil {
				return fmt.Errorf("%s: %w", t.Field(i).Name, err)
			}
		}

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

		if vx.Type().AssignableTo(reflect.TypeOf(Block{})) {
			return copyBlock(v.Field(namei), x.(Block))
		}

		if st, bt := f.Type, vx.Type(); !bt.AssignableTo(st) {
			// todo: implicit conversion of a BCL int to Go float64
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

type fieldMappingErr string

func (e fieldMappingErr) Error() string { return string(e) }

func unsnakeMatcher(snake string) func(string) bool {
	u := strings.ReplaceAll(snake, "_", "")
	return func(s string) bool {
		return strings.EqualFold(s, u)
	}
}

func unsnakeEq(orig, snake string) bool {
	return unsnakeMatcher(snake)(orig)
}

func snake(input string) string {
	nextRune := func(idx int) rune { r, _ := utf8.DecodeRuneInString(input[idx:]); return r }

	var b strings.Builder
	var prevUpper bool

	for i, v := range input {
		upper := unicode.IsUpper(v)
		if upper {
			if i > 0 && (!prevUpper ||
				unicode.IsLower(nextRune(i+utf8.RuneLen(v)))) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(v))
		} else {
			b.WriteRune(v)
		}
		prevUpper = upper
	}
	return b.String()
}
