package structdbpostgres

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

	stsql "github.com/go-phings/struct-sql-postgres"
)

const RawConjuctionOR = 1
const RawConjuctionAND = 2

type LoadOptions struct {
	Unused bool
}

type SaveOptions struct {
	NoInsert   bool
	ModifiedBy int64
	ModifiedAt int64
}

type GetOptions struct {
	Order               []string
	Limit               int
	Offset              int
	Filters             map[string]interface{}
	RowObjTransformFunc func(interface{}) interface{}
}

type DeleteOptions struct {
	Constructors map[string]func() interface{}
}

type DeleteMultipleOptions struct {
	Filters            map[string]interface{}
	CascadeDeleteDepth int
	Constructors       map[string]func() interface{}
}

type UpdateMultipleOptions struct {
	Filters                 map[string]interface{}
	CascadeDeleteDepth      int
	ConvertValuesFromString bool
}

type GetCountOptions struct {
	Filters map[string]interface{}
}

// Save takes object, validates its field values and saves it in the database.
// If ID is not present then an INSERT will be performed
// If ID is set then an "upsert" is performed
func (c Controller) Save(obj interface{}, options SaveOptions) *ErrController {
	h, err := c.getSQLGenerator(obj, nil, "")
	if err != nil {
		return err
	}

	b, invalidFields, err2 := c.Validate(obj, nil)
	if err2 != nil {
		return &ErrController{
			Op:  "Validate",
			Err: fmt.Errorf("Error when trying to validate: %w", err2),
		}
	}

	if !b {
		return &ErrController{
			Op: "Validate",
			Err: &ErrValidation{
				Fields: invalidFields,
			},
		}
	}

	var err3 error
	if c.GetObjIDValue(obj) != 0 {
		if options.ModifiedAt != 0 && options.ModifiedBy != 0 && h.HasModificationFields() {
			c.SetObjLastModified(obj, options.ModifiedAt, options.ModifiedBy)
		}

		// do no try to insert if NoInsert is set
		// TODO: error handling, we should check if object exists - for now nothing happens, UPDATE gets executed and updates nothing
		if options.NoInsert {
			_, err3 = c.dbConn.Exec(h.GetQueryUpdateById(), append(c.GetObjFieldInterfaces(obj, false), c.GetObjIDInterface(obj))...)
		} else {
			// try to insert - if ID already exists then try to update it
			_, err3 = c.dbConn.Exec(h.GetQueryInsertOnConflictUpdate(), append(c.GetObjFieldInterfaces(obj, true), c.GetObjFieldInterfaces(obj, false)...)...)
		}
	} else {
		if options.ModifiedAt != 0 && options.ModifiedBy != 0 && h.HasModificationFields() {
			c.SetObjCreated(obj, options.ModifiedAt, options.ModifiedBy)
		}
		err3 = c.dbConn.QueryRow(h.GetQueryInsert(), c.GetObjFieldInterfaces(obj, false)...).Scan(c.GetObjIDInterface(obj))
	}
	if err3 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err3),
		}
	}
	return nil
}

// Load sets object's fields with values from the database table with a specific id. If record does not exist
// in the database, all field values in the struct are zeroed
// TODO: Should it return an ErrNotExist?
func (c Controller) Load(obj interface{}, id string, options LoadOptions) *ErrController {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return &ErrController{
			Op:  "IDToInt",
			Err: fmt.Errorf("Error converting string to int: %w", err),
		}
	}

	h, err2 := c.getSQLGenerator(obj, nil, "")
	if err2 != nil {
		return err2
	}

	err3 := c.dbConn.QueryRow(h.GetQuerySelectById(), int64(idInt)).Scan(c.GetObjFieldInterfaces(obj, true)...)
	switch {
	case err3 == sql.ErrNoRows:
		c.ResetFields(obj)
		return nil
	case err3 != nil:
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err3),
		}
	default:
		return nil
	}
}

// Delete removes object from the database table and it does that only when ID field is set (greater than 0).
// Once deleted from the DB, all field values are zeroed
// TODO: Error handling probably needs re-designing
func (c Controller) Delete(obj interface{}, options DeleteOptions) *ErrController {
	h, err := c.getSQLGenerator(obj, nil, "")
	if err != nil {
		return err
	}

	id := c.GetObjIDValue(obj)

	if id == 0 {
		return nil
	}
	_, err2 := c.dbConn.Exec(h.GetQueryDeleteById(), c.GetObjIDInterface(obj))
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	c.ResetFields(obj)

	// Loop through fields to delete cascade
	err3 := c.runOnDelete(obj, c.tagName, []int64{id}, 0)
	if err3 != nil {
		return err3
	}

	return nil
}

// DeleteMultiple removes objects from the database based on specified filters
func (c Controller) DeleteMultiple(obj interface{}, options DeleteMultipleOptions) *ErrController {
	h, err := c.getSQLGenerator(obj, nil, "")
	if err != nil {
		return err
	}

	// TODO: Enable validation once struct-validator support reflect.Value
	if len(options.Filters) > 0 {
		b, invalidFields, err1 := c.Validate(obj, options.Filters)
		if err1 != nil {
			return &ErrController{
				Op:  "ValidateFilters",
				Err: fmt.Errorf("Error when trying to validate filters: %w", err1),
			}
		}

		if !b {
			return &ErrController{
				Op: "ValidateFilters",
				Err: &ErrValidation{
					Fields: invalidFields,
				},
			}
		}
	}

	// Run DELETE query and get IDs of deleted rows
	rows, err2 := c.dbConn.Query(h.GetQueryDeleteReturningID(options.Filters, nil), c.GetFiltersInterfaces(options.Filters)...)
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	defer rows.Close()

	returnedIds := []int64{}

	for rows.Next() {
		var returnedId int64
		err3 := rows.Scan(&returnedId)
		if err3 != nil {
			return &ErrController{
				Op:  "DBQueryRowsScan",
				Err: fmt.Errorf("Error scanning DB query row: %w", err3),
			}
		}

		returnedIds = append(returnedIds, returnedId)
	}

	if options.CascadeDeleteDepth < 3 {
		// Loop through fields to delete cascade
		err3 := c.runOnDelete(obj, c.tagName, returnedIds, options.CascadeDeleteDepth)
		if err3 != nil {
			return err3
		}
	}

	return nil
}

// UpdateMultiple updates specific fields in objects from the database based on specified filters
func (c Controller) UpdateMultiple(obj interface{}, values map[string]interface{}, options UpdateMultipleOptions) *ErrController {
	h, err := c.getSQLGenerator(obj, nil, "")
	if err != nil {
		return err
	}

	if len(values) < 1 {
		return &ErrController{
			Op:  "MissingValues",
			Err: fmt.Errorf("missing values for update"),
		}
	}

	if options.ConvertValuesFromString {
		values = c.StringToFieldValues(obj, values)
	}

	b, invalidFields, err1 := c.Validate(obj, values)
	if err1 != nil {
		return &ErrController{
			Op:  "ValidateValues",
			Err: fmt.Errorf("Error when trying to validate values: %w", err1),
		}
	}

	if !b {
		return &ErrController{
			Op: "ValidateValues",
			Err: &ErrValidation{
				Fields: invalidFields,
			},
		}
	}

	if len(options.Filters) > 0 {
		b, invalidFields, err1 := c.Validate(obj, options.Filters)
		if err1 != nil {
			return &ErrController{
				Op:  "ValidateFilters",
				Err: fmt.Errorf("Error when trying to validate filters: %w", err1),
			}
		}

		if !b {
			return &ErrController{
				Op: "ValidateFilters",
				Err: &ErrValidation{
					Fields: invalidFields,
				},
			}
		}
	}

	_, err2 := c.dbConn.Exec(h.GetQueryUpdate(values, options.Filters, nil, nil), append(c.GetFiltersInterfaces(values), c.GetFiltersInterfaces(options.Filters)...)...)
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}

	return nil
}

// Get runs a select query on the database with specified filters, order, limit and offset and returns a
// list of objects
func (c Controller) Get(newObjFunc func() interface{}, options GetOptions) ([]interface{}, *ErrController) {
	obj := newObjFunc()

	h, err := c.getSQLGenerator(obj, nil, "")
	if err != nil {
		return nil, err
	}

	if len(options.Filters) > 0 {
		b, invalidFields, err1 := c.Validate(obj, options.Filters)
		if err1 != nil {
			return nil, &ErrController{
				Op:  "ValidateFilters",
				Err: fmt.Errorf("Error when trying to validate filters: %w", err1),
			}
		}

		if !b {
			return nil, &ErrController{
				Op: "ValidateFilters",
				Err: &ErrValidation{
					Fields: invalidFields,
				},
			}
		}
	}

	var v []interface{}
	rows, err2 := c.dbConn.Query(h.GetQuerySelect(options.Order, options.Limit, options.Offset, options.Filters, nil, nil), c.GetFiltersInterfaces(options.Filters)...)
	if err2 != nil {
		return nil, &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	defer rows.Close()

	for rows.Next() {
		newObj := newObjFunc()
		err3 := rows.Scan(c.GetObjFieldInterfaces(newObj, true)...)
		if err3 != nil {
			return nil, &ErrController{
				Op:  "DBQueryRowsScan",
				Err: fmt.Errorf("Error scanning DB query row: %w", err3),
			}
		}

		// If options.RowObjTransformFunc is defined then call it on the row
		if options.RowObjTransformFunc != nil {
			v = append(v, options.RowObjTransformFunc(newObj))
			continue
		}

		// Normal append
		v = append(v, newObj)
	}

	return v, nil
}

// GetCount runs a 'SELECT COUNT(*)' query on the database with specified filters, order, limit and offset and returns count of rows
func (c Controller) GetCount(newObjFunc func() interface{}, options GetCountOptions) (int64, *ErrController) {
	obj := newObjFunc()
	h, err := c.getSQLGenerator(obj, nil, "")
	if err != nil {
		return 0, err
	}

	if len(options.Filters) > 0 {
		b, invalidFields, err1 := c.Validate(obj, options.Filters)
		if err1 != nil {
			return 0, &ErrController{
				Op:  "ValidateFilters",
				Err: fmt.Errorf("Error when trying to validate filters: %w", err1),
			}
		}

		if !b {
			return 0, &ErrController{
				Op: "ValidateFilters",
				Err: &ErrValidation{
					Fields: invalidFields,
				},
			}
		}
	}

	row := c.dbConn.QueryRow(h.GetQuerySelectCount(options.Filters, nil), c.GetFiltersInterfaces(options.Filters)...)
	var cnt int64
	err3 := row.Scan(&cnt)
	if err3 != nil {
		return 0, &ErrController{
			Op:  "DBQueryRowScan",
			Err: fmt.Errorf("Error scanning DB query row: %w", err3),
		}
	}

	return cnt, nil
}

// AddSQLGenerator adds StructSQL object to sqlGenerators
func (c *Controller) AddSQLGenerator(obj interface{}, parentObj interface{}, overwrite bool, forceName string, parentOnlyRoot bool) *ErrController {
	n := c.getSQLGeneratorName(obj, false)

	// If sql generator already exists and it should not be overwritten then finish
	if !overwrite {
		_, ok := c.sqlGenerators[n]
		if ok {
			return nil
		}
	}

	var sourceHelper *stsql.StructSQL
	if parentObj != nil {
		h, err := c.getSQLGenerator(parentObj, nil, "")
		if err != nil {
			return &ErrController{
				Op:  "GetHelper",
				Err: fmt.Errorf("Error getting StructSQL: %w", h.Err()),
			}
		}
		sourceHelper = h
		forceName = c.getSQLGeneratorName(parentObj, parentOnlyRoot)
	}

	h := stsql.NewStructSQL(obj, stsql.StructSQLOptions{
		DatabaseTablePrefix: c.dbTblPrefix,
		ForceName:           forceName,
		Base:                sourceHelper,
		TagName:             c.tagName,
	})
	if h.Err() != nil {
		return &ErrController{
			Op:  "GetHelper",
			Err: fmt.Errorf("Error getting StructSQL: %w", h.Err()),
		}
	}
	c.sqlGenerators[n] = h
	return nil
}

// GetDBCol returns column name used in the database
func (c *Controller) GetFieldNameFromDBCol(obj interface{}, dbCol string) (string, *ErrController) {
	h, err := c.getSQLGenerator(obj, nil, "")
	if err != nil {
		return "", err
	}
	fieldName := h.GetFieldNameFromDBCol(dbCol)
	return fieldName, nil
}

// StringToFieldValues converts map of field values which are in string to values of the same kind of fields are
func (c *Controller) StringToFieldValues(obj interface{}, values map[string]interface{}) map[string]interface{} {
	o := map[string]interface{}{}

	v := reflect.ValueOf(obj)
	i := reflect.Indirect(v)
	s := i.Type()

	if s.String() == "reflect.Value" {
		s = reflect.ValueOf(obj.(reflect.Value).Interface()).Type().Elem().Elem()
	}

	for k, v := range values {
		field, ok := s.FieldByName(k)
		if !ok {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Int64, reflect.Uint64:
			i, err := strconv.ParseInt(v.(string), 10, 64)
			if err == nil {
				o[k] = i
			}
		case reflect.Int32, reflect.Uint32:
			i, err := strconv.ParseInt(v.(string), 10, 32)
			if err == nil {
				o[k] = i
			}
		case reflect.Int16, reflect.Uint16:
			i, err := strconv.ParseInt(v.(string), 10, 16)
			if err == nil {
				o[k] = i
			}
		case reflect.Int8, reflect.Uint8:
			i, err := strconv.ParseInt(v.(string), 10, 8)
			if err == nil {
				o[k] = i
			}
		case reflect.Int, reflect.Uint:
			i, err := strconv.Atoi(v.(string))
			if err == nil {
				o[k] = i
			}
		case reflect.Float32:
			i, err := strconv.ParseFloat(v.(string), 32)
			if err == nil {
				o[k] = i
			}
		case reflect.Float64:
			i, err := strconv.ParseFloat(v.(string), 64)
			if err == nil {
				o[k] = i
			}
		case reflect.String:
			o[k] = v
		case reflect.Bool:
			if v.(string) == "true" {
				o[k] = true
			} else {
				o[k] = false
			}
		default:
			continue
		}
	}

	return o
}
