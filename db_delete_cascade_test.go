package structdbpostgres

import (
	"testing"
)

/*
	max DelParent.Delete*() depth -> |
	                                   |

DelParent --> []DelChildNone                        |

	          |-> []DelChildDelete --> []DelChildNone   |
						|	                   |-> []DelChildDelete +--> []DelChildDelete
						|									   |                    |`-> []DelChildUpdate
						|									   `-> []DelChildUpdate |
	          `-> []DelChildUpdate --> []DelChildNone   |
							                   |-> []DelChildDelete |
															   `-> []DelChildUpdate +--> []DelChildDelete
																                      |`-> []DelChildUpdate
*/
type DelParent struct {
	ID    int64
	Flags int64
	Name  string
	// Expect all guys below to have a field DelParentID
	// no action should be taken for the 'None' one
	DelChildrenNone []*DelChildNone
	// all these guys whose parent is this instance (so DelParentID) should be deleted from the database
	DelChildrenDelete []*DelChildDelete `2db:"on_del:del"`
	// all these guys whose parent is this instance should have the field DelParentID updated to 0
	// (because sometimes we do not want to remove, more update options will be added later)
	DelChildrenUpdate []*DelChildUpdate `2db:"on_del:upd del_upd_field:DelParentID del_upd_val:0"`
}

type DelChildNone struct {
	ID          int64
	DelParentID int64
}
type DelChildDelete struct {
	ID          int64
	DelParentID int64
	// "cdel" stands for cascade delete (not "del")
	// if this struct has children they are linked by the DelParentID column which is specified in "cdel_field"
	DelChildrenDelete []*DelChildDelete `2db:"on_cdel:del cdel_field:DelParentID"`
	DelChildrenUpdate []*DelChildUpdate `2db:"on_cdel:upd cdel_field:DelParentID cdel_upd_field:DelChildDeleteID cdel_upd_val:0"`
}
type DelChildUpdate struct {
	ID                int64
	DelParentID       int64
	DelChildrenDelete []*DelChildDelete `2db:"on_cdel:del cdel_field:DelParentID"`
	DelChildrenUpdate []*DelChildUpdate `2db:"on_cdel:upd cdel_field:DelParentID cdel_upd_field:DelChildDeleteID cdel_upd_val:0"`
}

func TestDeleteCascade(t *testing.T) {
	// Create a test parent (with children) and get its ID
	p := createTestDelParentWithChildren()

	// Check if test objects are added properly
	var cnt int
	err2 := dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_parents").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt != 1 {
		t.Fatalf("Number of objects in the database before running the test is invalid")
	}
	err2 = dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_child_nones WHERE del_child_none_id IN (1, 2, 111, 121, 211, 221, 112, 122, 212, 222) AND del_parent_id != 0").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt != 10 {
		t.Fatalf("Number of objects in the database before running the test is invalid")
	}
	err2 = dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_child_deletes WHERE del_child_delete_id IN (1, 2, 111, 121, 211, 221, 112, 122, 212, 222, 1001, 1003) AND del_parent_id != 0").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt != 12 {
		t.Fatalf("Number of objects in the database before running the test is invalid")
	}
	err2 = dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_child_updates WHERE del_child_update_id IN (1, 2, 111, 121, 211, 221, 112, 122, 212, 222, 1002, 1004) AND del_parent_id != 0").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt != 12 {
		t.Fatalf("Number of objects in the database before running the test is invalid")
	}

	// Delete the parent object
	err1 := testController.Delete(p, DeleteOptions{
		Constructors: map[string]func() interface{}{
			"DelChildNone":   func() interface{} { return &DelChildNone{} },
			"DelChildDelete": func() interface{} { return &DelChildDelete{} },
			"DelChildUpdate": func() interface{} { return &DelChildUpdate{} },
		},
	})
	if err1 != nil {
		t.Fatalf("Failed to run Delete successfully: %s", err1.Err.Error())
	}

	// Check things
	// 1 should be removed
	err2 = dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_parents").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt > 0 {
		t.Fatalf("Delete failed to remove parent object")
	}

	// 1, 2 should exist in del_child_nones
	// 111, 121, 211, 221 should exist in del_child_nones
	// 112, 122, 212, 222 should exist in del_child_nones
	err2 = dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_child_nones WHERE del_child_none_id IN (1, 2, 111, 121, 211, 221, 112, 122, 212, 222)").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt != 10 {
		t.Fatalf("Delete failed to not remove object that were not meant to be removed")
	}

	// 1, 2 should not exist in del_child_deletes
	// 111, 121 should not exist in del_child_deletes
	// 112, 122 should not exist in del_child_deletes
	err2 = dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_child_deletes WHERE del_child_delete_id IN (1, 2, 111, 121, 112, 122)").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt > 0 {
		t.Fatalf("Delete failed to remove children")
	}

	// 211, 221, 212, 222, 1001, 1003 should exist in del_child_deletes
	err2 = dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_child_deletes WHERE del_child_delete_id IN (211, 221, 212, 222, 1001, 1003)").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt != 6 {
		t.Fatalf("Delete failed to not remove children")
	}

	// 1, 2 should exist in del_child_updates and their del_parent_id should be updated to 0
	// 111, 121 should exist in del_child_updates and have their del_parent_id updated to 0
	// 112, 122 should exist in del_child_updates and have their del_parent_id updated to 0
	err2 = dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_child_updates WHERE del_child_update_id IN (1, 2, 111, 121, 112, 122) AND del_parent_id=0").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt != 6 {
		t.Fatalf("Delete failed to update children")
	}

	// 211, 221, 212, 222, 1002, 1004 should exist in del_child_updates and have their del_parent_id not updated to 0
	err2 = dbConn.QueryRow("SELECT COUNT(*) FROM struct2db_del_child_updates WHERE del_child_update_id IN (211, 221, 212, 222, 1002, 1004) AND del_parent_id!=0").Scan(&cnt)
	if err2 != nil {
		t.Fatalf("Failed to select count: %s", err2.Error())
	}
	if cnt != 6 {
		t.Fatalf("Delete failed to not update children")
	}
}

func createTestDelParentWithChildren() interface{} {
	recreateTestDelTables()

	// create DelParent
	p := &DelParent{Name: "Parent1", ID: 1}
	testController.Save(p, SaveOptions{})

	// create children
	for i := 0; i < 2; i++ {
		cNone := &DelChildNone{DelParentID: p.ID, ID: int64(i + 1)}     // 1,2
		cDelete := &DelChildDelete{DelParentID: p.ID, ID: int64(i + 1)} // 1,2
		cUpdate := &DelChildUpdate{DelParentID: p.ID, ID: int64(i + 1)} // 1,2
		testController.Save(cNone, SaveOptions{})
		testController.Save(cDelete, SaveOptions{})
		testController.Save(cUpdate, SaveOptions{})
	}

	// create grandchildren
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			cDeleteNone := &DelChildNone{DelParentID: int64(i + 1), ID: int64(((i + 1) * 100) + (j+1)*10 + 1)}     // 111, 121, 211, 221
			cDeleteDelete := &DelChildDelete{DelParentID: int64(i + 1), ID: int64(((i + 1) * 100) + (j+1)*10 + 1)} // 111, 121, 211, 221
			cDeleteUpdate := &DelChildUpdate{DelParentID: int64(i + 1), ID: int64(((i + 1) * 100) + (j+1)*10 + 1)} // 111, 121, 211, 221
			cUpdateNone := &DelChildNone{DelParentID: int64(i + 1), ID: int64(((i + 1) * 100) + (j+1)*10 + 2)}     // 112, 122, 212, 222
			cUpdateDelete := &DelChildDelete{DelParentID: int64(i + 1), ID: int64(((i + 1) * 100) + (j+1)*10 + 2)} // 112, 122, 212, 222
			cUpdateUpdate := &DelChildUpdate{DelParentID: int64(i + 1), ID: int64(((i + 1) * 100) + (j+1)*10 + 2)} // 112, 122, 212, 222
			testController.Save(cDeleteNone, SaveOptions{})
			testController.Save(cDeleteDelete, SaveOptions{})
			testController.Save(cDeleteUpdate, SaveOptions{})
			testController.Save(cUpdateNone, SaveOptions{})
			testController.Save(cUpdateDelete, SaveOptions{})
			testController.Save(cUpdateUpdate, SaveOptions{})
		}
	}

	// create grandgrandchildren
	cDeleteDeleteDelete := &DelChildDelete{DelParentID: int64(111), ID: 1001}
	cDeleteDeleteUpdate := &DelChildUpdate{DelParentID: int64(111), ID: 1002}
	cUpdateUpdateDelete := &DelChildDelete{DelParentID: int64(112), ID: 1003}
	cUpdateUpdateUpdate := &DelChildUpdate{DelParentID: int64(112), ID: 1004}
	testController.Save(cDeleteDeleteDelete, SaveOptions{})
	testController.Save(cDeleteDeleteUpdate, SaveOptions{})
	testController.Save(cUpdateUpdateUpdate, SaveOptions{})
	testController.Save(cUpdateUpdateDelete, SaveOptions{})

	return p
}

func recreateTestDelTables() {
	testController.DropTable(&DelParent{})
	testController.DropTable(&DelChildNone{})
	testController.DropTable(&DelChildDelete{})
	testController.DropTable(&DelChildUpdate{})
	testController.CreateTable(&DelParent{})
	testController.CreateTable(&DelChildNone{})
	testController.CreateTable(&DelChildDelete{})
	testController.CreateTable(&DelChildUpdate{})
}
