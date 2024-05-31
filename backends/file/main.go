package file

import (
	"time"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/backends/all"
)


type File[IDType subdb.IDConstraint] struct {
	flush *all.AllBackend[IDType]
	file  *RealFile[IDType]
	FlushFreq time.Duration
	t *time.Ticker
	stopChan chan bool
}

func NewFileBackend[IDType subdb.IDConstraint](opts *FileOpts, tpl *TplGroup[IDType], flushFrequency time.Duration) *File[IDType] {
	return &File[IDType]{
		flush: all.NewAllBackend[IDType](opts.NewestIsLargest),
		file:  NewFileOnly(opts, tpl),
		FlushFreq: flushFrequency,
	}
}

// Registers all the hooks except Load!
func (r *File[IDType]) Register(h *subdb.Hooks[IDType]) {
	h.Insert = append(h.Insert, r.Insert)
	h.DeleteID = append(h.DeleteID, r.DeleteID)
	h.Delete = append(h.Delete, r.Delete)
	h.Read = append(h.Read, r.Read)
	h.ReadID = append(h.ReadID, r.ReadID)
	h.Stop = append(h.Stop, r.Stop)
	h.Start = append(h.Start, r.Start)
}

func (r *File[IDType]) ReadID(ids ...IDType) []subdb.Group[IDType] {
	return subdb.HooksReadID([]subdb.ReadIDFunc[IDType]{
		r.flush.ReadID, r.file.ReadID,
	}, ids...)
}

func (r *File[IDType]) Read(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) ([]subdb.Group[IDType], bool) {
	return subdb.HooksRead([]subdb.ReadFunc[IDType]{
		r.flush.Read, r.file.Read,
	}, idPointer, oldToNew, f)
}

func (r *File[IDType]) DeleteID(ids ...IDType) {
	deletedIDs := r.flush.UtilDeleteIDs(ids...)

	if len(deletedIDs) == len(ids) {
		return
	}

	deletedMap := map[IDType]bool{}

	for _, id := range deletedIDs {
		deletedMap[id] = true
	}

	newIDs := make([]IDType, 0, len(ids) - len(deletedIDs))

	for _, id := range ids {
		if !deletedMap[id] {
			newIDs = append(newIDs, id)
		}
	}

	r.file.DeleteID(newIDs...)
}

func (r *File[IDType]) Delete(idPointer *subdb.IDPointer[IDType], oldToNew bool, f subdb.Filter[IDType]) {
	g, exitEarly := r.flush.UtilDelete(idPointer, oldToNew, f)
	if exitEarly {
		return
	}
	if g != nil {
		idPointer = &subdb.IDPointer[IDType]{
			ID:                    g.GetID(),
			ApproximationBehavior: subdb.APPROXIMATE_QUIT_EARLY,
		}
	}
	r.file.Delete(idPointer, oldToNew, f)
}

func (r *File[IDType]) Insert(groups ...subdb.Group[IDType]) {
	r.flush.Insert(groups...)
}

func (r *File[IDType]) Flush() {
	r.file.Insert(r.flush.Reset()...)
}

func (r *File[IDType]) loopFlush() {
	for {
		select {
		case <- r.stopChan:
			return
		case <- r.t.C:
			r.Flush()
		}
	}
}

func (r *File[IDType]) Start() {
	r.t = time.NewTicker(r.FlushFreq)
	r.stopChan = make(chan bool)
	go r.loopFlush()
}

func (r *File[IDType]) Stop() {
	close(r.stopChan)
	r.t.Stop()
	r.t = nil
	r.Flush()
}
