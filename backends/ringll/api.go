package ringll

import "shadygoat.eu/shitdb"

func NewRing[IDType shitdb.IDConstraint](size int) {

}

func (r *RingBackend[IDType]) Register(h *shitdb.Hooks[IDType]) {
	h.Insert = append(h.Insert, r.Insert)
	h.DeleteID = append(h.DeleteID, r.DeleteID)
	h.DeleteQuery = append(h.DeleteQuery, r.DeleteWithFilter)
	h.ReadID = append(h.ReadID, r.ReadIDs)
	h.Read = append(h.Read, r.ReadWithFilter)
}

func (r *RingBackend[IDType]) DeleteID(ids ...IDType) {
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

func (r *RingBackend[IDType]) Insert(groups ...shitdb.Group[IDType]) {
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

func (r *RingBackend[IDType]) ReadIDs(ids ...IDType) []shitdb.Group[IDType] {
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

type nextFunc[IDType shitdb.IDConstraint] func (n *Node[IDType]) *Node[IDType]

func nextOldToNew[IDType shitdb.IDConstraint](n *Node[IDType]) *Node[IDType] {
	return n.Next
}
func nextNewToOld[IDType shitdb.IDConstraint](n *Node[IDType]) *Node[IDType] {
	return n.Prev
}

func (r *RingBackend[IDType]) DeleteWithFilter(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) {
	r.lock.RLock()
	defer r.lock.Unlock()

	r.queryFunc(idPointer, oldToNew, f, func(g shitdb.Group[IDType]) {
		r.delete(g.GetID())
	})
}

func (r *RingBackend[IDType]) ReadWithFilter(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType]) []shitdb.Group[IDType] {
	r.lock.RLock()
	defer r.lock.RUnlock()

	o := []shitdb.Group[IDType]{}

	r.queryFunc(idPointer, oldToNew, f, func(g shitdb.Group[IDType]) {
		o = append(o, g)
	})

	return o
}