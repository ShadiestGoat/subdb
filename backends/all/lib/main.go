package lib

import (
	"sort"

	"shadygoat.eu/shitdb"
)

func (r *CommonArrayBackendUtil[IDType]) genQueryCMP(t IDType) func(int) int {
	if r.NewestIsLargest {
		return func(i int) int {
			v := r.Items[i].GetID()

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
		v := r.Items[i].GetID()

		if t > v {
			return -1
		}
		if t == v {
			return 0
		}

		return 1
	}
}

func (r *CommonArrayBackendUtil[IDType]) queryFunc(idPointer *shitdb.IDPointer[IDType], oldToNew bool, f shitdb.Filter[IDType], action func(g shitdb.Group[IDType], i int)) bool {
	var i int

	d := -1

	if oldToNew {
		d = 1
	}

	if idPointer != nil {
		idp := idPointer.ID

		if idPointer.ApproximationBehavior == shitdb.APPROXIMATE_QUIT_EARLY && r.IDCache[idp] == nil {
			return false
		}

		var large, small IDType

		large = r.Items[0].GetID()
		small = r.Items[len(r.Items)-1].GetID()

		if r.NewestIsLargest {
			large, small = small, large
		}
		if idp < small || idp > large {
			return false
		}

		closest, ok := sort.Find(len(r.Items), r.genQueryCMP(idp))

		if !ok && idPointer.ApproximationBehavior == shitdb.APPROXIMATE_OLDEST {
			closest--
		}

		i = closest
	} else if !oldToNew {
		i = len(r.Items) - 1
	}

	for {
		v := r.Items[i]

		ok, early := f.Match(v)

		if ok {
			action(v, i)
		}

		i += d

		if early || i == -1 || i == len(r.Items) {
			return early
		}
	}
}
