package types

import "shadygoat.eu/shitdb"

type String struct {
	HelperValue[string]

	// The amount of bytes to allocate for the size of the field size. This also determined the max size of the string you are storing.
	// Valid values are 2 (65,535), 4 (4,294,967,295), 8 (18,446,744,073,709,551,615)
	Size int
}

func (s String) Encode() []byte {
	return append(encodeInt(len(s.Value), s.Size), []byte(s.Value)...)
}

func (s *String) Load(v []byte) {
	s.Value = string(v[s.Size:])
}

func (s String) New() shitdb.Field {
	return &String{
		Size: s.Size,
	}
}

func (s String) DynamicSize(v []byte) int {
	return decodeInt[int](s.Size, v) + s.Size
}

func (s String) DynamicSizeSize() int {
	return s.Size
}
