package testutils

import (
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
	T *testing.T
}

func (f *Filter[IDType]) Match(g subdb.Group[IDType]) (bool, bool) {
	f.T.Logf("Filter Match: %v (limit: %v)", g.GetID(), f.Limit)
	f.Limit--
	return true, f.Limit == 0
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
