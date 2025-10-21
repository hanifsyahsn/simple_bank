package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/hanifsyahsn/simple_bank/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

// This test will be triggered before all the test inside the db package due to its function name (TestMain)
func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Cannot load config file: ", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
