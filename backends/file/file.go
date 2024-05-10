package file

import (
	"os"
	"slices"
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
	groupSizeSize int
}

type FileOpts struct {
	// The size of the uint in bytes. Valid values are 2 (uint16), 4 (uint32), 8 (uint64)
	GroupSizeSize int
	// The newest group has the highest id
	NewestIsLargest bool
	// Path to the file
	Path string
	// Perms for the file
	Perms os.FileMode
}

type TplGroup[IDType subdb.IDConstraint] struct {
	// Example group structure
	Group subdb.Group[IDType]
	// Example fields of the group. First field must be the ID
	Fields []subdb.Field
}

// Make a new file-only backend. In practice, NewFile should be used.
func NewFileOnly[IDType subdb.IDConstraint](opts *FileOpts, tpl *TplGroup[IDType]) *RealFile[IDType] {
	if opts == nil {
		panic("Can't create file backend - no opts")
	}
	if opts.GroupSizeSize == 0 {
		opts.GroupSizeSize = 4
	}
	if opts.Path == "" {
		panic("Couldn't create file - path is empty")
	}
	if opts.Perms == 0 {
		opts.Perms = 0755
	}

	f, err := os.OpenFile(opts.Path, os.O_RDWR | os.O_CREATE, opts.Perms)
	if err != nil {
		panic(err)
	}

	return &RealFile[IDType]{
		f:               f,
		lock:            &sync.Mutex{},
		templateGroup:   tpl.Group,
		templateFields:  tpl.Fields,
		newestIsLargest: opts.NewestIsLargest,
		groupSizeSize:   opts.GroupSizeSize,
	}
}

func (r *RealFile[IDType]) ReadID(ids ...IDType) []subdb.Group[IDType] {
	m := make(map[IDType]bool, len(ids))

	for _, id := range ids {
		m[id] = true
	}

	oldToNew := false

	o := make([]subdb.Group[IDType], 0, len(m))

	r.lock.Lock()
	defer r.lock.Unlock()

	r.readFunc(oldToNew, oldToNew, 0, func(gData []byte, s, e int64) bool {
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

	if !oldToNew {
		slices.Reverse(o)
	}

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

	oldToNew := false

	r.lock.Lock()
	defer r.lock.Unlock()

	r.readFunc(oldToNew, oldToNew, 0, func(gData []byte, s, e int64) bool {
		f, _ := parseField(r.templateFields[0], gData)
		id := f.GetValue().(IDType)
		if !m[id] {
			return false
		}
		
		o = append(o, [2]int64{s, e})

		delete(m, id)
		return len(m) == 0
	})

	r.deleteRanges(oldToNew, o)
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
