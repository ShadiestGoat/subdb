package ringll

import (
	"sync"

	subdb "github.com/shadiestgoat/subdb"
)

// Creates a new linked list ring backend
func NewRing[IDType subdb.IDConstraint](size int) *RingLinkedListBackend[IDType] {
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
func (r *RingLinkedListBackend[IDType]) Register(h *subdb.Hooks[IDType]) {
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

func (r *RingLinkedListBackend[IDType]) Insert(groups ...subdb.Group[IDType]) {
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

func (r *RingLinkedListBackend[IDType]) ReadIDs(ids ...IDType) []subdb.Group[IDType] {
	r.lock.RLock()
	defer r.lock.RUnlock()

	o := []subdb.Group[IDType]{}

	for _, id := range ids {
		if n := r.idCache[id]; n != nil {
			o = append(o, n.Value)
		}
	}

	return o
}

type nextFunc[IDType subdb.IDConstraint] func(n *Node[IDType]) *Node[IDType]

func nextOldToNew[IDType subdb.IDConstraint](n *Node[IDType]) *Node[IDType] {
	return n.Next
}
func nextNewToOld[IDType subdb.IDConstraint](n *Node[IDType]) *Node[IDType] {
	return n.Prev
}

func (r *RingLinkedListBackend[IDType]) DeleteWithFilter(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) {
	r.lock.RLock()
	defer r.lock.Unlock()

	r.queryFunc(idPointer, oldToNew, f, func(g subdb.Group[IDType]) {
		r.delete(g.GetID())
	})
}

func (r *RingLinkedListBackend[IDType]) ReadWithFilter(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) ([]subdb.Group[IDType], bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	o := []subdb.Group[IDType]{}

	exitEarly := r.queryFunc(idPointer, oldToNew, f, func(g subdb.Group[IDType]) {
		o = append(o, g)
	})

	return o, exitEarly
}
