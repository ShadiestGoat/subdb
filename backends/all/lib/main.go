package lib

import (
	"slices"

	subdb "github.com/shadiestgoat/subdb"
)

func cmpNewIsLarge[IDType subdb.IDConstraint](g subdb.Group[IDType], t IDType) int {
	v := g.GetID()

	if t < v {
		return 1
	}
	if t == v {
		return 0
	}

	return -1
}


func cmpOldIsLarge[IDType subdb.IDConstraint](g subdb.Group[IDType], t IDType) int {
	return -cmpNewIsLarge(g, t)
}

func (r *CommonArrayBackendUtil[IDType]) queryFunc(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType], action func(g subdb.Group[IDType], i int)) bool {
	var i int

	d := -1

	if oldToNew {
		d = 1
	}

	if idPointer != nil {
		idp := idPointer.ID

		if idPointer.ApproximationBehavior == subdb.APPROXIMATE_QUIT_EARLY && r.IDCache[idp] == nil {
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

		cmpFunc := cmpOldIsLarge[IDType]

		if r.NewestIsLargest {
			cmpFunc = cmpNewIsLarge
		}

		closest, ok := slices.BinarySearchFunc(r.Items, idp, cmpFunc)

		if !ok && idPointer.ApproximationBehavior == subdb.APPROXIMATE_OLDEST {
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
