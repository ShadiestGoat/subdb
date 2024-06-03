package file

import (
	"slices"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/types"
)

func (r *RealFile[IDType]) parseGrpSize(grpSizeRaw []byte) int64 {
	sizeCol := types.NewUint(0, r.groupSizeSize)
	sizeCol.Load(grpSizeRaw)
	return int64(sizeCol.Value)
}

// s - start of the data, including g size
// e - end of the data, including g size
func (r *RealFile[IDType]) readFunc(oldToNew, offsetFromStart bool, offset int64, h func(gData []byte, s, e int64) bool) bool {
	stat, err := r.f.Stat()
	if err != nil {
		panic("db file stat failed: " + err.Error())
	}
	size := stat.Size()

	if !offsetFromStart {
		offset = size - offset
	}

	if !oldToNew {
		// were pointing to the start of a row that we need to read, but then we need to read in the other direction so...
		exitEarly := false
		r.readFunc(true, true, offset, func(gData []byte, s, e int64) bool {
			exitEarly = h(gData, s, e)
			return true
		})
		if exitEarly {
			return true
		}

		offset -= int64(r.groupSizeSize)
	}

	for {
		if oldToNew {
			if offset >= size {
				break
			}
		} else {
			if offset < 0 {
				break
			}
		}

		r.f.Seek(offset, 0)

		// get the grp size
		grpSizeRaw := make([]byte, r.groupSizeSize)
		r.f.Read(grpSizeRaw)
		grpSize := r.parseGrpSize(grpSizeRaw)

		if !oldToNew {
			r.f.Seek(-(grpSize + int64(r.groupSizeSize)), 1)
		}

		grpData := make([]byte, grpSize)
		r.f.Read(grpData)

		s := offset
		if !oldToNew {
			s -= grpSize + int64(r.groupSizeSize)
		}
		e := s + grpSize + int64(r.groupSizeSize)*2 - 1

		exitEarly := h(grpData, s, e)

		if exitEarly {
			return true
		}

		if oldToNew {
			offset = e + 1
		} else {
			offset = s - int64(r.groupSizeSize)
		}
	}

	return false
}

func parseField(t subdb.Field, data []byte) (subdb.Field, int) {
	f := t.New()
	var s int
	var offset int

	if fStatic, ok := f.(subdb.FieldStaticSize); ok {
		s = fStatic.StaticSize()
	} else {
		f := f.(subdb.FieldDynamicSize)
		offset = f.DynamicSizeSize()
		s = f.DynamicSize(data[:offset])
	}

	f.Load(data[offset : offset+s])

	return f, offset + s
}

func parseGroupWithoutID[IDType subdb.IDConstraint](tplGroup subdb.Group[IDType], tplFields []subdb.Field, idField subdb.Field, data []byte) subdb.Group[IDType] {
	g := tplGroup.New()

	fields := make([]subdb.Field, len(tplFields)+1)
	fields[0] = idField

	for i, tplField := range tplFields {
		f, o := parseField(tplField, data)
		data = data[o:]
		fields[i+1] = f
	}

	g.Load(fields)

	return g
}

func parseGroup[IDType subdb.IDConstraint](tplGroup subdb.Group[IDType], tplFields []subdb.Field, data []byte) subdb.Group[IDType] {
	id, off := parseField(tplFields[0], data)

	return parseGroupWithoutID(tplGroup, tplFields[1:], id, data[off:])
}

// Returns the offset of the start, if we should exitEarly
func (r *RealFile[IDType]) findIDP(idp *subdb.IDPointer[IDType], qOldToNew bool) (int64, bool) {
	oldToNew := false

	if idp.Hint == subdb.LOCATION_HINT_OLDEST {
		oldToNew = true
	}

	offset := int64(-1)

	idLtIDP := r.newestIsLargest != oldToNew

	exactFound := false
	needToReturnAsap := false

	lastStart := [2]int64{-1, -1}

	idpExcludeFunc := func(e int64, lastStartIndex int) bool {
		if qOldToNew {
			offset = e + 1
			return true
		}
		if oldToNew {
			offset = lastStart[lastStartIndex]
			return true
		}

		needToReturnAsap = true
		return false
	}

	exitEarly := r.readFunc(oldToNew, oldToNew, 0, func(gData []byte, s, e int64) bool {
		if needToReturnAsap {
			offset = s
			needToReturnAsap = false

			return true
		}

		f, _ := parseField(r.templateFields[0], gData)
		id := f.GetValue().(IDType)

		if id == idp.ID {
			exactFound = true

			if idp.ExcludePointer {
				return idpExcludeFunc(e, 0)
			}

			offset = s

			return true
		}

		if idLtIDP == (id < idp.ID) {
			if idp.ApproximationBehavior == subdb.APPROXIMATE_QUIT_EARLY {
				return true
			}

			approxOldest := idp.ApproximationBehavior == subdb.APPROXIMATE_OLDEST

			if oldToNew == approxOldest {
				offset = lastStart[0]

				return true
			} else {
				if idp.ExcludePointer {
					return idpExcludeFunc(e, 1)
				}
				offset = s

				return true
			}
		}

		lastStart[0], lastStart[1] = s, lastStart[0]

		return false
	})

	if exitEarly && !exactFound && idp.ApproximationBehavior == subdb.APPROXIMATE_QUIT_EARLY {
		return 0, true
	}

	return offset, offset == -1 || needToReturnAsap
}

func (r *RealFile[IDType]) queryFunc(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType], action func(g subdb.Group[IDType], s, e int64)) bool {
	idpMet := idPointer == nil
	offset := int64(0)
	offsetFromStart := oldToNew

	if !idpMet {
		idpOffset, exitEarly := r.findIDP(idPointer, oldToNew)
		if exitEarly {
			return false
		}
		offset = idpOffset
		offsetFromStart = true
	}

	exitEarly := r.readFunc(oldToNew, offsetFromStart, offset, func(gData []byte, s, e int64) bool {
		g := parseGroup(r.templateGroup, r.templateFields, gData)

		ok, exitEarly := f.Match(g)

		if ok {
			action(g, s, e)
		}

		return exitEarly
	})

	return exitEarly
}

func (r *RealFile[IDType]) deleteRanges(oldToNew bool, inp [][2]int64) {
	if !oldToNew {
		slices.Reverse(inp)
	}

	ranges := make([][2]int64, 0, len(inp))

	for i, v := range inp {
		lastRangeI := len(ranges) - 1
		if i != 0 && v[0]-ranges[lastRangeI][1] == 1 {
			ranges[lastRangeI][1] = v[1]
		} else {
			ranges = append(ranges, v)
		}
	}

	stat, _ := r.f.Stat()

	size := stat.Size()

	delSize := int64(0)

	for i, l := range ranges {
		readRange := [2]int64{l[1] + 1}

		if i == len(ranges)-1 {
			readRange[1] = size
		} else {
			readRange[1] = ranges[i+1][0]
		}

		buff := make([]byte, readRange[1]-readRange[0])

		r.f.ReadAt(buff, readRange[0])
		r.f.WriteAt(buff, l[0]-delSize)

		delSize += l[1] - l[0] + 1
	}

	r.f.Truncate(size - delSize)
}
