package ring

import (
	"slices"
	"sort"
	"sync"

	"shadygoat.eu/shitdb"
)

type RingArrayBackend[IDType shitdb.IDConstraint] struct {
	lock            *sync.RWMutex
	maxLen          int
	items           []shitdb.Group[IDType]
	newestIsLargest bool
	idCache         map[IDType]shitdb.Group[IDType]
}

func (r *RingArrayBackend[IDType]) Insert(groups ...shitdb.Group[IDType]) {
	if len(groups) == 0 {
		return
	}
	if len(groups) > r.maxLen {
		groups = groups[len(groups)-r.maxLen:]
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	// maps are pointers, so for the non-full-replacement-mode, updating this var will update r.idCache
	ids := r.idCache

	if len(groups) == r.maxLen {
		ids = make(map[IDType]shitdb.Group[IDType])
	}

	for _, g := range groups {
		ids[g.GetID()] = g
	}

	if len(groups) == r.maxLen {
		r.idCache = ids
		r.items = groups
		return
	}

	slStart := len(r.items) + len(groups) - r.maxLen

	if slStart < 0 {
		slStart = 0
	}

	if slStart != 0 {
		for _, g := range r.items[:slStart] {
			delete(r.idCache, g.GetID())
		}
	}

	r.items = append(r.items[slStart:], groups...)
}

func (r *RingArrayBackend[IDType]) DeleteID(inp ...IDType) {
	ids := make(map[IDType]bool, len(inp))

	r.lock.RLock()

	for _, id := range inp {
		_, ok := r.idCache[id]
		ids[id] = ok
	}

	r.lock.RUnlock()
	r.lock.Lock()

	newItems := make([]shitdb.Group[IDType], 0, r.maxLen)

	for _, g := range r.items {
		id := g.GetID()
		if !ids[id] {
			newItems = append(newItems, g)
		} else {
			delete(r.idCache, id)
		}
	}

	r.items = newItems

	r.lock.Unlock()
}

func (r *RingArrayBackend[IDType]) ReadID(inp ...IDType) []shitdb.Group[IDType] {
	r.lock.RLock()
	defer r.lock.RUnlock()

	o := make([]shitdb.Group[IDType], 0, len(inp))

	for _, id := range inp {
		g := r.idCache[id]
		if g != nil {
			o = append(o, g)
		}
	}

	return o
}

func (r *RingArrayBackend[IDType]) Read(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) []shitdb.Group[IDType] {
	r.lock.RLock()
	defer r.lock.RLock()

	o := []shitdb.Group[IDType]{}

	r.queryFunc(idPointer, oldToNew, f, func(g shitdb.Group[IDType], _ int) {
		o = append(o, g)
	})

	return o
}

func (r *RingArrayBackend[IDType]) DeleteQuery(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) {
	r.lock.Lock()
	defer r.lock.Lock()

	badI := []int{}

	r.queryFunc(idPointer, oldToNew, f, func(_ shitdb.Group[IDType], i int) {
		badI = append(badI, i)
	})

	newItems := make([]shitdb.Group[IDType], 0, r.maxLen)

	lastI := 0

	slices.Sort(badI)

	for _, i := range badI {
		delete(r.idCache, r.items[i].GetID())
		newItems = append(newItems, r.items[lastI:i]...)
		lastI = i
	}

	r.items = append(newItems, r.items[lastI:]...)
}
