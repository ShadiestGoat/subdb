package ringTest

import (
	"testing"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/testutils"
)

const (
	RING_SIZE = 100
)


var (
	TestSizes = []int{
		RING_SIZE/10,
		RING_SIZE - 1,
		RING_SIZE,
		RING_SIZE + 1,
		RING_SIZE * 2,
	}
	TEST_IDS = []int{
		-1, 0, 1, RING_SIZE - 1, RING_SIZE - 1, RING_SIZE,
	}
)

type NewBackend func(dataSize int, newestIsBiggest bool) subdb.BackendWithEverything[int]

func backendWrap(f NewBackend) func (bool) subdb.BackendWithEverything[int] {
	return func(b bool) subdb.BackendWithEverything[int] {
		return f(RING_SIZE, b)
	}
}

func Insert(t *testing.T, newBackend NewBackend) {
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
		b := newBackend(RING_SIZE, true)
		d := testutils.MakeData(s)
		b.Insert(d...)
		
		dumped, _ := b.Read(nil, true, testutils.MatchAll[int]{})
		neededLen := s
		if neededLen > RING_SIZE {
			neededLen = RING_SIZE
		}
		eFirstI := len(d) - 1 - (RING_SIZE - 1)
		if eFirstI < 0 {
			eFirstI = 0
		}

		eFirst, eLast := d[eFirstI].GetID(), d[len(d)-1].GetID()
		rFirst, rLast := dumped[0].GetID(), dumped[len(dumped)-1].GetID()

		if len(dumped) != neededLen {
			t.Errorf("Expected data len to be %v, got %v", neededLen, len(dumped))
		}
		if eLast != rLast {
			t.Errorf("The latest data isn't what is expected - expected %v, got %v", eLast, rLast)
		}
		if eFirst != rFirst {
			t.Errorf("The first val isn't what is expected - expected %v, got %v", eFirst, rFirst)
		}
	}
}

func Read(t *testing.T, testNew NewBackend) {
	d := testNew(RING_SIZE, true)
	d.Insert(testutils.MakeData(RING_SIZE)...)

	t.Run("IDs", func(t *testing.T) {
		o := d.ReadID(TEST_IDS...)

		if len(o) != len(TEST_IDS) - 2 {
			t.Fatalf("Read back incorrect amt of IDs: inserted: %#v; read: %#v", TEST_IDS, o)
		}

		expectedData := TEST_IDS[1:len(TEST_IDS)-1]

		for i, v := range o {
			if v.GetID() != expectedData[i] {
				t.Errorf("Failed to read back IDs: i: %v, expected: %v, got: %v", i, expectedData[i], v.GetID())
			}
		}
	})

	testutils.GenerateGenericQueryTest(RING_SIZE, t, backendWrap(testNew), testutils.GenericReadQueryTest)
}


func Delete(t *testing.T, newBackend NewBackend) {
	d := newBackend(RING_SIZE, true)
	d.Insert(testutils.MakeData(RING_SIZE)...)

	t.Run("IDs", func(t *testing.T) {
		d.DeleteID(TEST_IDS...)
		o := d.ReadID(TEST_IDS...)

		if len(o) != 0 {
			t.Fatalf("Didn't delete all IDs: tried deleting: %#v; read: %#v", TEST_IDS, o)
		}
	})

	testutils.GenerateGenericQueryTest(RING_SIZE, t, backendWrap(newBackend), testutils.GenericDeleteQueryTest)
}
