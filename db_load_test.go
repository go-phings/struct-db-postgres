package structdbpostgres

import (
	"fmt"
	"testing"
)

// TestLoad tests if Load properly gets row from the database table and populate object fields with its value
func TestLoad(t *testing.T) {
	recreateTestStructTable()

	// Insert an object first
	ts := getTestStructWithData()
	testController.Save(ts, SaveOptions{})

	// Get the object
	ts2 := &TestStruct{}
	err := testController.Load(ts2, fmt.Sprintf("%d", ts.ID), LoadOptions{})
	if err != nil {
		t.Fatalf("Load failed to get data: %s", err.Op)
	}

	if !areTestStructObjectsSame(ts, ts2) {
		t.Fatalf("Load failed to set struct with data: %s", err.Op)
	}
}
