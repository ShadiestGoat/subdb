package ringll

import (
	"sync"

	"shadygoat.eu/shitdb"
)

type Node[IDType shitdb.IDConstraint] struct {
	Value      shitdb.Group[IDType]
	Next, Prev *Node[IDType]
}

// A linked list ring backend. Works in a similar way to a regular ring backend, but instead of an array, uses a linked list
// Comparison table between an array-backend and a linked list ring backend: TODO:
type RingLinkedListBackend[IDType shitdb.IDConstraint] struct {
	idCache map[IDType]*Node[IDType]
	size    int
	maxSize int
	lock    *sync.RWMutex

	newest, oldest *Node[IDType]
}

func (r *RingLinkedListBackend[IDType]) delete(id IDType) bool {
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

// Returns true if exit early due to filter
func (r *RingLinkedListBackend[IDType]) queryFunc(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType], action func(g shitdb.Group[IDType])) bool {
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
		idp := r.idCache[idPointer.ID]
		if idp == nil {
			if idPointer.ApproximationBehavior == shitdb.APPROXIMATE_QUIT_EARLY {
				return false
			}

			// TODO: Implement approximation behavior.
			// id := idPointer.ID
			// n = r.newest
			// idpNext := nextNewToOld[IDType]
	        
			// if idPointer.Hint == shitdb.LOCATION_HINT_OLDEST {
			// 	n = r.oldest
			// 	idpNext = nextOldToNew[IDType]
			// }
	
			// for {
			// 	if n == nil {
			// 		// we couldn't find the starting id, so quit early
			// 		return
			// 	}
			// 	if n.Value.GetID() == id {
			// 		// Gotcha
			// 		break
			// 	}
			// 	n = idpNext(n)
			// }
		} else {
			n = idp
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
				return true
			}
		}
		n = next(n)
	}

	return false
}
