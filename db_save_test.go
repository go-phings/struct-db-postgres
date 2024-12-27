package structdbpostgres

import "testing"

// TestSave tests if Save properly inserts and updates object in the database
func TestSave(t *testing.T) {
	recreateTestStructTable()

	// Create an object in the database first
	ts := getTestStructWithData()
	err := testController.Save(ts, SaveOptions{})
	if err != nil {
		t.Fatalf("Save failed to insert struct to the table: %s", err.(ErrController).Op)
	}

	ts2 := &TestStruct{}
	ts2FieldsPtrs := []interface{}{&ts2.ID}
	ts2FieldsPtrs = append(ts2FieldsPtrs, testController.GetObjFieldInterfaces(ts2, false)...)
	err2 := dbConn.QueryRow("SELECT * FROM struct2db_test_structs ORDER BY test_struct_id DESC LIMIT 1").Scan(ts2FieldsPtrs...)
	if err2 != nil {
		t.Fatalf("Save failed to insert struct to the table: %s", err2.Error())
	}

	if ts2.ID == 0 || ts2.Flags != ts.Flags || ts2.PrimaryEmail != ts.PrimaryEmail || ts2.EmailSecondary != ts.EmailSecondary || ts2.FirstName != ts.FirstName || ts2.LastName != ts.LastName || ts2.Age != ts.Age || ts2.Price != ts.Price || ts2.PostCode != ts.PostCode || ts2.PostCode2 != ts.PostCode2 || ts2.CreatedByUserID != ts.CreatedByUserID || ts2.Key != ts.Key || ts2.Password != ts.Password {
		t.Fatalf("Save failed to insert struct to the table")
	}

	// Now, update the object in the database
	ts.Flags = 7
	ts.PrimaryEmail = "primary1@example.com"
	ts.EmailSecondary = "secondary2@example.com"
	ts.FirstName = "Johnny"
	ts.LastName = "Smithsy"
	ts.Age = 50
	ts.Price = 222
	ts.PostCode = "22-222"
	ts.PostCode2 = "33-333"
	ts.Password = "xxx"
	ts.CreatedByUserID = 7
	ts.Key = "123456789012345678901234567890aaa"

	err3 := testController.Save(ts, SaveOptions{})
	if err3 != nil {
		t.Fatalf("Save failed to update struct")
	}

	ts2 = &TestStruct{}
	ts2FieldsPtrs = []interface{}{&ts2.ID}
	ts2FieldsPtrs = append(ts2FieldsPtrs, testController.GetObjFieldInterfaces(ts2, false)...)
	err2 = dbConn.QueryRow("SELECT * FROM struct2db_test_structs ORDER BY test_struct_id DESC LIMIT 1").Scan(ts2FieldsPtrs...)
	if err2 != nil {
		t.Fatalf("Save failed to update struct in the table: %s", err2.Error())
	}

	if ts2.ID == 0 {
		t.Fatalf("Save failed to update struct in the table")
	}
}

// TestSaveInsertWithID tests if an element with provided ID will be inserted
func TestSaveInsertWithID(t *testing.T) {
	recreateTestStructTable()

	ts := getTestStructWithData()
	ts.ID = 99999
	ts.FirstName = "ProvidedID"
	err := testController.Save(ts, SaveOptions{})
	if err != nil {
		t.Fatalf("Save failed to insert struct with provided ID to the table: %s", err.(ErrController).Op)
	}

	ts2 := &TestStruct{}
	ts2FieldsPtrs := testController.GetObjFieldInterfaces(ts2, true)
	err2 := dbConn.QueryRow("SELECT * FROM struct2db_test_structs WHERE test_struct_id=99999 ORDER BY test_struct_id DESC LIMIT 1").Scan(ts2FieldsPtrs...)
	if err2 != nil {
		t.Fatalf("Save failed to insert struct with provided ID in the table: %s", err2.Error())
	}

	if ts2.ID != 99999 || ts2.FirstName != ts.FirstName {
		t.Fatalf("Save failed to insert struct with provided ID in the table")
	}

	// update that object
	ts2.FirstName = "UpdatedProvidedID"
	err = testController.Save(ts2, SaveOptions{})
	if err != nil {
		t.Fatalf("Save failed to update struct previously inserted with provided ID to the table: %s", err.(ErrController).Op)
	}

	ts3 := &TestStruct{}
	ts3FieldsPtrs := testController.GetObjFieldInterfaces(ts3, true)
	err2 = dbConn.QueryRow("SELECT * FROM struct2db_test_structs WHERE test_struct_id=99999 ORDER BY test_struct_id DESC LIMIT 1").Scan(ts3FieldsPtrs...)
	if err2 != nil {
		t.Fatalf("Save failed to update struct previously inserted with provided ID in the table: %s", err2.Error())
	}
	if ts3.ID != 99999 || ts3.FirstName != ts2.FirstName {
		t.Fatalf("Save failed to update struct previously inserted with provided ID in the table")
	}
}

// TestSaveInsertWithIDAndNoInsert tests if an element with provided ID is not inserted when NoInsert is true
func TestSaveInsertWithIDAndNoInsert(t *testing.T) {
	recreateTestStructTable()

	ts := getTestStructWithData()
	ts.ID = 99999
	ts.FirstName = "ProvidedID"
	err := testController.Save(ts, SaveOptions{
		NoInsert: true,
	})
	if err != nil {
		t.Fatalf("Save failed to not insert struct with provided ID to the table when NoInsert: %s", err.(ErrController).Op)
	}

	var cnt int
	err2 := dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_test_structs WHERE test_struct_id=99999").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Save failed to not insert struct with provided ID in the table when NoInsert: %s", err2.Error())
	}

	if cnt > 0 {
		t.Fatalf("Save failed to not insert struct with provided ID in the table when NoInsert")
	}
}
