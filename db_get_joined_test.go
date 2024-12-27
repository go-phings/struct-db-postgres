package structdbpostgres

import (
	"fmt"
	"testing"
)

type ProductKind struct {
	ID   int64
	Name string
}

type ProductGroup struct {
	ID          int64
	Name        string
	Description string
	Code        string
}

type Product struct {
	ID            int64
	Name          string
	Price         int
	ProductKindID int64
	ProductGrpID  int64
}

type Product_WithDetails struct {
	ID               int64
	Name             string
	Price            int
	ProductKindID    int64
	ProductGrpID     int64
	ProductKind      *ProductKind `2db:"join"`
	ProductKind_Name string
	ProductGrp       *ProductGroup `2db:"join"`
	ProductGrp_Code  string
}

// Scenarios to test:
// * Load()
// * Get()
// There is no plan on making Save() at this point

func TestJoinedLoad(t *testing.T) {
	createTestJoinedStructs()

	p := &Product_WithDetails{}
	err := testController.Load(p, fmt.Sprintf("%d", 6), LoadOptions{})
	if err != nil {
		t.Fatalf("Load failed to get data for struct with other joined structs: %s", err.(ErrController).Err.Error())
	}

	if p.Name != "Product Name" || p.Price != 1234 || p.ProductKindID != 33 || p.ProductGrpID != 113 {
		t.Fatalf("Load failed to set struct fields: %s", err.(ErrController).Err.Error())
	}

	if p.ProductKind_Name != "Kind 1" || p.ProductGrp_Code != "GRP1" {
		t.Fatalf("Load failed to set 'joined' fields: %s", err.(ErrController).Op)
	}
}

func TestJoinedGet(t *testing.T) {
	createTestJoinedStructs()

	ps, err := testController.Get(func() interface{} {
		return &Product_WithDetails{}
	}, GetOptions{
		Order:  []string{"ID", "asc"},
		Limit:  22,
		Offset: 0,
		Filters: map[string]interface{}{
			"Name":             "Product Name",
			"ProductKind_Name": "Kind 1",
			"ProductGrp_Code":  "GRP1",
		},
	})
	if err != nil {
		t.Fatalf("Get failed to return list of joined structs: %s", err.(ErrController).Op)
	}
	if len(ps) != 1 {
		t.Fatalf("Get failed to return list of joined structs, want %v, got %v", 1, len(ps))
	}
	if ps[0].(*Product_WithDetails).Name != "Product Name" || ps[0].(*Product_WithDetails).Price != 1234 || ps[0].(*Product_WithDetails).ProductKindID != 33 || ps[0].(*Product_WithDetails).ProductGrpID != 113 {
		t.Fatalf("Load failed to set struct fields: %s", err.(ErrController).Op)
	}
	if ps[0].(*Product_WithDetails).ProductKind_Name != "Kind 1" || ps[0].(*Product_WithDetails).ProductGrp_Code != "GRP1" {
		t.Fatalf("Get failed to return correct list of joined objects")
	}

	ps, err = testController.Get(func() interface{} {
		return &Product_WithDetails{}
	}, GetOptions{
		Order:  []string{"ID", "asc"},
		Limit:  22,
		Offset: 0,
		Filters: map[string]interface{}{
			"Name":             "Product Name 1",
			"ProductKind_Name": "Kind 12",
			"ProductGrp_Code":  "GRP12",
			"_raw": []interface{}{
				"(.Name=? OR .ProductGrp_Code=? OR .ProductKind_Name IN (?))",
				"Product Name",
				"GRP1",
				[]string{"Kind 1", "Kind 2"},
			},
			"_rawConjuction": RawConjuctionOR,
		},
	})
	if err != nil {
		t.Fatalf("Get failed to return list of joined structs using raw filter: %s", err.(ErrController).Op)
	}
	if len(ps) != 1 {
		t.Fatalf("Get failed to return list of joined structs using raw filter, want %v, got %v", 1, len(ps))
	}
	if ps[0].(*Product_WithDetails).Name != "Product Name" || ps[0].(*Product_WithDetails).Price != 1234 || ps[0].(*Product_WithDetails).ProductKindID != 33 || ps[0].(*Product_WithDetails).ProductGrpID != 113 {
		t.Fatalf("Get failed to set fields on returned list of joined structs using raw filter")
	}
	if ps[0].(*Product_WithDetails).ProductKind_Name != "Kind 1" || ps[0].(*Product_WithDetails).ProductGrp_Code != "GRP1" {
		t.Fatalf("Get failed to set fields on returned list of joined structs using raw filter")
	}
}

func createTestJoinedStructs() {
	recreateTestJoinedStructTables()

	pg := &ProductGroup{
		ID:          113,
		Name:        "Group 1",
		Description: "A group of products",
		Code:        "GRP1",
	}
	testController.Save(pg, SaveOptions{})

	pk := &ProductKind{
		ID:   33,
		Name: "Kind 1",
	}
	testController.Save(pk, SaveOptions{})

	p := &Product{
		ID:            6,
		Name:          "Product Name",
		Price:         1234,
		ProductKindID: 33,
		ProductGrpID:  113,
	}
	testController.Save(p, SaveOptions{})
}

func recreateTestJoinedStructTables() {
	testController.DropTable(&Product{})
	testController.DropTable(&ProductGroup{})
	testController.DropTable(&ProductKind{})
	testController.CreateTable(&Product{})
	testController.CreateTable(&ProductGroup{})
	testController.CreateTable(&ProductKind{})
}
