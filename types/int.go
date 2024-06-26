// AUTOGENERATED, DO NOT EDIT.
package types

import (
	"github.com/shadiestgoat/subdb"
)

// Util for int-based fields (uint & int)
type intBase[T int | uint] struct {
	HelperValue[T]
	// The size of the int in bytes. Valid values are 2 (uint16), 4 (uint32), 8 (uint64)
	Size int
}

func (i intBase[T]) Encode() []byte {
	return encodeInt[T](i.Value, i.Size)
}

func (i *intBase[T]) Load(v []byte) {
	i.Value = decodeInt[T](i.Size, v)
}

func (i intBase[T]) New() subdb.Field {
	return &intBase[T]{
		Size: i.Size,
	}
}

func (i intBase[T]) StaticSize() int {
	return i.Size
}

type Int struct {
	intBase[int]
}

type Uint struct {
	intBase[uint]
}

func newIntBase[T int | uint](v T, size int) intBase[T] {
	return intBase[T]{
		HelperValue: HelperValue[T]{
			Value: v,
		},
		Size: size,
	}
}

func NewUint(v uint, size int) *Uint {
	return &Uint{
		intBase: newIntBase(v, size),
	}
}

func NewInt(v int, size int) *Int {
	return &Int{
		intBase: newIntBase(v, size),
	}
}
