package shitdb

import "sync"

// AUTO GENERATED, DO NOT EDIT

type InsertFunc[IDType IDConstraint] func(groups ...Group[IDType])
type DeleteIDFunc[IDType IDConstraint] func(ids ...IDType)
type DeleteQueryFunc[IDType IDConstraint] func(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType])
type ReadFunc[IDType IDConstraint] func(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) ([]Group[IDType], bool)
type ReadIDFunc[IDType IDConstraint] func(ids ...IDType) []Group[IDType]

type Hooks[IDType IDConstraint] struct {
	Insert      []InsertFunc[IDType]
	DeleteID    []DeleteIDFunc[IDType]
	DeleteQuery []DeleteQueryFunc[IDType]
	Read        []ReadFunc[IDType]
	ReadID      []ReadIDFunc[IDType]
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

func (h *Hooks[IDType]) DoDeleteQuery(cb chan bool, idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) {
	l := &sync.WaitGroup{}
	l.Add(len(h.DeleteQuery))

	for _, h := range h.DeleteQuery {
		go func() {
			h(idPointer, oldToNew, f)
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

func (h *Hooks[IDType]) DoReadID(ids ...IDType) []Group[IDType] {
	o := []Group[IDType]{}

	for _, f := range h.ReadID {
		buff := f(ids...)
		o = append(o, buff...)

		if len(o) == len(ids) {
			break
		}
	}

	return o
}

func (h *Hooks[IDType]) DoRead(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) ([]Group[IDType], bool) {
	o := []Group[IDType]{}
	cutFirst := 0

	for _, h := range h.Read {
		buf, exitEarly := h(idPointer, oldToNew, f)
		o = append(o, buf[cutFirst:]...)
		if exitEarly {
			return o, true
		}
		if len(buf) != 0 {
			idPointer = &IDPointer[IDType]{
				ID:                    buf[len(buf)-1].GetID(),
				ApproximationBehavior: APPROXIMATE_NEWEST,
			}
			cutFirst = 1
		}
	}

	return o, false
}
