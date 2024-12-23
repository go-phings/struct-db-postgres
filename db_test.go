package structdbpostgres

import (
	"fmt"
	"time"
)

// Test struct for all the tests
type TestStruct struct {
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

func getTestStructWithData() *TestStruct {
	ts := &TestStruct{}
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

func recreateTestStructTable() {
	testController.DropTable(&TestStruct{})
	testController.CreateTable(&TestStruct{})
}

func areTestStructObjectsSame(ts1 *TestStruct, ts2 *TestStruct) bool {
	if ts1.Flags != ts2.Flags {
		return false
	}
	if ts1.PrimaryEmail != ts2.PrimaryEmail {
		return false
	}
	if ts1.EmailSecondary != ts2.EmailSecondary {
		return false
	}
	if ts1.FirstName != ts2.FirstName {
		return false
	}
	if ts1.LastName != ts2.LastName {
		return false
	}
	if ts1.Age != ts2.Age {
		return false
	}
	if ts1.Price != ts2.Price {
		return false
	}
	if ts1.PostCode != ts2.PostCode {
		return false
	}
	if ts1.PostCode2 != ts2.PostCode2 {
		return false
	}
	if ts1.Password != ts2.Password {
		return false
	}
	if ts1.CreatedByUserID != ts2.CreatedByUserID {
		return false
	}
	if ts1.Key != ts2.Key {
		return false
	}
	return true
}
