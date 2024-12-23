package structdbpostgres

import (
	"testing"
)

// TestDeleteMultiple tests if DeleteMultiple removes objects from database based on specified filters
func TestDeleteMultiple(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should not be removed
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts, SaveOptions{})
	}

	// Insert data that should be deleted
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts, SaveOptions{})
	}

	// Delete multiple rows from the database
	err := testController.DeleteMultiple(&TestStruct{}, DeleteMultipleOptions{
		Filters: map[string]interface{}{"Price": 444, "PrimaryEmail": "primary@example.com"},
	})
	if err != nil {
		t.Fatalf("DeleteMultiple failed to delete objects: %s", err.Op)
	}

	cnt, _ := testController.GetCount(func() interface{} { return &TestStruct{} }, GetCountOptions{})
	if cnt != 50 {
		t.Fatalf("DeleteMultiple removed invalid number of rows, there are %d rows left, instead of %d", cnt, 50)
	}
}

// TestDeleteMultipleWithRawQuery tests if DeleteMultiple removes objects from database based on specified filters,
// and a condition which is somewhat raw query
func TestDeleteMultipleWithRawQuery(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should not be removed
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts, SaveOptions{})
	}

	// Insert data that should be deleted
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts, SaveOptions{})
	}

	// Delete multiple rows from the database
	err := testController.DeleteMultiple(&TestStruct{}, DeleteMultipleOptions{
		Filters: map[string]interface{}{
			"Price":        444,
			"PrimaryEmail": "primary@example.com",
			"_raw": []interface{}{
				".Age = ? OR .Age IN (?) OR (.Age = ? AND .PrimaryEmail = ?)",
				31,
				[]int{32, 33, 34},
				35,
				"miko@example.com",
			},
			"_rawConjuction": RawConjuctionOR,
		},
	})
	if err != nil {
		t.Fatalf("DeleteMultiple failed to delete objects: %s", err.Op)
	}

	cnt, _ := testController.GetCount(func() interface{} { return &TestStruct{} }, GetCountOptions{})
	if cnt != 46 {
		t.Fatalf("DeleteMultiple removed invalid number of rows, there are %d rows left, instead of %d", cnt, 46)
	}
}

// TestDeleteMultipleWithRawQueryOnly tests if DeleteMultiple removes objects from database based on a condition which is somewhat raw query
func TestDeleteMultipleWithRawQueryOnly(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should not be removed
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts, SaveOptions{})
	}

	// Insert data that should be deleted
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts, SaveOptions{})
	}

	// Delete multiple rows from the database
	err := testController.DeleteMultiple(&TestStruct{}, DeleteMultipleOptions{
		Filters: map[string]interface{}{
			"_raw": []interface{}{
				"(.Price = ? AND .PrimaryEmail = ?) OR (.Age = ? OR .Age IN (?) OR (.Age = ? AND .PrimaryEmail = ?))",
				444,
				"primary@example.com",
				31,
				[]int{32, 33, 34},
				35,
				"miko@example.com",
			},
			"_rawConjuction": RawConjuctionOR,
		},
	})
	if err != nil {
		t.Fatalf("DeleteMultiple failed to delete objects: %s", err.Op)
	}

	cnt, _ := testController.GetCount(func() interface{} { return &TestStruct{} }, GetCountOptions{})
	if cnt != 46 {
		t.Fatalf("DeleteMultiple removed invalid number of rows, there are %d rows left, instead of %d", cnt, 46)
	}
}
