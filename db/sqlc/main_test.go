package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testDB *sql.DB

// main entry point to the tests
func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open(dbDriver, dbSource)

	if err != nil {
		log.Fatal("Cannot connnect to database", err)
	}

	testQueries = New(testDB)

	// start running the unit tests
	os.Exit(m.Run())
}
