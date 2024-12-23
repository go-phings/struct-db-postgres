package structdbpostgres

import (
	"testing"
)

type TableTestStruct struct {
	ID    int64 `json:"db_table_struct_id"`
	Flags int64 `json:"db_table_struct_flags"`
}

// TestCreateTables tests if CreateTables creates tables in the database
func TestCreateTables(t *testing.T) {
	st := &TableTestStruct{}
	err := testController.CreateTables(st)
	if err != nil {
		t.Fatalf("CreateTables failed to create table for a struct: %s", err.Error())
	}

	cnt, err2 := getTableNameCnt("struct2db_table_test_structs")
	if err2 != nil {
		t.Fatalf("CreateTables failed to create table for a struct: %s", err2.Error())
	}
	if cnt == 0 {
		t.Fatalf("CreateTables failed to create the table")
	}
}

// TestDropTables tests if DropTables drops tables in the database
func TestDropTables(t *testing.T) {
	st := &TableTestStruct{}
	err := testController.DropTables(st)
	if err != nil {
		t.Fatalf("DropTables failed to drop table for a struct: %s", err.Error())
	}

	cnt, err2 := getTableNameCnt("struct2db_table_test_structs")
	if err2 != nil {
		t.Fatalf("DropTables failed to drop table for a struct: %s", err2.Error())
	}
	if cnt != 0 {
		t.Fatalf("DropTables failed to drop the table")
	}
}
