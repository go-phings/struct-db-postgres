package structdbpostgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
)

// Global vars used across all the tests
var dbUser = "struct2db"
var dbPass = "struct2db"
var dbName = "struct2db"
var dbConn *sql.DB

var dockerPool *dockertest.Pool
var dockerResource *dockertest.Resource

var testController *Controller

func TestMain(m *testing.M) {
	createDocker()
	defer removeDocker()

	createController()

	code := m.Run()
	os.Exit(code)
}

func createDocker() {
	var err error
	dockerPool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	dockerResource, err = dockerPool.Run("postgres", "13", []string{"POSTGRES_PASSWORD=" + dbPass, "POSTGRES_USER=" + dbUser, "POSTGRES_DB=" + dbName})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	if err = dockerPool.Retry(func() error {
		var err error
		dbConn, err = sql.Open("postgres", fmt.Sprintf("host=localhost user=%s password=%s port=%s dbname=%s sslmode=disable", dbUser, dbPass, dockerResource.GetPort("5432/tcp"), dbName))
		if err != nil {
			return err
		}
		return dbConn.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
}

func createController() {
	testController = NewController(dbConn, "struct2db_", nil)
}

func removeDocker() {
	dockerPool.Purge(dockerResource)
}

func getTableNameCnt(n string) (int64, error) {
	var cnt int64
	err := dbConn.QueryRow("SELECT COUNT(table_name) AS c FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1", n).Scan(&cnt)
	return cnt, err
}
