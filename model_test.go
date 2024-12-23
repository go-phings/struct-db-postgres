package structdbpostgres

import (
	"log"
	"testing"
)

// TestGetObjIDInterface tests if GetObjIDInterface return pointer to ID field
func TestGetObjIDInterface(t *testing.T) {
	ts := &TestStruct{}
	ts.ID = 123
	i := testController.GetObjIDInterface(ts)
	if *(i.(*int64)) != int64(123) {
		log.Fatalf("GetObjIDInterface failed to get ID")
	}
}

// TestGetObjIDValue tests if GetObjIDValue returns values of the ID field
func TestGetObjIDValue(t *testing.T) {
	ts := &TestStruct{}
	ts.ID = 123
	v := testController.GetObjIDValue(ts)
	if v != 123 {
		log.Fatalf("GetObjIDValue failed to get ID")
	}
}

// TestObjFieldInterfaces tests if GetObjFieldInterfaces returns interfaces to object fields
func TestObjFieldInterfaces(t *testing.T) {
	// TODO
}

// TestResetFields tests if ResetFields zeroes object fields
func TestResetFields(t *testing.T) {
	// TODO
}
