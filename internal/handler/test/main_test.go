package handler_test

import (
	"os"
	"testing"

	"github.com/akhilr007/tasks/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testpool *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		panic("TEST_DB_URL not set")
	}

	var err error
	testpool, err = db.NewPool(dsn)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	testpool.Close()
	os.Exit(code)
}