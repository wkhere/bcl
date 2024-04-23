package bcl

func u16ToBytes(p []byte, x uint16) {
	p[0] = byte(x & 0xff)
	p[1] = byte(x >> 8 & 0xff)

}

func u16FromBytes(b []byte) uint16 {
	return uint16(b[1])<<8 | uint16(b[0])
}
