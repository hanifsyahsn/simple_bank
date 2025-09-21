package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"testing"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://postgres:12345@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries

// This test will be triggered before all the test inside the db package due to its function name (TestMain)
func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to db:", err)
	}

	testQueries = New(conn)

	os.Exit(m.Run())
}
