package file_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shadiestgoat/subdb"
	"github.com/shadiestgoat/subdb/backends/file"
	"github.com/shadiestgoat/subdb/testutils"
)

func testFilePath(fName string) string {
	return "./" + filepath.Join("test_results", fName)
}

func newBackend(fName string, newestIsLargest bool) *file.RealFile[int] {
	err := os.Mkdir("test_results", 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}
	n := testFilePath(fName)
	os.Remove(n)

	g := &testutils.TestGroup{
		ID: 1,
	}

	return file.NewFileOnly(&file.FileOpts{
		GroupSizeSize:   4,
		NewestIsLargest: newestIsLargest,
		Path:            n,
		Perms:           0755,
	}, &file.TplGroup[int]{
		Group:  g,
		Fields: g.Store(),
	})
}

func cleanData(fName string) {
	os.Remove(testFilePath(fName))
}

func TestFileReadEmpty(t *testing.T) {
	b := newBackend("readEmpty", true)
	defer cleanData("readEmpty")

	for _, oldToNew := range []bool{true, false} {
		n := "newToOld"
		if oldToNew {
			n = "oldToNew"
		}

		t.Run("readEmpty" + n, func(t *testing.T) {
			readData, _ := b.Read(nil, true, testutils.MatchAll[int]{})

			if len(readData) != 0 {
				t.Errorf("Unexpected read output of empty - len() == %v", len(readData))
			}
		})
	}
}

func TestFileInsert(t *testing.T) {
	b := newBackend("insert", true)
	defer cleanData("insert")

	d := testutils.MakeData(10)
	b.Insert(d...)
	readData, _ := b.Read(nil, true, testutils.MatchAll[int]{})

	if len(readData) != len(d) {
		t.Fatalf("Unexpected len() of data - %v/%v", len(readData), len(d))
	}

	for i, g := range readData {
		e, r := d[i].GetID(), g.GetID()
		if r != e {
			t.Errorf("Unexpected data at %v: expected %v, got %v", i, e, r)
		}
	}
}

func TestFileReadID(t *testing.T) {
	b := newBackend("ids", true)
	defer cleanData("ids")

	d := testutils.MakeData(10)

	ids := []int{}
	for i := 0; i < 10; i++ {
		ids = append(ids, i)
	}

	b.Insert(d...)
	readData := b.ReadID(ids...)

	if len(readData) != len(d) {
		t.Errorf("Got bad read id len, exp: %v, got %v", len(d), len(readData))
	}

	for i, v := range readData {
		e, r := d[i].GetID(), v.GetID()
		if r != e {
			t.Errorf("Unexpected data at %v: expected %v, got %v", i, e, r)
		}
	}
}

func TestFileRead(t *testing.T) {
	i := 0
	testutils.GenerateGenericQueryTest(500, t, func(newestIsBiggest bool) subdb.BackendWithEverything[int] {
		b := newBackend("read" + fmt.Sprint(i), newestIsBiggest)
		i++
		return b
	}, testutils.GenericReadQueryTest)

	for j := 0; j < i; j++ {
		cleanData("read" + fmt.Sprint(j))
	}
}

func TestFileDeleteID(t *testing.T) {
	b := newBackend("delIDs", true)
	defer cleanData("delIDs")

	d := testutils.MakeData(10)

	ids := []int{}
	for i := 0; i < 10; i++ {
		ids = append(ids, i)
	}

	b.Insert(d...)
	b.DeleteID(ids[0], ids[len(ids)-1])
	readData, _ := b.Read(nil, true, testutils.MatchAll[int]{})

	expectedData := ids[1:len(ids)-1]

	if len(readData) != len(expectedData) {
		t.Errorf("Got bad read id len, exp: %v, got %v", len(expectedData), len(readData))
	}

	for i, v := range readData {
		e, r := expectedData[i], v.GetID()
		if r != e {
			t.Errorf("Unexpected data at %v: expected %v, got %v", i, e, r)
		}
	}
}

func TestFileDelete(t *testing.T) {
	i := 0
	testutils.GenerateGenericQueryTest(500, t, func(newestIsBiggest bool) subdb.BackendWithEverything[int] {
		b := newBackend("del" + fmt.Sprint(i), newestIsBiggest)
		i++
		return b
	}, testutils.GenericDeleteQueryTest)

	for j := 0; j < i; j++ {
		cleanData("del" + fmt.Sprint(j))
	}
}