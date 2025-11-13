package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/julienschmidt/httprouter"
	invCtrl "github.com/pobyzaarif/belajarGo2/app/http-server/controller/inventory"
	invRepo "github.com/pobyzaarif/belajarGo2/repository/inventory"
	invSvc "github.com/pobyzaarif/belajarGo2/service/inventory"
	"github.com/pobyzaarif/belajarGo2/util/database"
	cfg "github.com/pobyzaarif/go-config"
)

var loggerOption = slog.HandlerOptions{AddSource: true}
var logger = slog.New(slog.NewJSONHandler(os.Stdout, &loggerOption))

type Config struct {
	AppHost string `env:"APP_HOST"`
	AppPort string `env:"APP_PORT_HTTP_SERVER"`

	DBDriver string `env:"DB_DRIVER"`

	DBMySQLHost     string `env:"DB_MYSQL_HOST"`
	DBMySQLPort     string `env:"DB_MYSQL_PORT"`
	DBMySQLUser     string `env:"DB_MYSQL_USER"`
	DBMySQLPassword string `env:"DB_MYSQL_PASSWORD"`
	DBMySQLName     string `env:"DB_MYSQL_NAME"`

	DBSQLiteName string `env:"DB_SQLITE_NAME"`

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
		DBMySQLHost:          config.DBMySQLHost,
		DBMySQLPort:          config.DBMySQLPort,
		DBMySQLUser:          config.DBMySQLUser,
		DBMySQLPassword:      config.DBMySQLPassword,
		DBMySQLName:          config.DBMySQLName,
		DBSQLiteName:         config.DBSQLiteName,
		DBPostgreSQLHost:     config.DBPostgreSQLHost,
		DBPostgreSQLPort:     config.DBPostgreSQLPort,
		DBPostgreSQLUser:     config.DBPostgreSQLUser,
		DBPostgreSQLPassword: config.DBPostgreSQLPassword,
		DBPostgreSQLName:     config.DBPostgreSQLName,
	}
	db := databaseConfig.GetDatabaseConnection()
	logger.Info("Database client connected!")

	// Dependency Injection
	inventoryRepo := invRepo.NewGormRepository(db)
	inventorySvc := invSvc.NewService(inventoryRepo)
	inventoryCtrl := invCtrl.NewController(logger, inventorySvc)

	// Setup router
	router := httprouter.New()

	router.GET("/ping", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	})

	// Inventories routes
	router.GET("/inventories", inventoryCtrl.GetAll)
	router.GET("/inventories/:code", inventoryCtrl.GetByCode)
	router.POST("/inventories", inventoryCtrl.Create)
	router.PUT("/inventories/:code", inventoryCtrl.Update)
	router.DELETE("/inventories/:code", inventoryCtrl.Delete)

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"data":{}, "message":"route not found"}`))
	})

	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"data":{}, "message":"method not allowed"}`))
	})

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		logger.Error("Panic handler", slog.Any("error", err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": http.StatusText(http.StatusInternalServerError)})
	}

	// Run server
	logger.Info("Api service running in " + config.AppHost + ":" + config.AppPort)
	server := &http.Server{
		Addr:    config.AppHost + ":" + config.AppPort,
		Handler: router,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
