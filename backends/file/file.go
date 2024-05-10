package file

import (
	"os"
	"sync"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/types"
)

//
// | group size | [id]group data | group size |
//
// group size doesn't include itself x2 - only the size of the data\
//
// 1234 1234
//    ^ ^
//    e s

// Shouldn't be used as a backend, but if you REALLY want to then go ahead ig
type RealFile[IDType subdb.IDConstraint] struct {
	f *os.File
	lock *sync.Mutex
	templateGroup subdb.Group[IDType]
	templateFields []subdb.Field
	newestIsLargest bool
	// The size of the uint in bytes. Valid values are 2 (uint16), 4 (uint32), 8 (uint64)
	groupSizeSize int
}


func (r *RealFile[IDType]) ReadID(ids ...IDType) []subdb.Group[IDType] {
	m := make(map[IDType]bool, len(ids))

	for _, id := range ids {
		m[id] = true
	}

	o := make([]subdb.Group[IDType], 0, len(m))

	r.lock.Lock()
	defer r.lock.Unlock()

	r.readFunc(true, 0, func(gData []byte, s, e int64) bool {
		f, off := parseField(r.templateFields[0], gData)
		id := f.GetValue().(IDType)
		if !m[id] {
			return false
		}

		g := parseGroupWithoutID(r.templateGroup, r.templateFields[1:], f, gData[off:])
		o = append(o, g)
		
		delete(m, id)
		return len(m) == 0
	})

	return o
}

func (r *RealFile[IDType]) Read(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) ([]subdb.Group[IDType], bool) {
	r.lock.Lock()
	defer r.lock.Unlock()

	o := []subdb.Group[IDType]{}

	exitEarly := r.queryFunc(idPointer, oldToNew, f, func(g subdb.Group[IDType], s, e int64) {
		o = append(o, g)
	})

	return o, exitEarly
}

func (r *RealFile[IDType]) DeleteID(ids ...IDType) {
	m := make(map[IDType]bool, len(ids))

	for _, id := range ids {
		m[id] = true
	}

	o := make([][2]int64, 0, len(m))

	r.lock.Lock()
	defer r.lock.Unlock()

	r.readFunc(true, 0, func(gData []byte, s, e int64) bool {
		f, _ := parseField(r.templateFields[0], gData)
		id := f.GetValue().(IDType)
		if !m[id] {
			return false
		}
		
		o = append(o, [2]int64{s, e})

		delete(m, id)
		return len(m) == 0
	})

	r.deleteRanges(true, o)
}

func (r *RealFile[IDType]) Delete(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) {
	r.lock.Lock()
	defer r.lock.Unlock()

	o := [][2]int64{}

	r.queryFunc(idPointer, oldToNew, f, func(_ subdb.Group[IDType], s, e int64) {
		o = append(o, [2]int64{s, e})
	})

	r.deleteRanges(oldToNew, o)
}

func (r *RealFile[IDType]) Insert(inp ...subdb.Group[IDType]) {
	buff := []byte{}

	for _, g := range inp {
		fields := g.Store()
		gBuff := []byte{}

		for _, f := range fields {
			gBuff = append(gBuff, f.Encode()...)
		}

		sCol := types.NewUint(uint(len(gBuff)), r.groupSizeSize)
		sizeBuff := sCol.Encode()

		gBuff = append(sizeBuff, gBuff...)

		buff = append(buff, append(gBuff, sizeBuff...)...)
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.f.Seek(0, 2)
	_, err := r.f.Write(buff)

	if err != nil {
		panic(err)
	}
}
