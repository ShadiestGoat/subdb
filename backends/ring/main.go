package ring

import (
	"sort"

	"shadygoat.eu/shitdb"
)

func (r *RingArrayBackend[IDType]) genQueryCMP(t IDType) func(int) int {
	if r.newestIsLargest {
		return func(i int) int {
			v := r.items[i].GetID()

			if t < v {
				return -1
			}
			if t == v {
				return 0
			}
	
			return 1
		}
	}

	return func(i int) int {
		v := r.items[i].GetID()

		if t > v {
			return -1
		}
		if t == v {
			return 0
		}

		return 1
	}
}

func (r *RingArrayBackend[IDType]) queryFunc(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType], action func(g shitdb.Group[IDType], i int)) {
	var i int

	d := -1

	if oldToNew {
		d = 1
	}

	if idPointer != nil {
		idp := idPointer.ID

		if idPointer.ApproximationBehavior == shitdb.APPROXIMATE_QUIT_EARLY && r.idCache[idp] == nil {
			return
		}

		var large, small IDType

		large = r.items[0].GetID()
		small = r.items[len(r.items)-1].GetID()

		if r.newestIsLargest {
			large, small = small, large
		}
		if idp < small || idp > large {
			return
		}

		closest, ok := sort.Find(len(r.items), r.genQueryCMP(idp))

		if !ok && idPointer.ApproximationBehavior == shitdb.APPROXIMATE_OLDEST {
			closest--
		}

		i = closest
	} else if !oldToNew {
		i = len(r.items) - 1
	}

	for {
		v := r.items[i]

		ok, early := f.Match(v)

		if ok {
			action(v, i)
		}
		
		i += d

		if early || i == -1 || i == len(r.items) {
			return
		}
	}
}
