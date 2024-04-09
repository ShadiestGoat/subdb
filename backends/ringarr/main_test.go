package ringarr_test

import (
	"testing"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/backends/ringarr"
	"github.com/shadiestgoat/subdb/testutils"
)

const (
	RING_SIZE = 100
)

var (
	TEST_IDS = []int{
		-1, 0, 1, RING_SIZE - 1, RING_SIZE - 1, RING_SIZE,
	}
)

func testNew(newestIsBiggest bool) subdb.BackendWithEverything[int] {
	return ringarr.NewRingArrayBackend[int](RING_SIZE, newestIsBiggest)
}

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
		}
		if len(dumped) != neededLen {
			t.Logf("Expected data len to be %v, got %v", neededLen, len(dumped))
			t.Fail()
		}
		if expectedLastID != realLastID {
			t.Logf("The latest data isn't what is expected - expected %v, got %v", expectedLastID, realLastID)
			t.Fail()
		}
	}
}

func TestRead(t *testing.T) {
	d := ringarr.NewRingArrayBackend[int](RING_SIZE, true)
	d.Insert(testutils.MakeData(RING_SIZE)...)

	t.Run("IDs", func(t *testing.T) {
		o := d.ReadID(TEST_IDS...)

		if len(o) != len(TEST_IDS) - 2 {
			t.Logf("Read back incorrect amt of IDs: inserted: %#v; read: %#v", TEST_IDS, o)
			t.FailNow()
		}

		expectedData := TEST_IDS[1:len(TEST_IDS)-1]

		for i, v := range o {
			if v.GetID() != expectedData[i] {
				t.Logf("Failed to read back IDs: i: %v, expected: %v, got: %v", i, expectedData[i], v.GetID())
				t.Fail()
			}
		}
	})

	testutils.GenerateGenericQueryTest(RING_SIZE, t, testNew, testutils.GenericReadQueryTest)
}


func TestDelete(t *testing.T) {
	d := ringarr.NewRingArrayBackend[int](RING_SIZE, true)
	d.Insert(testutils.MakeData(RING_SIZE)...)

	t.Run("IDs", func(t *testing.T) {
		d.DeleteID(TEST_IDS...)
		o := d.ReadID(TEST_IDS...)

		if len(o) != 0 {
			t.Logf("Didn't delete all IDs: tried deleting: %#v; read: %#v", TEST_IDS, o)
			t.FailNow()
		}
	})

	testutils.GenerateGenericQueryTest(RING_SIZE, t, testNew, testutils.GenericDeleteQueryTest)
}
