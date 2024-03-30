package all

import (
	"shadygoat.eu/shitdb"
	"shadygoat.eu/shitdb/backends/all/lib"
)

// A backend that stores all the data
type AllBackend[IDType shitdb.IDConstraint] struct {
	real lib.CommonArrayBackendUtil[IDType]
}

func NewAllBackend[IDType shitdb.IDConstraint](newestIsLargest bool) AllBackend[IDType] {
	return AllBackend[IDType]{
		real: lib.NewCommonArrayUtil[IDType](func() []shitdb.Group[IDType] {
			return make([]shitdb.Group[IDType], 0)
		}, newestIsLargest),
	}
}

// Appends the ring's hooks at the end of the current hook list.
func (r *AllBackend[IDType]) Register(h *shitdb.Hooks[IDType]) {
	h.Insert = append(h.Insert, r.Insert)
	r.real.Register(h)
}

func (r *AllBackend[IDType]) Insert(groups ...shitdb.Group[IDType]) {
	if len(groups) == 0 {
		return
	}
	
	r.real.Lock.Lock()
	defer r.real.Lock.Unlock()

	r.real.Items = append(r.real.Items, groups...)
}

func (r *AllBackend[IDType]) DeleteIDs(inp ...IDType) {
	r.real.DeleteID(inp...)
}

func (r *AllBackend[IDType]) ReadIDs(inp ...IDType) []shitdb.Group[IDType] {
	return r.real.ReadID(inp...)
}

func (r *AllBackend[IDType]) ReadQuery(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) ([]shitdb.Group[IDType], bool) {
	return r.real.ReadQuery(idPointer, oldToNew, f)
}

func (r *AllBackend[IDType]) DeleteQuery(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) {
	r.real.DeleteQuery(idPointer, oldToNew, f)
}

// Resets all the items & returns them
func (r *AllBackend[IDType]) Reset() []shitdb.Group[IDType] {
	r.real.Lock.Lock()
	defer r.real.Lock.Unlock()

	curItems := r.real.Items
	r.real.Items = make([]shitdb.Group[IDType], 0)
	r.real.IDCache = make(map[IDType]shitdb.Group[IDType])

	return curItems
}