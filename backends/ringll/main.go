package ringll

import (
	"sync"

	"shadygoat.eu/shitdb"
)

type Node[IDType shitdb.IDConstraint] struct {
	Value shitdb.Group[IDType]
	Next, Prev *Node[IDType]
}

type RingBackend[IDType shitdb.IDConstraint] struct {
	idCache map[IDType]*Node[IDType]
	size int
	maxSize int
	lock *sync.RWMutex

	newest, oldest *Node[IDType]
}

func (r *RingBackend[IDType]) delete(id IDType) bool {
	n := r.idCache[id]
	
	if n == nil {
		return false
	}

	delete(r.idCache, id)

	if n.Next == nil {
		r.newest = n.Prev
		n.Prev.Next = nil
		return true
	}
	if n.Prev == nil {
		r.oldest = n.Next
		n.Next.Prev = nil
		return true
	}

	n.Prev.Next = n.Next
	n.Next.Prev = n.Prev
	
	return true
}

func (r *RingBackend[IDType]) queryFunc(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType], action func (g shitdb.Group[IDType])) {
	var n *Node[IDType]
	// Change n to the next value
	var next nextFunc[IDType]

	if oldToNew {
		n = r.oldest
		next = nextOldToNew[IDType]
	} else {
		n = r.newest
		next = nextNewToOld[IDType]
	}

	if idPointer != nil && r.newest != r.oldest {
		id := idPointer.ID

		n = r.newest
		idpNext := nextNewToOld[IDType]

		if idPointer.Hint == shitdb.LOCATION_HINT_OLDEST {
			n = r.oldest
			idpNext = nextOldToNew[IDType]
		}

		for {
			if n == nil {
				// we couldn't find the starting id, so quit early
				return
			}
			if n.Value.GetID() == id {
				// Gotcha
				break
			}
			n = idpNext(n)
		}
	}

	for {
		if n == nil {
			break
		}
		if f == nil {
			action(n.Value)
		} else {
			ok, exitEarly := f.Match(n.Value)
			if ok {
				action(n.Value)
			}
			if exitEarly {
				return
			}
		}
		n = next(n)
	}
}