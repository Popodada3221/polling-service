package main

import (
	"flag"
	"fmt"
	"log"
	"polling-service/internal/config"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var direction string
	flag.StringVar(&direction, "direction", "up", "Migration direction: up or down")
	flag.Parse()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode)

	var m *migrate.Migrate

	for i := 0; i < 5; i++ {
		m, err = migrate.New("file://migrations", dbURL)
		if err == nil {
			break
		}
		log.Printf("Waiting for database to be ready... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}

	// 	m, err := migrate.New(
	// 		"file://migrations",
	// 		dbURL,
	// 	)

	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migration up completed successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migration down completed successfully")
	default:
		log.Fatalf("Invalid migration direction: %s. Use 'up' or 'down'.", direction)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Printf("Warning: Failed to get migration version: %v", err)
	} else {
		log.Printf("Current migration version: %d, dirty: %t", version, dirty)
	}

}
