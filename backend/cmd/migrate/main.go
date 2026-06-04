package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"icaris-logistic/backend/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down|version>")
	}

	cfg := config.Load()

	m, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("init migrate: %v", err)
	}
	defer m.Close()

	command := os.Args[1]
	switch command {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "version":
		printVersion(m)
		return
	default:
		log.Fatalf("unknown command %q (use up|down|version)", command)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate %s: %v", command, err)
	}

	printVersion(m)
}

func printVersion(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("schema version: none (no migrations applied)")
		return
	}
	if err != nil {
		log.Fatalf("read version: %v", err)
	}

	fmt.Printf("schema version: %d (dirty=%v)\n", version, dirty)
}
