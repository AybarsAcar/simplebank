package main

import (
	"database/sql"
	"github.com/aybarsacar/simplebank/api"
	db "github.com/aybarsacar/simplebank/db/sqlc"
	"github.com/aybarsacar/simplebank/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot load config", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)

	if err != nil {
		log.Fatal("Cannot connect to database", err)
	}

	runDatabaseMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot start the server", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Cannot start the server", err)
	}
}

func runDatabaseMigration(migrationURL string, dbSource string) {
	m, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal("cannot create new migrate instance", err)
		return
	}

	if err := m.Up(); err != nil {
		log.Fatal("failed to run migrate up", err)
	}

	log.Println("db migration successful")
}
