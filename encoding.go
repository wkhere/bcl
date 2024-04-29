package bcl

import "github.com/mohae/uvarint"

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
