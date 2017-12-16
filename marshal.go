package cc111x

// Marshaling of ints in big-endian order.

func marshalUint16(n uint16) []byte {
	return []byte{byte(n >> 8), byte(n & 0xFF)}
}

func marshalUint32(n uint32) []byte {
	return append(marshalUint16(uint16(n>>16)), marshalUint16(uint16(n&0xFFFF))...)
}
