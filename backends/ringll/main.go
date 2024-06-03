package ringll

import (
	"sync"

	"github.com/shadiestgoat/subdb"
)

type Node[IDType subdb.IDConstraint] struct {
	Value      subdb.Group[IDType]
	Next, Prev *Node[IDType]
}

// A linked list ring backend. Works in a similar way to a regular ring backend, but instead of an array, uses a linked list
// Comparison table between an array-backend and a linked list ring backend: TODO:
type RingLinkedListBackend[IDType subdb.IDConstraint] struct {
	idCache         map[IDType]*Node[IDType]
	size            int
	maxSize         int
	newestIsLargest bool
	lock            *sync.RWMutex

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
func (r *RingLinkedListBackend[IDType]) queryFunc(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType], action func(g subdb.Group[IDType])) bool {
	if len(r.idCache) == 0 {
		return false
	}

	var idpID IDType

	if idPointer != nil {
		idpID = idPointer.ID

		largest, smallest := r.oldest.Value.GetID(), r.newest.Value.GetID()

		if r.newestIsLargest {
			largest, smallest = smallest, largest
		}

		if idpID > largest || idpID < smallest {
			// idp is outside of this cache.
			return false
		}
	}

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
			if idPointer.ApproximationBehavior == subdb.APPROXIMATE_QUIT_EARLY {
				return false
			}

			idpNode := r.newest
			idpNext := nextNewToOld[IDType]

			if idPointer.Hint == subdb.LOCATION_HINT_OLDEST {
				idpNode = r.oldest
				idpNext = nextOldToNew[IDType]
			}

			// xor, lmao
			nLtIDP := r.newestIsLargest != oldToNew

			for {
				if idpNode == nil {
					// we couldn't find the starting id, so quit early
					// Its not a filter quit early, so return false.
					return false
				}

				idpNodeID := idpNode.Value.GetID()

				if nLtIDP != (idpNodeID < idpID) {
					idp = idpNode
					break
				}

				// Basically what were looking for is a value thats either larger
				idpNode = idpNext(idpNode)
			}

			// by this point we definitely, 100% gottem (I think)

			if oldToNew && idPointer.ApproximationBehavior == subdb.APPROXIMATE_OLDEST {
				idp = idp.Prev
			} else if !oldToNew && idPointer.ApproximationBehavior == subdb.APPROXIMATE_NEWEST {
				idp = idp.Next
			}

			if idPointer.ExcludePointer {
				if oldToNew {
					idp = idp.Next
				} else {
					idp = idp.Prev
				}
			}
		}

		// This is basically a sanity check. It should always be false, but best make sure I guess?
		if idp == nil {
			return false
		}

		n = idp
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
