package ringll_test

import (
	"testing"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/backends/ringll"
	ringTest "github.com/shadiestgoat/subdb/testutils/ring_test"
)

func newBackend(dataSize int, newestIsBiggest bool) subdb.BackendWithEverything[int] {
	return ringll.NewRing[int](dataSize, newestIsBiggest)
}


func TestInsert(t *testing.T) {
	ringTest.Insert(t, newBackend)
}

func TestRead(t *testing.T) {
	ringTest.Insert(t, newBackend)
}

func TestDelete(t *testing.T) {
	ringTest.Insert(t, newBackend)
}
