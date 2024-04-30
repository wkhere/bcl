package bcl

import (
	"math"
	"testing"
	"testing/quick"
)

var qcConf = &quick.Config{MaxCount: 1000}

func TestEncodingU16(t *testing.T) {
	f := func(x uint16) bool {
		var b [2]byte
		var p []byte = b[:]
		u16ToBytes(p, x)
		return x == u16FromBytes(p)
	}
	if err := quick.Check(f, qcConf); err != nil {
		t.Error(err)
	}
}

func TestI64U64(t *testing.T) {
	f := func(x int64) bool {
		return x == u64ToI64(i64ToU64(x))
	}
	if err := quick.Check(f, qcConf); err != nil {
		t.Error(err)
	}
}

func TestEncodingUvarint(t *testing.T) {
	f := func(x uint64) bool {
		var b [9]byte
		var p []byte = b[:]
		n := uvarintToBytes(p, x)
		y, m := uvarintFromBytes(p)
		return n == m && x == y
	}
	if err := quick.Check(f, qcConf); err != nil {
		t.Error(err)
	}
}

func TestEncodingVarint(t *testing.T) {
	f := func(x int64) bool {
		var b [9]byte
		var p []byte = b[:]
		n := varintToBytes(p, x)
		y, m := varintFromBytes(p)
		return n == m && x == y
	}
	if err := quick.Check(f, qcConf); err != nil {
		t.Error(err)
	}
}

func testEncodingValue(t *testing.T, x value) {
	t.Helper()

	var b [12]byte // for the longest string below
	var p []byte = b[:]

	n := valueToBytes(p, x)
	y, m := valueFromBytes(p)

	if n != m {
		t.Errorf("x=%v: size mismatch n=%d m=%d", x, n, m)
	}
	if x != y {
		t.Errorf("mismatch x=%v y=%v", x, y)
	}
}

func TestEncodingValuesTab(t *testing.T) {
	tab := []value{
		nil,
		0, 1, 2, -1, -2, 10, 127, 128, -127, -128,
		math.MaxInt64, math.MaxInt64 - 1, math.MinInt64, math.MinInt64 + 1,
		0.0, 0.5, -0.5, 1234.5, -1234.5,
		math.MaxFloat64, -math.MaxFloat64,
		true, false,
		"", "foo", "1234567890",
	}
	for _, v := range tab {
		testEncodingValue(t, v)
	}
}

func qcEncodingValue(x value) bool {
	var b [11]byte
	var p []byte = b[:]

	n := valueToBytes(p, x)
	y, m := valueFromBytes(p)

	return n == m && x == y
}

func qcEncodingString(s string) bool {
	p := make([]byte, 1+9+len(s))

	n := valueToBytes(p, s)
	y, m := valueFromBytes(p)

	return n == m && s == y
}

func TestEncodingValuesQC(t *testing.T) {
	var (
		fInt    = func(x int) bool { return qcEncodingValue(x) }
		fFloat  = func(x float64) bool { return qcEncodingValue(x) }
		fString = func(x string) bool { return qcEncodingString(x) }
	)
	for _, f := range []any{fInt, fFloat, fString} {
		if err := quick.Check(f, qcConf); err != nil {
			t.Error(err)
		}
	}
}
