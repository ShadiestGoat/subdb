package testutils

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/types"
)

type TestGroup struct {
	ID int
}

type MatchAll[IDType subdb.IDConstraint] struct {}

func (v MatchAll[IDType]) Match(subdb.Group[IDType]) (bool, bool) {
	return true, false
}

func (v MatchAll[IDType]) Copy() subdb.Filter[IDType] {
	return MatchAll[IDType]{}
}

// Matches the first {Limit} values
type Filter[IDType subdb.IDConstraint] struct {
	Limit int
	T     *testing.T
}

func (f *Filter[IDType]) Match(g subdb.Group[IDType]) (bool, bool) {
	f.T.Logf("Filter Match: %v (limit: %v)", g.GetID(), f.Limit)
	f.Limit--
	return true, f.Limit == 0
}

func (f *Filter[IDType]) Copy() subdb.Filter[IDType] {
	return &Filter[IDType]{
		Limit: f.Limit,
		T:     f.T,
	}
}

type QuerySetupOpts struct {
	OldToNew,
	NewestIsLargest bool
	DataSize,
	QuerySize int
}

type NewBackendFunc = func(newestIsLargest bool) subdb.BackendWithEverything[int]
type PrepBackendFunc = func (idp *subdb.IDPointer[int], q *QuerySetupOpts) subdb.BackendWithEverything[int]

func prepBackend(mkBackend NewBackendFunc) PrepBackendFunc {
	return func(idp *subdb.IDPointer[int], q *QuerySetupOpts) subdb.BackendWithEverything[int] {
		b := mkBackend(q.NewestIsLargest)
		d := MakeData(q.DataSize)

		if !q.NewestIsLargest {
			slices.Reverse(d)
		}

		b.Insert(d...)

		return b
	}
}

func GenerateGenericQueryTest(
	dataSize int, t *testing.T, newBackend func(newestIsBiggest bool) subdb.BackendWithEverything[int],
	testFunc func (idp *subdb.IDPointer[int], opts *QuerySetupOpts, t *testing.T, b PrepBackendFunc),
	) {
	arr := []bool{true, false}

	for _, newestIsLargest := range arr {
		b := newBackend(newestIsLargest)

		d := MakeData(dataSize)

		if !newestIsLargest {
			slices.Reverse(d)
		}

		b.Insert(d...)

		for _, hasIDP := range arr {
			for _, exclIDP := range arr {
							NewestIsLargest: newestIsLargest,
						}, t, prepBackend(newBackend))
				}
			}
		}
	}
}

func GenericQueryExpectation(idp *subdb.IDPointer[int], opts *QuerySetupOpts) (eFirst, eLast, eLen int, eEarly bool) {
	dir := 1

	if opts.OldToNew != opts.NewestIsLargest {
		dir = -1
	}

	if idp == nil {
		if opts.OldToNew {
			eFirst = 0
		} else {
			eFirst = opts.DataSize - 1
		}
		if !opts.NewestIsLargest {
			eFirst = opts.DataSize - 1 - eFirst
		}
	} else {
		eFirst = idp.ID
		if idp.ExcludePointer {
			eFirst += dir
		}
	}

	d := opts.QuerySize - 1
	eLast = eFirst + (d * dir)

	if eFirst < 0 {
		eLen = 0
		eFirst = 0
	}
	if eLast < 0 {
		eLast = 0
	}
	
	if eFirst >= opts.DataSize {
		eFirst = opts.DataSize - 1
	}
	if eLast >= opts.DataSize {
		eLast = opts.DataSize - 1
	}

	if eFirst == eLast && idp != nil && idp.ExcludePointer {
		eLen = 0
	} else {
		eLen = int(math.Abs(float64(eFirst) - float64(eLast))) + 1
		eEarly = eLen == opts.QuerySize
	}

	return
}

func GenerateGenericTestName(n string, idp *subdb.IDPointer[int], opts *QuerySetupOpts) string {
	name := []string{n}
	if idp != nil {
		name = append(name, "idp", fmt.Sprint(idp.ID))
		if idp.ExcludePointer {
			name = append(name, "excl")
		}
	}

	if opts.OldToNew {
		name = append(name, "oldToNew")
	} else {
		name = append(name, "newToOld")
	}

	if opts.NewestIsLargest {
		name = append(name, "newBig")
	} else {
		name = append(name, "newSmall")
	}

	return strings.Join(name, "_")
}

func GenericReadQueryTest(idp *subdb.IDPointer[int], opts *QuerySetupOpts, t *testing.T, prepBackend PrepBackendFunc) {
	t.Run(GenerateGenericTestName("read", idp, opts), func(t *testing.T) {
		b := prepBackend(idp, opts)

		f := &Filter[int]{
			Limit: opts.QuerySize,
			T:     t,
		}

		eFirst, eLast, eLen, eEarly := GenericQueryExpectation(idp, opts)
		o, early := b.Read(idp, opts.OldToNew, f)

		if len(o) != eLen {
			t.Fatalf("Unexpected output len: %v, expected %v", len(o), eLen)
		}

		if early != eEarly {
			t.Fatalf("Failed query, unexpected early exit status: %v, expected %v", early, eEarly)
		}

		if eLen == 0 {
			return
		}

		rFirst, rLast := o[0].GetID(), o[len(o)-1].GetID()

		if eFirst != rFirst {
			t.Errorf("Unexpected first value! Expected: %v, got: %v", eFirst, rFirst)
		}
		if eLast != rLast {
			t.Errorf("Unexpected last value! Expected: %v, got: %v", eLast, rLast)
		}
	})
}

func GenericDeleteQueryTest(idp *subdb.IDPointer[int], opts *QuerySetupOpts, t *testing.T, prepBackend PrepBackendFunc) {
	eFirst, eLast, eLen, _ := GenericQueryExpectation(idp, opts)

	if eLen == 0 {
		return
	}

	t.Run(GenerateGenericTestName("delete", idp, opts), func(t *testing.T) {
		b := prepBackend(idp, opts)

		f := &Filter[int]{
			Limit: opts.QuerySize,
			T:     t,
		}

		b.Delete(idp, opts.OldToNew, f)

		o := b.ReadID(eFirst, eLast)

		if len(o) != 0 {
			t.Errorf("Unexpected output! Got %#v, expected nothing!", o)
		}
	})
}

func (t TestGroup) GetID() int {
	return t.ID
}

func (t *TestGroup) Load(f []subdb.Field) {
	t.ID = f[0].GetValue().(int)
}

func (t TestGroup) New() subdb.Group[int] {
	return &TestGroup{}
}

func (t TestGroup) Store() []subdb.Field {
	return []subdb.Field{
		types.NewInt(t.ID, 2),
	}
}

func MakeData(n int) []subdb.Group[int] {
	o := []subdb.Group[int]{}

	for i := 0; i < n; i++ {
		o = append(o, &TestGroup{
			ID: i,
		})
	}

	return o
}
