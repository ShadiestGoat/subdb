package ringll

import (
	"sync"

	"shadygoat.eu/shitdb"
)

// Creates a new linked list ring backend
func NewRing[IDType shitdb.IDConstraint](size int) *RingLinkedListBackend[IDType] {
	return &RingLinkedListBackend[IDType]{
		idCache: map[IDType]*Node[IDType]{},
		size:    0,
		maxSize: size,
		lock:    &sync.RWMutex{},
		newest:  nil,
		oldest:  nil,
	}
}

// Appends the ring's hooks at the end of the current hook list.
func (r *RingLinkedListBackend[IDType]) Register(h *shitdb.Hooks[IDType]) {
	h.Insert = append(h.Insert, r.Insert)
	h.DeleteID = append(h.DeleteID, r.DeleteID)
	h.DeleteQuery = append(h.DeleteQuery, r.DeleteWithFilter)
	h.ReadID = append(h.ReadID, r.ReadIDs)
	h.Read = append(h.Read, r.ReadWithFilter)
}

func (r *RingLinkedListBackend[IDType]) DeleteID(ids ...IDType) {
	r.lock.Lock()
	defer r.lock.Unlock()

	deletedIDs := 0

	for _, id := range ids {
		if r.delete(id) {
			deletedIDs++
		}
	}

	r.size -= deletedIDs
}

func (r *RingLinkedListBackend[IDType]) Insert(groups ...shitdb.Group[IDType]) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, g := range groups {
		n := &Node[IDType]{
			Value: g,
			Next:  nil,
			Prev:  r.newest,
		}
		if r.oldest == nil {
			r.oldest = n
		} else {
			r.newest.Next = n
		}

		r.newest = n
		r.size++
		r.idCache[g.GetID()] = n
	}

	for r.size > r.maxSize {
		r.delete(r.oldest.Value.GetID())
		r.size--
	}
}

func (r *RingLinkedListBackend[IDType]) ReadIDs(ids ...IDType) []shitdb.Group[IDType] {
	r.lock.RLock()
	defer r.lock.RUnlock()

	o := []shitdb.Group[IDType]{}

	for _, id := range ids {
		if n := r.idCache[id]; n != nil {
			o = append(o, n.Value)
		}
	}

	return o
}

type nextFunc[IDType shitdb.IDConstraint] func(n *Node[IDType]) *Node[IDType]

func nextOldToNew[IDType shitdb.IDConstraint](n *Node[IDType]) *Node[IDType] {
	return n.Next
}
func nextNewToOld[IDType shitdb.IDConstraint](n *Node[IDType]) *Node[IDType] {
	return n.Prev
}

func (r *RingLinkedListBackend[IDType]) DeleteWithFilter(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) {
	r.lock.RLock()
	defer r.lock.Unlock()

	r.queryFunc(idPointer, oldToNew, f, func(g shitdb.Group[IDType]) {
		r.delete(g.GetID())
	})
}

func (r *RingLinkedListBackend[IDType]) ReadWithFilter(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) []shitdb.Group[IDType] {
	r.lock.RLock()
	defer r.lock.RUnlock()

	o := []shitdb.Group[IDType]{}

	r.queryFunc(idPointer, oldToNew, f, func(g shitdb.Group[IDType]) {
		o = append(o, g)
	})

	return o
}
