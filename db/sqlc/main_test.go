package db

import (
	"database/sql"
	"github.com/aybarsacar/simplebank/util"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

// main entry point to the tests
func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Cannot load config", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)

	if err != nil {
		log.Fatal("Cannot connect to database", err)
	}

	testQueries = New(testDB)

	// start running the unit tests
	os.Exit(m.Run())
}
