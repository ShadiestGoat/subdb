package ringarr_test

import (
	"testing"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/backends/ringarr"
	"github.com/shadiestgoat/subdb/testutils"
	ringTest "github.com/shadiestgoat/subdb/testutils/ring_test"
)

func newBackend(dataSize int, newestIsBiggest bool) subdb.BackendWithEverything[int] {
	return ringarr.NewRingArrayBackend[int](dataSize, newestIsBiggest)
}

const DUMP_RING_SIZE = 10

func TestDump(t *testing.T) {
	b := ringarr.NewRingArrayBackend[int](DUMP_RING_SIZE, true)
	b.Insert(testutils.MakeData(DUMP_RING_SIZE)...)
	groups, idCache := b.Dump()
	if len(groups) != len(idCache) {
		t.Errorf("Expected to have a same length groups & idCache, got: %v & %v", len(groups), len(idCache))
	}

	b.Insert(testutils.MakeData(DUMP_RING_SIZE * 3)[DUMP_RING_SIZE:]...)
	groups, idCache = b.Dump()

	if len(groups) != len(idCache) || len(groups) != DUMP_RING_SIZE {
		t.Errorf("Expected to have a valid post 2nd insert dump, got: %v & %v (expected %v)", len(groups), len(idCache), DUMP_RING_SIZE)
	}
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
