package structdbpostgres

import (
	"testing"
)

// TestGetCount tests if Get properly gets count of objects from the database, filtered
func TestGetCount(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should be ignored by GetCount later on
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.Price = 444
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts, SaveOptions{})
	}

	// Insert data that should be selected by filters
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts, SaveOptions{})
	}

	// Get the data from the database
	cnt, err := testController.GetCount(func() interface{} {
		return &TestStruct{}
	}, GetCountOptions{
		Filters: map[string]interface{}{"Price": 444, "PrimaryEmail": "primary@example.com"},
	})
	if err != nil {
		t.Fatalf("Get failed to return list of objects: %s", err.Op)
	}
	if cnt != 150 {
		t.Fatalf("Get failed to return list of objects, want %v, got %v", 150, cnt)
	}
}

// TestGetCountWithRawQuery tests if Get properly gets count of objects from the database, filtered, with additional raw query
func TestGetCountWithRawQuery(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should be ignored by GetCount later on
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.Price = 444
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts, SaveOptions{})
	}

	// Insert data that should be selected by filters
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		testController.Save(ts, SaveOptions{})
	}

	// Get the data from the database
	cnt, err := testController.GetCount(func() interface{} {
		return &TestStruct{}
	}, GetCountOptions{
		Filters: map[string]interface{}{
			"Price":        444,
			"PrimaryEmail": "primary@example.com",
			"_raw": []interface{}{
				".PrimaryEmail = ? AND .Age IN (?)",
				"another@example.com",
				[]int{32, 33, 34},
			},
			"_rawConjuction": RawConjuctionOR,
		},
	})
	if err != nil {
		t.Fatalf("Get failed to return list of objects: %s", err.Op)
	}
	if cnt != 153 {
		t.Fatalf("Get failed to return list of objects, want %v, got %v", 153, cnt)
	}
}
