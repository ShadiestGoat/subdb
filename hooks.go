package subdb

import "sync"

// AUTO GENERATED, DO NOT EDIT

type InsertFunc[IDType IDConstraint] func(groups ...Group[IDType])
type DeleteIDFunc[IDType IDConstraint] func(ids ...IDType)
type DeleteFunc[IDType IDConstraint] func(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType])
type ReadFunc[IDType IDConstraint] func(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) ([]Group[IDType], bool)
type ReadIDFunc[IDType IDConstraint] func(ids ...IDType) []Group[IDType]

type BackendWithInsertFunc[IDType IDConstraint] interface {
	Insert(groups ...Group[IDType])
}
type BackendWithDeleteIDFunc[IDType IDConstraint] interface {
	DeleteID(ids ...IDType)
}
type BackendWithDeleteFunc[IDType IDConstraint] interface {
	Delete(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType])
}
type BackendWithReadFunc[IDType IDConstraint] interface {
	Read(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) ([]Group[IDType], bool)
}
type BackendWithReadIDFunc[IDType IDConstraint] interface {
	ReadID(ids ...IDType) []Group[IDType]
}

type BackendWithEverything[IDType IDConstraint] interface {
	BackendWithInsertFunc[IDType]
	BackendWithDeleteIDFunc[IDType]
	BackendWithDeleteFunc[IDType]
	BackendWithReadFunc[IDType]
	BackendWithReadIDFunc[IDType]
}

type Hooks[IDType IDConstraint] struct {
	Insert   []InsertFunc[IDType]
	DeleteID []DeleteIDFunc[IDType]
	Delete   []DeleteFunc[IDType]
	Read     []ReadFunc[IDType]
	ReadID   []ReadIDFunc[IDType]
}

func (h *Hooks[IDType]) DoInsert(cb chan bool, groups ...Group[IDType]) {
	l := &sync.WaitGroup{}
	l.Add(len(h.Insert))

	for _, h := range h.Insert {
		go func() {
			h(groups...)
			l.Done()
		}()
	}

	if cb != nil {
		go func() {
			l.Wait()
			cb <- true
		}()
	}
}

func (h *Hooks[IDType]) DoDeleteID(cb chan bool, ids ...IDType) {
	l := &sync.WaitGroup{}
	l.Add(len(h.DeleteID))

	for _, h := range h.DeleteID {
		go func() {
			h(ids...)
			l.Done()
		}()
	}

	if cb != nil {
		go func() {
			l.Wait()
			cb <- true
		}()
	}
}

func (h *Hooks[IDType]) DoDelete(cb chan bool, idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) {
	l := &sync.WaitGroup{}
	l.Add(len(h.Delete))

	filters := make([]Filter[IDType], len(h.Delete))
	for i := range filters {
		filters[i] = f.Copy()
	}

	for i, h := range h.Delete {
		go func() {
			h(idPointer, oldToNew, filters[i])
			l.Done()
		}()
	}

	if cb != nil {
		go func() {
			l.Wait()
			cb <- true
		}()
	}
}

func (h *Hooks[IDType]) DoRead(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) ([]Group[IDType], bool) {
	return HooksRead(h.Read, idPointer, oldToNew, f)
}

func (h *Hooks[IDType]) DoReadID(ids ...IDType) []Group[IDType] {
	return HooksReadID(h.ReadID, ids...)
}
