package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projects/go/01_simple_bank/util"
)

var testQueries *Queries
var testDB DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../../")

	if err != nil {
		log.Fatal("cannot load configuration:", err)
	}

	testDB, err = pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())

}
