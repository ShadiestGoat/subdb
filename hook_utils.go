package subdb

func HooksReadID[IDType IDConstraint](hooks []ReadIDFunc[IDType], ids ...IDType) []Group[IDType] {
	o := []Group[IDType]{}

	idMap := map[IDType]bool{}

	for _, id := range ids {
		idMap[id] = true
	}

	for i, f := range hooks {
		buff := f(ids...)
		o = append(o, buff...)

		if len(o) == len(ids) || i == len(hooks) - 1 {
			break
		}

		ids = make([]IDType, 0, len(idMap))

		for id := range idMap {
			ids = append(ids, id)
		}
	}

	return o
}

func HooksRead[IDType IDConstraint](hooks []ReadFunc[IDType], idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) ([]Group[IDType], bool) {
	o := []Group[IDType]{}
	cutFirst := 0

	for _, h := range hooks {
		buf, exitEarly := h(idPointer, oldToNew, f)
		o = append(o, buf[cutFirst:]...)
		if exitEarly {
			return o, true
		}
		if len(buf) != 0 {
			idPointer = &IDPointer[IDType]{
				ID:                    buf[len(buf)-1].GetID(),
				ApproximationBehavior: APPROXIMATE_QUIT_EARLY,
			}
			cutFirst = 1
		}
	}

	return o, false
}
