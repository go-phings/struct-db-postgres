package structdbpostgres

import (
	"fmt"
	"testing"
)

// TestDelete tests if Delete removes object from the database
func TestDelete(t *testing.T) {
	recreateTestStructTable()

	// Insert an object first
	ts := getTestStructWithData()
	testController.Save(ts, SaveOptions{})

	// Delete it
	err := testController.Delete(ts, DeleteOptions{})
	if err != nil {
		t.Fatalf("Delete failed to remove: %s", err.Op)
	}

	var cnt int64
	err2 := dbConn.QueryRow(fmt.Sprintf("SELECT COUNT(*) AS c FROM struct2db_test_structs WHERE test_struct_id = %d", ts.ID)).Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Delete failed to delete struct from the table")
	}
	if cnt > 0 {
		t.Fatalf("Delete failed to delete struct from the table")
	}
	if ts.ID != 0 {
		t.Fatalf("Delete failed to set ID to 0 on the struct")
	}
}
