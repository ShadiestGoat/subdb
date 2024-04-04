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
		data := []int{
			0, 1, 99, RING_SIZE - 1, RING_SIZE,
		}
		o := d.ReadIDs(data...)

		if len(o) == len(data) {
			t.Logf("Read back all IDs (needed to not read last ID): inserted: %#v; read: %#v", data, o)
			t.FailNow()
		}

		for i, v := range o {
			if v.GetID() != data[i] {
				t.Logf("Failed to read back IDs: i: %v, expected: %v, got: %v", i, data[i], v.GetID())
				t.Fail()
			}
		}
	})

	arr := []bool{true, false}

	for _, hasIDP := range arr {
		for _, oldToNew := range arr {
			name := "Query"
			var idp *subdb.IDPointer[int]

			if hasIDP {
				name += "IDP"
				idp = &subdb.IDPointer[int]{
					ID:                    RING_SIZE/2,
				}
			} else {
				name += "NoIDP"
			}

			if oldToNew {
				name += "oldToNew"
			} else {
				name += "newToOld"
			}

			t.Run(name, func(t *testing.T) {
				querySize := 10
				f := &testutils.Filter[int]{
					Limit: querySize,
					T: t,
				}

				if idp != nil {
					t.Logf("IDP: %v", idp.ID)
				}

				o, early := d.ReadQuery(idp, oldToNew, f)
		
				if !early || len(o) != querySize {
					t.Logf("Failed query, expected early exit with 10 values, got: early: %v, vals: %#v", early, o)
					t.FailNow()
				}

				var eFirst int
				rFirst, rLast := o[0].GetID(), o[len(o)-1].GetID()


				if idp == nil {
					if oldToNew {
						eFirst = 0
					} else {
						eFirst = RING_SIZE - 1
					}
				} else {
					eFirst = idp.ID
				}

				if eFirst != rFirst {
					t.Logf("Unexpected first value! Expected: %v, got: %v", eFirst, rFirst)
					t.Fail()
				}

				d := querySize - 1
				if !oldToNew {
					d *= -1
				}
				eLast := eFirst + d

				if eLast != rLast {
					t.Logf("Unexpected last value! Expected: %v, got: %v", eLast, rLast)
					t.Fail()
				}
			})
		}
	}
}
