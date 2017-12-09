package cc111x

// Marshaling of ints in little-endian order.

func marshalUint16(n uint16) []byte {
	return []byte{byte(n & 0xFF), byte(n >> 8)}
}

func marshalUint32(n uint32) []byte {
	return append(marshalUint16(uint16(n&0xFFFF)), marshalUint16(uint16(n>>16))...)
}
