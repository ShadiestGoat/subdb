package ring

import (
	subdb "github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/backends/all/lib"
)

type RingArrayBackend[IDType subdb.IDConstraint] struct {
	real   lib.CommonArrayBackendUtil[IDType]
	maxLen int
}

func NewRingArrayBackend[IDType subdb.IDConstraint](maxSize int, newestIsLargest bool) RingArrayBackend[IDType] {
	return RingArrayBackend[IDType]{
		real: lib.NewCommonArrayUtil[IDType](func() []subdb.Group[IDType] {
			return make([]subdb.Group[IDType], 0, maxSize)
		}, newestIsLargest),
		maxLen: maxSize,
	}
}

// Appends the ring's hooks at the end of the current hook list.
func (r *RingArrayBackend[IDType]) Register(h *subdb.Hooks[IDType]) {
	h.Insert = append(h.Insert, r.Insert)
	r.real.Register(h)
}

func (r *RingArrayBackend[IDType]) Insert(groups ...subdb.Group[IDType]) {
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
		ids = make(map[IDType]subdb.Group[IDType])
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

func (r *RingArrayBackend[IDType]) ReadIDs(inp ...IDType) []subdb.Group[IDType] {
	return r.real.ReadID(inp...)
}

func (r *RingArrayBackend[IDType]) ReadQuery(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) ([]subdb.Group[IDType], bool) {
	return r.real.ReadQuery(idPointer, oldToNew, f)
}

func (r *RingArrayBackend[IDType]) DeleteQuery(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) {
	r.real.DeleteQuery(idPointer, oldToNew, f)
}
