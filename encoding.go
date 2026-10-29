package bcl

import (
	"bufio"
	stdbinary "encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/wkhere/uvarint"
)

func u16ToBytes(p []byte, x uint16) {
	p[0] = byte(x >> 8)
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
	p, err := r.Peek(1)
	if err != nil {
		return 0, partialEOF(err)
	}
	p, err = r.Peek(uvarint.UvarintLen(p))
	if err != nil {
		return 0, partialEOF(err)
	}
	x, n := uvarintFromBytes(p)
	_, err = r.Discard(n)
	return x, err
}

func stringFromBuf(r *bufio.Reader) (string, error) {
	k, err := uvarintFromBuf(r)
	if err != nil {
		return "", err
	}
	if k > maxStringLen {
		return "", fmt.Errorf("string too long: %d bytes, max %d", k, maxStringLen)
	}
	p := make([]byte, k)
	_, err = readPart(r, p)
	return string(p), err
}

func readPart(r io.Reader, p []byte) (n int, err error) {
	n, err = io.ReadFull(r, p)
	return n, partialEOF(err)
}

func partialEOF(err error) error {
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return err
}

func varintToBytes(p []byte, x int64) int {
	return uvarintToBytes(p, i64ToU64(x))
}

func varintFromBytes(p []byte) (int64, int) {
	u, n := uvarintFromBytes(p)
	return u64ToI64(u), n
}

func varintFromBuf(r *bufio.Reader) (int64, error) {
	u, err := uvarintFromBuf(r)
	return u64ToI64(u), err
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
	switch x := v.(type) {
	case int:
		p[0] = byte(typeINT)
		n = 1 + varintToBytes(p[1:], int64(x))

	case float64:
		p[0] = byte(typeFLOAT)
		stdbinary.BigEndian.PutUint64(p[1:], math.Float64bits(x))
		n = 1 + 8

	case string:
		p[0] = byte(typeSTR)
		p = p[1:]
		i := uvarintToBytes(p, uint64(len(x)))
		p = p[i:]
		if len(p) < len(x) {
			n = 1 + i
			panic("no space")
		}
		copy(p, []byte(x))
		n = 1 + i + len(x)

	case bool:
		p[0] = byte(typeBOOL)
		if x {
			p[1] = 1
		} else {
			p[1] = 0
		}
		n = 1 + 1

	default:
		if v == nil {
			p[0] = byte(typeNIL)
			return 1 + 0 // don't emit nil value, just the type
		}
		panic(errInvalidValue{v})
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
	_, err := readPart(r, b[:1])
	if err != nil {
		return nil, err
	}

	switch c := typecode(b[0]); c {
	case typeINT:
		x, err := varintFromBuf(r)
		return int(x), err

	case typeFLOAT:
		var bf [8]byte
		_, err = readPart(r, bf[:])
		return math.Float64frombits(stdbinary.BigEndian.Uint64(bf[:])), err

	case typeSTR:
		return stringFromBuf(r)

	case typeBOOL:
		_, err = readPart(r, b[:1])
		return b[0] != 0, err

	case typeNIL:
		return nil, nil

	default:
		panic(errInvalidType{b[0]})
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
