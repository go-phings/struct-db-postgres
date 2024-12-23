package structdbpostgres

import (
	"database/sql"

	stsql "github.com/go-phings/struct-sql-postgres"
)

// Controller is the main component that gets and saves objects in the database.
type Controller struct {
	dbConn        *sql.DB
	dbTblPrefix   string
	sqlGenerators map[string]*stsql.StructSQL
	tagName       string
}

type ControllerConfig struct {
	TagName string
}

// NewController returns new Controller object
func NewController(dbConn *sql.DB, tblPrefix string, cfg *ControllerConfig) *Controller {
	c := &Controller{
		dbConn:      dbConn,
		dbTblPrefix: tblPrefix,
	}

	if cfg != nil && cfg.TagName != "" {
		c.tagName = cfg.TagName
	}

	if c.tagName == "" {
		c.tagName = "2db"
	}

	c.sqlGenerators = make(map[string]*stsql.StructSQL)
	return c
}
