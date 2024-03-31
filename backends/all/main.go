package all

import (
	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/backends/all/lib"
)

// A backend that stores all the data
type AllBackend[IDType subdb.IDConstraint] struct {
	real lib.CommonArrayBackendUtil[IDType]
}

func NewAllBackend[IDType subdb.IDConstraint](newestIsLargest bool) AllBackend[IDType] {
	return AllBackend[IDType]{
		real: lib.NewCommonArrayUtil[IDType](func() []subdb.Group[IDType] {
			return make([]subdb.Group[IDType], 0)
		}, newestIsLargest),
	}
}

// Appends the ring's hooks at the end of the current hook list.
func (r *AllBackend[IDType]) Register(h *subdb.Hooks[IDType]) {
	h.Insert = append(h.Insert, r.Insert)
	r.real.Register(h)
}

func (r *AllBackend[IDType]) Insert(groups ...subdb.Group[IDType]) {
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

func (r *AllBackend[IDType]) ReadIDs(inp ...IDType) []subdb.Group[IDType] {
	return r.real.ReadID(inp...)
}

func (r *AllBackend[IDType]) ReadQuery(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) ([]subdb.Group[IDType], bool) {
	return r.real.ReadQuery(idPointer, oldToNew, f)
}

func (r *AllBackend[IDType]) DeleteQuery(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) {
	r.real.DeleteQuery(idPointer, oldToNew, f)
}

// Resets all the items & returns them
func (r *AllBackend[IDType]) Reset() []subdb.Group[IDType] {
	r.real.Lock.Lock()
	defer r.real.Lock.Unlock()

	curItems := r.real.Items
	r.real.Items = make([]subdb.Group[IDType], 0)
	r.real.IDCache = make(map[IDType]subdb.Group[IDType])

	return curItems
}