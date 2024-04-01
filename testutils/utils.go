package testutils

import (
	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/types"
)

type TestGroup struct {
	ID int
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
