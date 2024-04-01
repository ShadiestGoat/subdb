package ringarr_test

import (
	"testing"

	"github.com/shadiestgoat/subdb/backends/ringarr"
	"github.com/shadiestgoat/subdb/testutils"
)

const (
	RING_SIZE = 100
)

func TestInsert(t *testing.T) {
	sizes := []int{
		RING_SIZE/10,
		RING_SIZE - 1,
		RING_SIZE,
		RING_SIZE + 1,
		RING_SIZE * 2,
	}

	// Basically, we are testing to make that insert works:
	// Only a max of RING_SIZE is kept
	// Latest data is kept only
	for _, s := range sizes {
		b := ringarr.NewRingArrayBackend[int](RING_SIZE, true)
		d := testutils.MakeData(s)
		b.Insert(d...)
		
		dumped, cache := b.Dump()
		neededLen := s
		if neededLen > RING_SIZE {
			neededLen = RING_SIZE
		}
		expectedLastID := d[len(d)-1].GetID()
		realLastID := dumped[len(dumped)-1].GetID()

		if len(dumped) != len(cache) {
			t.Logf("Dumped data len != len of dumped cache (%v vs %v)", len(dumped), len(cache))
			t.Fail()
			return
		}
		if len(dumped) != neededLen {
			t.Logf("Expected data len to be %v, got %v", neededLen, len(dumped))
			t.Fail()
			return
		}
		if expectedLastID != realLastID {
			t.Logf("The latest data isn't what is expected - expected %v, got %v", expectedLastID, realLastID)
			t.Fail()
			return
		}
	}
}
