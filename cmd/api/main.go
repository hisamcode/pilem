package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"pilem/internal/server"

	"github.com/joho/godotenv"
)

type config struct {
	db struct {
		dsn string
	}
}

func main() {
	var cfg config

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// database config
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")

	flag.Parse()

	// OpenDB
	db, err := openDB(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	server := server.NewServer(db)

	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
