package types

import (
	"encoding/binary"
)

func encodeInt[T int | uint](v T, size int) []byte {
	b := make([]byte, size)
	switch size {
	case 2:
		binary.LittleEndian.PutUint16(b, uint16(v))
	case 4:
		binary.LittleEndian.PutUint32(b, uint32(v))
	case 8, 0:
		binary.LittleEndian.PutUint64(b, uint64(v))
	}

	return b
}

func decodeInt[RT int | uint](size int, v []byte) RT {
	switch size {
	case 2:
		return RT(binary.LittleEndian.Uint16(v))
	case 4:
		return RT(binary.LittleEndian.Uint32(v))
	case 8, 0:
		return RT(binary.LittleEndian.Uint64(v))
	}

	return 0
}

type HelperValue[T any] struct {
	Value T
}

func (v HelperValue[T]) GetValue() any {
	return v.Value
}
