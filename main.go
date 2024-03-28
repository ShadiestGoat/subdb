package shitdb

import (
	"golang.org/x/exp/constraints"
)

type Field interface {
	Encode() []byte
	Load(v []byte)
	New() Field
	GetValue() any
}

type FieldDynamicSize interface {
	DynamicSize(v []byte) int
	// Gets the amount of bytes needed to generate a full dynamic size. Should be static
	DynamicSizeSize() int
}

type FieldStaticSize interface {
	StaticSize() int
}

type IDConstraint = constraints.Ordered

type Group[IDType IDConstraint] interface {
	GetID() IDType
	Load([]Field)
	Store() []Field
}

type Filter[IDType IDConstraint] interface {
	// Return true if matches
	Match(g Group[IDType]) (ok bool, returnEarly bool)
}

type InsertFunc[IDType IDConstraint] func(groups ...Group[IDType])
type DeleteIDFunc[IDType IDConstraint] func(ids ...IDType)

type LocationHint int

const (
	LOCATION_HINT_UNSPECIFIED LocationHint = iota
	LOCATION_HINT_OLDEST
	LOCATION_HINT_NEWEST
)

type ApproximationBehavior int

const (
	// If the backend doesn't have the idp, quit
	APPROXIMATE_QUIT_EARLY ApproximationBehavior = iota
	APPROXIMATE_OLDEST
	APPROXIMATE_NEWEST
)

type IDPointer[IDType IDConstraint] struct {
	ID   IDType
	Hint LocationHint
	// If the id is not in a backend but there are surrounding ids, how to approximate
	// Eg. the pointer is 3, and db layout is this:
	// oldest -> 1, 2, 4, 5 <- newest
	// If APPROXIMATE_OLDEST, idp will approximate to 2. If APPROXIMATE_NEWEST, idp will be 4.
	ApproximationBehavior ApproximationBehavior
}

// idPointer is a pointer to the starting id at which to apply f. If the pointer is nil, it will be ignored.
// oldToNew indicated which direction to read. If the data is believed to be recent, it is faster to query for new to old (ie. false), though this does not matter if the filter is limitless & the id pointer is nil
// oldToNew also affects the idPointer. It decides which direction to go after the id is met.
type GeneralQueryFunc[IDType IDConstraint] func(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType])

// See GeneralQueryFunc
type DeleteQueryFunc[IDType IDConstraint] GeneralQueryFunc[IDType]

// See GeneralQueryFunc
type ReadFunc[IDType IDConstraint] func(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) []Group[IDType]

type ReadIDFunc[IDType IDConstraint] func(ids ...IDType) []Group[IDType]

type Hooks[IDType IDConstraint] struct {
	Insert      []InsertFunc[IDType]
	DeleteID    []DeleteIDFunc[IDType]
	DeleteQuery []DeleteQueryFunc[IDType]
	Read        []ReadFunc[IDType]
	ReadID      []ReadIDFunc[IDType]
}

type Domain[IDType IDConstraint] struct {
	Fields []Field
	Hooks  Hooks[IDType]
}

// domain -> group -> field
