package bcl

import (
	"bufio"
	stdbinary "encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/mohae/uvarint"
)

func u16ToBytes(p []byte, x uint16) {
	p[0] = byte(x >> 8 & 0xff)
	p[1] = byte(x & 0xff)
}

func u16FromBytes(b []byte) uint16 {
	return uint16(b[0])<<8 | uint16(b[1])
}

func uvarintToBytes(p []byte, x uint64) int {
	return uvarint.Encode(p, x)
}

func uvarintFromBytes(p []byte) (uint64, int) {
	return uvarint.Decode(p)
}

func uvarintFromBuf(r *bufio.Reader) (uint64, error) {
	p, err := r.Peek(9)
	if err != nil && err != io.EOF {
		return 0, err
	}
	x, n := uvarintFromBytes(p)
	_, err = r.Discard(n)
	return x, err
}

func varintToBytes(p []byte, x int64) int {
	return uvarintToBytes(p, i64ToU64(x))
}

func varintFromBytes(p []byte) (int64, int) {
	u, n := uvarintFromBytes(p)
	return u64ToI64(u), n
}

func i64ToU64(x int64) uint64 {
	if x < 0 {
		return uint64(-x) ^ 0xffffffffffffffff + 1
	}
	return uint64(x)
}

func u64ToI64(x uint64) int64 {
	if x > math.MaxInt64 {
		return -(int64(x^0xffffffffffffffff) + 1)
	}
	return int64(x)
}

func valueToBytes(p []byte, v value) (n int) {
	p[0] = byte(typecodeOf(v))
	p = p[1:] // local copy
	n++

	switch x := v.(type) {
	case int:
		n += varintToBytes(p, int64(x))

	case float64:
		stdbinary.BigEndian.PutUint64(p, math.Float64bits(x))
		n += 8

	case string:
		i := uvarintToBytes(p, uint64(len(x)))
		p = p[i:]
		if len(p) < len(x) {
			n += i
			panic("no space")
		}
		copy(p, []byte(x))
		n += i + len(x)

	case bool:
		if x {
			p[0] = 1
		} else {
			p[0] = 0
		}
		n++

	default:
		// whether it's nil or invalid type, don't emit the value
	}
	return n
}

func valueFromBytes(p []byte) (value, int) {
	switch c, p := typecode(p[0]), p[1:]; c {
	case typeINT:
		x, n := varintFromBytes(p)
		return int(x), 1 + n

	case typeFLOAT:
		return math.Float64frombits(stdbinary.BigEndian.Uint64(p)), 1 + 8

	case typeSTR:
		k, i := uvarintFromBytes(p)
		return string(p[i : i+int(k)]), 1 + i + int(k)

	case typeBOOL:
		return p[0] != 0, 1 + 1

	case typeNIL:
		return nil, 1

	default:
		panic(errInvalidType{p[0]})
	}
}

func valueFromBuf(r *bufio.Reader) (value, error) {
	var b [1]byte
	_, err := io.ReadFull(r, b[:1])
	if err != nil {
		return nil, err
	}

	switch c := typecode(b[0]); c {
	case typeINT:
		p, _ := r.Peek(9)
		x, i := varintFromBytes(p)
		_, err = r.Discard(i)
		return int(x), err

	case typeFLOAT:
		p, _ := r.Peek(8)
		_, err = r.Discard(len(p))
		return math.Float64frombits(stdbinary.BigEndian.Uint64(p)), err

	case typeSTR:
		p, _ := r.Peek(9)
		k, i := uvarintFromBytes(p)
		r.Discard(i)
		p, _ = r.Peek(int(k))
		_, err = r.Discard(len(p))
		return string(p), err

	case typeBOOL:
		_, err = io.ReadFull(r, b[:1])
		return b[0] != 0, err

	case typeNIL:
		return nil, nil

	default:
		panic(errInvalidType{b[0]})
	}
}

func typecodeOf(v value) typecode {
	switch v.(type) {
	case int:
		return typeINT
	case float64:
		return typeFLOAT
	case string:
		return typeSTR
	case bool:
		return typeBOOL
	default:
		if v == nil {
			return typeNIL
		}
		panic(errInvalidValue{v})
	}
}

type errInvalidType struct{ byte }
type errInvalidValue struct{ value }

func (e errInvalidType) Error() string {
	return fmt.Sprintf("invalid type: %q", e.byte)
}

func (e errInvalidValue) Error() string {
	return fmt.Sprintf("invalid value: %v", e.value)
}
