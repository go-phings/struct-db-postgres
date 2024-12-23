package structdbpostgres

import (
	"fmt"
	"testing"
	"time"
)

// Test struct for validation tests
type ValidationTestStruct struct {
	ID    int64 `json:"test_struct_id"`
	Flags int64 `json:"test_struct_flags"`

	// Test email validation
	PrimaryEmail   string `json:"email" 2db:"req"`
	EmailSecondary string `json:"email2" 2db:"req email"`

	// Test length validation
	FirstName string `json:"first_name" 2db:"req lenmin:2 lenmax:30"`
	LastName  string `json:"last_name" 2db:"req lenmin:0 lenmax:255"`

	// Test int value validation
	Age   int `json:"age" 2db:"valmin:1 valmax:120"`
	Price int `json:"price" 2db:"valmin:0 valmax:999"`

	// Test regular expression
	PostCode  string `json:"post_code" 2db:"req lenmin:6 regexp:^[0-9]{2}\\-[0-9]{3}$"`
	PostCode2 string `json:"post_code2" 2db:"lenmin:6" 2db_regexp:"^[0-9]{2}\\-[0-9]{3}$"`

	// Some other fields
	Password        string `json:"password"`
	CreatedByUserID int64  `json:"created_by_user_id"`

	// Test unique tag
	Key string `json:"key" 2db:"req uniq lenmin:30 lenmax:255"`
}

func getValidationTestStructWithData() *ValidationTestStruct {
	ts := &ValidationTestStruct{}
	ts.Flags = 4
	ts.PrimaryEmail = "primary@example.com"
	ts.EmailSecondary = "secondary@example.com"
	ts.FirstName = "John"
	ts.LastName = "Smith"
	ts.Age = 37
	ts.Price = 444
	ts.PostCode = "00-000"
	ts.PostCode2 = "11-111"
	ts.Password = "yyy"
	ts.CreatedByUserID = 4
	ts.Key = fmt.Sprintf("12345679012345678901234567890%d", time.Now().UnixNano())
	return ts
}

// TestValidateWithValidStruct tests if Validate successfully validates object with valid values
func TestValidateWithValidStruct(t *testing.T) {
	ts := getValidationTestStructWithData()
	b, failedFields, err := testController.Validate(ts, nil)
	if !b {
		t.Fatalf("Validate failed validate valid struct")
	}
	if len(failedFields) > 0 {
		t.Fatalf("Validate return non-empty failed field list when validating a valid struct")
	}
	if err != nil {
		t.Fatalf("Validate failed to validate valid struct: %s", err.Error())
	}
}

// TestValidateWithValidStructAndListOfFields tests if Validate successfully validates object with valid values
func TestValidateWithValidStructAndListOfFields(t *testing.T) {
	ts := getValidationTestStructWithData()
	ts.Age = 0
	ts.FirstName = "x"
	ts.Key = "tooshort"
	ts.PrimaryEmail = "thisis@valid.email.com"
	b, failedFields, err := testController.Validate(ts, map[string]interface{}{
		"PrimaryEmail": true,
		"Price":        true,
	})
	if !b {
		t.Fatalf("Validate failed to validate listed fields")
	}
	if len(failedFields) > 0 {
		t.Fatalf("Validate return non-empty failed field list when validating listed fields")
	}
	if err != nil {
		t.Fatalf("Validate failed to validate listed fields: %s", err.Error())
	}
}

// TestValidateWithInvalidStruct tests if Validate invalidates object with invalid values
func TestValidateWithInvalidStruct(t *testing.T) {
	ts := getValidationTestStructWithData()
	ts.PrimaryEmail = "invalidemail"
	ts.EmailSecondary = "invalidemail"
	ts.FirstName = "x"
	ts.LastName = "aFbdsZFYxMpUNKCkBrHhhODrMBEHtmRAJjoqSSfUotvsfMXcJGPrCRaDOsyuyrXYfACjsJEMUoxNvTwRMUaWYruOxgzTXJRzobmxaFbdsZFYxMpUNKCkBrHhhODrMBEHtmRAJjoqSSfUotvsfMXcJGPrCRaDOsyuyrXYfACjsJEMUoxNvTwRMUaWYruOxgzTXJRzobmxaFbdsZFYxMpUNKCkBrHhhODrMBEHtmRAJjoqSSfUotvsfMXcJGPrCRaDOsyuyrXYfACjsJEMUoxNvTwRMUaWYruOxgzTXJRzobmxaFbdsZFYxMpUNKCkBrHhhODrMBEHtmRAJjoqSSfUotvsfMXcJGPrCRaDOsyuyrXYfACjsJEMUoxNvTwRMUaWYruOxgzTXJRzobmx"
	ts.Age = 0
	ts.Price = 1000
	ts.PostCode = "inv"
	ts.PostCode2 = "inv"
	ts.Key = "tooshort"
	b, failedFields, err := testController.Validate(ts, nil)
	if err != nil {
		t.Fatalf("Validate failed with an err")
	}
	if b {
		t.Fatalf("Validate failed to return false for struct with invalid field values")
	}
	for _, f := range []string{"PrimaryEmail", "EmailSecondary", "FirstName", "LastName", "Age", "Price", "PostCode", "PostCode2", "Key"} {
		if failedFields[f] == 0 {
			t.Fatalf(fmt.Sprintf("Validate failed to return field %s in failed fields", f))
		}
	}
}
