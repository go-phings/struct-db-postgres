package structdbpostgres

import "testing"

// TestUpdateMultiple tests if UpdateMultiple update objects from database based on specified filters
func TestUpdateMultiple(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should not be removed
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts, SaveOptions{})
	}

	// Insert data that should be updated
	for i := 1; i < 151; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30
		ts.PrimaryEmail = "changeme@example.com"
		testController.Save(ts, SaveOptions{})
	}

	// Update multiple rows from the database
	err := testController.UpdateMultiple(&TestStruct{}, map[string]interface{}{
		"PrimaryEmail": "newemail@example.com",
		"Age":          98,
	},
		UpdateMultipleOptions{
			Filters: map[string]interface{}{
				"Price":        444,
				"PrimaryEmail": "changeme@example.com",
			},
		})
	if err != nil {
		t.Fatalf("UpdateMultiple failed to update objects: %s", err.Op)
	}

	cnt, _ := testController.GetCount(func() interface{} { return &TestStruct{} }, GetCountOptions{
		Filters: map[string]interface{}{
			"PrimaryEmail": "newemail@example.com",
			"Age":          98,
		},
	})
	if cnt != 150 {
		t.Fatalf("UpdateMultiple updated invalid number of rows, there are %d rows left, instead of %d", cnt, 150)
	}
}
