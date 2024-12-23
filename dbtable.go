package structdbpostgres

import (
	"fmt"
)

// DropTables drop tables in the database for specified objects (see DropTable for a single struct)
func (c Controller) DropTables(xobj ...interface{}) *ErrController {
	for _, obj := range xobj {
		err := c.DropTable(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateTables creates tables in the database for specified objects (see CreateTable for a single struct)
func (c Controller) CreateTables(xobj ...interface{}) *ErrController {
	for _, obj := range xobj {
		err := c.CreateTable(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateTable creates database table to store specified type of objects. It takes struct name and its fields,
// converts them into table and columns names (all lowercase with underscore), assigns column type based on the
// field type, and then executes "CREATE TABLE" query on attached DB connection
func (c Controller) CreateTable(obj interface{}) *ErrController {
	h, err := c.getSQLGenerator(obj, nil, "")
	if err != nil {
		return err
	}

	_, err2 := c.dbConn.Exec(h.GetQueryCreateTable())
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	return nil
}

// DropTable drops database table used to store specified type of objects. It just takes struct name, converts
// it to lowercase-with-underscore table name and executes "DROP TABLE" query using attached DB connection
func (c Controller) DropTable(obj interface{}) *ErrController {
	h, err := c.getSQLGenerator(obj, nil, "")
	if err != nil {
		return err
	}

	_, err2 := c.dbConn.Exec(h.GetQueryDropTable())
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	return nil
}
