package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pobyzaarif/belajarGo2/util/database"
	cfg "github.com/pobyzaarif/go-config"
)

// CPItem is a sample model for the demo.
type CPItem struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255;not null"`
}

var loggerOption = slog.HandlerOptions{}
var logger = slog.New(slog.NewJSONHandler(os.Stdout, &loggerOption))

type Config struct {
	DBDriver string `env:"DB_DRIVER"`

	DBPostgreSQLHost     string `env:"DB_POSTGRESQL_HOST"`
	DBPostgreSQLPort     string `env:"DB_POSTGRESQL_PORT"`
	DBPostgreSQLUser     string `env:"DB_POSTGRESQL_USER"`
	DBPostgreSQLPassword string `env:"DB_POSTGRESQL_PASSWORD"`
	DBPostgreSQLName     string `env:"DB_POSTGRESQL_NAME"`
}

func main() {
	spew.Dump() // Debuger

	// Init config
	config := Config{}
	err := cfg.LoadConfig(&config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	logger.Info("Config loaded")

	// Init db connection
	databaseConfig := database.Config{
		DBDriver:             config.DBDriver,
		DBPostgreSQLHost:     config.DBPostgreSQLHost,
		DBPostgreSQLPort:     config.DBPostgreSQLPort,
		DBPostgreSQLUser:     config.DBPostgreSQLUser,
		DBPostgreSQLPassword: config.DBPostgreSQLPassword,
		DBPostgreSQLName:     config.DBPostgreSQLName,
	}
	db := databaseConfig.GetDatabaseConnection()
	logger.Info("Database client connected!")

	// Configure the connection pool and to acquire/release connections.
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB from gorm: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logger.Error("failed to close DB", "error", err)
		} else {
			logger.Info("Database connection closed")
		}
	}()

	// Configure pool settings
	sqlDB.SetMaxOpenConns(14)                  // maximum open connections
	sqlDB.SetMaxIdleConns(2)                   // maximum idle connections
	sqlDB.SetConnMaxLifetime(15 * time.Minute) // connection max lifetime

	logger.Info("delay 5 seconds, this is to observe initial connections")
	time.Sleep(5 * time.Second)

	if err := db.AutoMigrate(&CPItem{}); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	db.FirstOrCreate(&CPItem{}, CPItem{Name: "alpha"})
	db.FirstOrCreate(&CPItem{}, CPItem{Name: "beta"})

	// Demonstrate Acquire and Release using sql.DB.Conn
	// We'll run more goroutines than MaxOpenConns to show pooling in action.
	var wg sync.WaitGroup
	workers := 8
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			// Acquire a dedicated connection from the pool with a timeout
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			conn, err := sqlDB.Conn(ctx) // Acquire
			if err != nil {
				logger.Info("Failed to acquire connection", "worker", i, "error", err)
				return
			}

			// Always release the connection when done
			defer func() {
				if err := conn.Close(); err != nil {
					logger.Info("Failed to release connection", "worker", i, "error", err)
				}
			}()

			// Use the acquired connection to run a simple query
			var name string
			// Postgres uses $1 placeholder
			if err := conn.QueryRowContext(ctx, "SELECT name FROM cp_items ORDER BY id LIMIT 1 OFFSET $1", i%2).Scan(&name); err != nil {
				if err == sql.ErrNoRows {
					logger.Info("No rows found", "worker", i)
					return
				}
				logger.Info("Query error", "worker", i, "error", err)
				return
			}
			logger.Info("Query result", "worker", i, "name", name)

			// Simulate some work while holding the connection
			time.Sleep(2 * time.Second)
		}(i)
	}

	wg.Wait()

	logger.Info("All workers done. Connection pool demo complete.")
}
