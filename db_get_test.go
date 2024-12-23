package structdbpostgres

import (
	"fmt"
	"html"
	"reflect"
	"strings"
	"testing"
)

// TestGet tests if Get properly gets many objects from the database, filtered and ordered, with results limited to specific number
func TestGet(t *testing.T) {
	recreateTestStructTable()

	// Insert some data that should be ignored by Get later on
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 10 + i
		ts.Price = 444
		ts.PrimaryEmail = "another@example.com"
		testController.Save(ts, SaveOptions{})
	}

	// Insert data that should be selected by filters
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30 + i
		testController.Save(ts, SaveOptions{})
	}

	// Get the data from the database
	testStructs, err := testController.Get(func() interface{} {
		return &TestStruct{}
	}, GetOptions{
		Order:   []string{"Age", "asc", "Price", "asc"},
		Limit:   10,
		Offset:  20,
		Filters: map[string]interface{}{"Price": 444, "PrimaryEmail": "primary@example.com"},
	})
	if err != nil {
		t.Fatalf("Get failed to return list of objects: %s", err.Op)
	}
	if len(testStructs) != 10 {
		t.Fatalf("Get failed to return list of objects, want %v, got %v", 10, len(testStructs))
	}
	if testStructs[2].(*TestStruct).Age != 53 {
		t.Fatalf("Get failed to return correct list of objects, want %v, got %v", 53, testStructs[2].(*TestStruct).Age)
	}
}

// TestGetWithoutFilters tests if Get properly gets many objects from the database, without any filters
func TestGetWithoutFilters(t *testing.T) {
	recreateTestStructTable()

	// Insert data to the database
	for i := 1; i < 51; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30 + i
		testController.Save(ts, SaveOptions{})
	}

	// Get the data
	testStructs, err := testController.Get(func() interface{} {
		return &TestStruct{}
	}, GetOptions{
		Order:  []string{"Age", "asc", "Price", "asc"},
		Limit:  13,
		Offset: 14,
	})
	if err != nil {
		t.Fatalf("Get failed to return list of objects: %s", err.Op)
	}
	if len(testStructs) != 13 {
		t.Fatalf("Get failed to return list of objects, want %v, got %v", 10, len(testStructs))
	}
	if testStructs[2].(*TestStruct).Age != 47 {
		t.Fatalf("Get failed to return correct list of objects, want %v, got %v", 47, testStructs[2].(*TestStruct).Age)
	}
}

// TestGetWithRowObjTransformFunc tests if Get can properly return a list of custom elements (eg. string)
// where each object (row from the database) is transform with a specific function
func TestGetWithRowObjTransformFunc(t *testing.T) {
	recreateTestStructTable()

	// Insert data to the database
	for i := 1; i < 3; i++ {
		ts := getTestStructWithData()
		ts.ID = 0
		ts.Age = 30 + i
		ts.FirstName = fmt.Sprintf("%s %d", ts.FirstName, i)
		testController.Save(ts, SaveOptions{})
	}

	// Get the data
	testCustomList, err := testController.Get(func() interface{} {
		return &TestStruct{}
	}, GetOptions{
		Order: []string{"Age", "asc"},
		RowObjTransformFunc: func(obj interface{}) interface{} {
			out := "<tr>"

			v := reflect.ValueOf(obj)
			elem := v.Elem()
			i := reflect.Indirect(v)
			s := i.Type()
			for j := 0; j < s.NumField(); j++ {
				out += "<td>"
				field := s.Field(j)
				fieldType := field.Type.Kind()
				if fieldType == reflect.String {
					out += html.EscapeString(elem.Field(j).String())
				}
				if fieldType == reflect.Bool {
					out += fmt.Sprintf("%v", elem.Field(j).Bool())
				}
				if fieldType == reflect.Int || fieldType == reflect.Int64 {
					out += fmt.Sprintf("%d", elem.Field(j).Int())
				}
				out += "</td>"
			}

			out += "</tr>"

			return out
		},
	})
	if err != nil {
		t.Fatalf("Get failed to return list of objects modified with transform func: %s", err.Op)
	}
	if len(testCustomList) != 2 {
		t.Fatalf("Get with transform func returned invalid number of objects, wanted %d got %d", 2, len(testCustomList))
	}
	// One of the columns is a random number so testing just the beginning
	if !strings.HasPrefix(testCustomList[0].(string), "<tr><td>1</td><td>4</td><td>primary@example.com</td><td>secondary@example.com</td><td>John 1</td><td>Smith</td><td>31</td><td>444</td><td>00-000</td><td>11-111</td><td>yyy</td><td>4</td><td>") || !strings.HasPrefix(testCustomList[1].(string), "<tr><td>2</td><td>4</td><td>primary@example.com</td><td>secondary@example.com</td><td>John 2</td><td>Smith</td><td>32</td><td>444</td><td>00-000</td><td>11-111</td><td>yyy</td><td>4</td><td>") {
		t.Fatalf("Get with transform func returned invalid objects")
	}
}
