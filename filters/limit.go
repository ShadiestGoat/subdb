package filters

import "github.com/shadiestgoat/subdb"

type Limit[IDType subdb.IDConstraint] struct {
	// The amount of matches left
	// 0 - no more matches
	Left int
}

func (v *Limit[IDType]) Match(subdb.Group[IDType]) (bool, bool) {
	v.Left--
	return v.Left >= 0, v.Left <= 0
}

func (f Limit[IDType]) Copy() subdb.Filter[IDType] {
	return &Limit[IDType]{
		Left: f.Left,
	}
}
