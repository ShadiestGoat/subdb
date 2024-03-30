package ring

import (
	"shadygoat.eu/shitdb"
	"shadygoat.eu/shitdb/backends/all/lib"
)

type RingArrayBackend[IDType shitdb.IDConstraint] struct {
	real   lib.CommonArrayBackendUtil[IDType]
	maxLen int
}

func NewRingArrayBackend[IDType shitdb.IDConstraint](maxSize int, newestIsLargest bool) RingArrayBackend[IDType] {
	return RingArrayBackend[IDType]{
		real: lib.NewCommonArrayUtil[IDType](func() []shitdb.Group[IDType] {
			return make([]shitdb.Group[IDType], 0, maxSize)
		}, newestIsLargest),
		maxLen: maxSize,
	}
}

// Appends the ring's hooks at the end of the current hook list.
func (r *RingArrayBackend[IDType]) Register(h *shitdb.Hooks[IDType]) {
	h.Insert = append(h.Insert, r.Insert)
	r.real.Register(h)
}

func (r *RingArrayBackend[IDType]) Insert(groups ...shitdb.Group[IDType]) {
	if len(groups) == 0 {
		return
	}
	if len(groups) > r.maxLen {
		groups = groups[len(groups)-r.maxLen:]
	}

	r.real.Lock.Lock()
	defer r.real.Lock.Unlock()

	// maps are pointers, so for the non-full-replacement-mode, updating this var will update r.idCache
	ids := r.real.IDCache

	if len(groups) == r.maxLen {
		ids = make(map[IDType]shitdb.Group[IDType])
	}

	for _, g := range groups {
		ids[g.GetID()] = g
	}

	if len(groups) == r.maxLen {
		r.real.IDCache = ids
		r.real.Items = groups
		return
	}

	slStart := len(r.real.Items) + len(groups) - r.maxLen

	if slStart < 0 {
		slStart = 0
	}

	if slStart != 0 {
		for _, g := range r.real.Items[:slStart] {
			delete(r.real.IDCache, g.GetID())
		}
	}

	r.real.Items = append(r.real.Items[slStart:], groups...)
}

func (r *RingArrayBackend[IDType]) DeleteIDs(inp ...IDType) {
	r.real.DeleteID(inp...)
}

func (r *RingArrayBackend[IDType]) ReadIDs(inp ...IDType) []shitdb.Group[IDType] {
	return r.real.ReadID(inp...)
}

func (r *RingArrayBackend[IDType]) ReadQuery(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) ([]shitdb.Group[IDType], bool) {
	return r.real.ReadQuery(idPointer, oldToNew, f)
}

func (r *RingArrayBackend[IDType]) DeleteQuery(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) {
	r.real.DeleteQuery(idPointer, oldToNew, f)
}
