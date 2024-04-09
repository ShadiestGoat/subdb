package testutils

import (
	"slices"
	"strings"
	"testing"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/types"
)

type TestGroup struct {
	ID int
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

type QuerySetupOpts struct {
	OldToNew,
	NewestIsBiggest bool
	DataSize int
}

func GenerateGenericQueryTest(dataSize int, t *testing.T, newBackend func(newestIsBiggest bool) interface { subdb.BackendWithInsertFunc[int]; subdb.BackendWithReadFunc[int] }) {
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
				var idp *subdb.IDPointer[int]
	
				if hasIDP {
					idp = &subdb.IDPointer[int]{
				ID:                    dataSize/2,
						ExcludePointer: exclIDP,
					}
				}
	
				if !hasIDP && exclIDP {
					continue
				}
	
				for _, oldToNew := range arr {
					GenericQueryTest(idp, &QuerySetupOpts{
						OldToNew:        oldToNew,
						NewestIsBiggest: newestIsLargest,
						DataSize:        dataSize,
					}, t, b.Read)
				}
			}
		}
	}
}

func GenericQueryTest(idp *subdb.IDPointer[int], opts *QuerySetupOpts, t *testing.T, rf subdb.ReadFunc[int]) {
	name := []string{"read"}
	if idp != nil {
		name = append(name, "idp")
		if idp.ExcludePointer {
			name = append(name, "excl")
		}
	}

	if opts.OldToNew {
		name = append(name, "oldToNew")
	} else {
		name = append(name, "newToOld")
	}

	if opts.NewestIsBiggest {
		name = append(name, "newBig")
	} else {
		name = append(name, "newSmall")
	}

	t.Run(strings.Join(name, "_"), func(t *testing.T) {
		querySize := 10
		f := &Filter[int]{
			Limit: querySize,
			T:     t,
		}

		if idp != nil {
			t.Logf("IDP: %v", idp.ID)
		}

		o, early := rf(idp, opts.OldToNew, f)

		if !early || len(o) != querySize {
			t.Logf("Failed query, expected early exit with 10 values, got: early: %v, vals: %#v", early, o)
			t.FailNow()
		}

		var eFirst int
		rFirst, rLast := o[0].GetID(), o[len(o)-1].GetID()

		dir := 1

		if opts.OldToNew != opts.NewestIsBiggest {
			dir = -1
		}

		if idp == nil {
			if opts.OldToNew {
				eFirst = 0
			} else {
				eFirst = opts.DataSize - 1
			}
			if !opts.NewestIsBiggest {
				eFirst = opts.DataSize - 1 - eFirst
			}
		} else {
			eFirst = idp.ID
			if idp.ExcludePointer {
				eFirst += dir
			}
		}

		if eFirst != rFirst {
			t.Logf("Unexpected first value! Expected: %v, got: %v", eFirst, rFirst)
			t.Fail()
		}

		d := querySize - 1
		eLast := eFirst + (d * dir)

		if eLast != rLast {
			t.Logf("Unexpected last value! Expected: %v, got: %v", eLast, rLast)
			t.Fail()
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
