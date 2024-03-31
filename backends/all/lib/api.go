package lib

import (
	"slices"
	"sync"

	"github.com/shadiestgoat/subdb"
)

type NewArrayFunc[IDType subdb.IDConstraint] func() []subdb.Group[IDType]

// Library util, don't use as a real backend
type CommonArrayBackendUtil[IDType subdb.IDConstraint] struct {
	Lock            *sync.RWMutex
	Items           []subdb.Group[IDType]
	NewestIsLargest bool
	IDCache         map[IDType]subdb.Group[IDType]
	newArray        NewArrayFunc[IDType]
}

// Appends the ring's hooks at the end of the current hook list.
func (r *CommonArrayBackendUtil[IDType]) Register(h *subdb.Hooks[IDType]) {
	h.DeleteID = append(h.DeleteID, r.DeleteID)
	h.DeleteQuery = append(h.DeleteQuery, r.DeleteQuery)
	h.ReadID = append(h.ReadID, r.ReadID)
	h.Read = append(h.Read, r.ReadQuery)
}

func NewCommonArrayUtil[IDType subdb.IDConstraint](newArray NewArrayFunc[IDType], newestIsLargest bool) CommonArrayBackendUtil[IDType] {
	return CommonArrayBackendUtil[IDType]{
		Lock:            &sync.RWMutex{},
		Items:           newArray(),
		NewestIsLargest: newestIsLargest,
		IDCache:         map[IDType]subdb.Group[IDType]{},
		newArray:        newArray,
	}
}

func (r *CommonArrayBackendUtil[IDType]) DeleteID(inp ...IDType) {
	ids := make(map[IDType]bool, len(inp))

	r.Lock.RLock()

	for _, id := range inp {
		_, ok := r.IDCache[id]
		ids[id] = ok
	}

	r.Lock.RUnlock()
	r.Lock.Lock()

	newItems := r.newArray()

	for _, g := range r.Items {
		id := g.GetID()
		if !ids[id] {
			newItems = append(newItems, g)
		} else {
			delete(r.IDCache, id)
		}
	}

	r.Items = newItems

	r.Lock.Unlock()
}

func (r *CommonArrayBackendUtil[IDType]) ReadID(inp ...IDType) []subdb.Group[IDType] {
	r.Lock.RLock()
	defer r.Lock.RUnlock()

	o := make([]subdb.Group[IDType], 0, len(inp))

	for _, id := range inp {
		g := r.IDCache[id]
		if g != nil {
			o = append(o, g)
		}
	}

	return o
}

func (r *CommonArrayBackendUtil[IDType]) ReadQuery(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) ([]subdb.Group[IDType], bool) {
	r.Lock.RLock()
	defer r.Lock.RLock()

	o := []subdb.Group[IDType]{}

	exitEarly := r.queryFunc(idPointer, oldToNew, f, func(g subdb.Group[IDType], _ int) {
		o = append(o, g)
	})

	return o, exitEarly
}

func (r *CommonArrayBackendUtil[IDType]) DeleteQuery(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) {
	r.Lock.Lock()
	defer r.Lock.Lock()

	badI := []int{}

	r.queryFunc(idPointer, oldToNew, f, func(_ subdb.Group[IDType], i int) {
		badI = append(badI, i)
	})

	newItems := r.newArray()

	lastI := 0

	slices.Sort(badI)

	for _, i := range badI {
		delete(r.IDCache, r.Items[i].GetID())
		newItems = append(newItems, r.Items[lastI:i]...)
		lastI = i + 1
	}

	r.Items = append(newItems, r.Items[lastI:]...)
}
