package file

import (
	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/backends/all"
)


type File[IDType subdb.IDConstraint] struct {
	flush *all.AllBackend[IDType]
	file  *all.AllBackend[IDType]
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
