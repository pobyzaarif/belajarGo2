package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/pobyzaarif/belajarGo2/util/database"
	cfg "github.com/pobyzaarif/go-config"
)

var loggerOption = slog.HandlerOptions{AddSource: true}
var logger = slog.New(slog.NewJSONHandler(os.Stdout, &loggerOption))

type Config struct {
	AppHost string `env:"APP_HOST"`
	AppPort string `env:"APP_PORT"`

	DBDriver        string `env:"DB_DRIVER"`
	DBMySQLHost     string `env:"DB_MYSQL_HOST"`
	DBMySQLPort     string `env:"DB_MYSQL_PORT"`
	DBMySQLUser     string `env:"DB_MYSQL_USER"`
	DBMySQLPassword string `env:"DB_MYSQL_PASSWORD"`
	DBMySQLName     string `env:"DB_MYSQL_NAME"`
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
		DBDriver:        config.DBDriver,
		DBMySQLHost:     config.DBMySQLHost,
		DBMySQLPort:     config.DBMySQLPort,
		DBMySQLUser:     config.DBMySQLUser,
		DBMySQLPassword: config.DBMySQLPassword,
		DBMySQLName:     config.DBMySQLName,
	}
	_ = databaseConfig.GetDatabaseConnection()
	logger.Info("Database client connected!")

	// Setup router
	router := httprouter.New()
	router.GET("/ping", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	})

	router.GET("/querypath/:id", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		id := p.ByName("id")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": id})
	})

	router.GET("/queryparam", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		queryParam := r.URL.Query()
		a := map[string][]string{}
		for k, v := range queryParam {
			a[k] = v
		}
		spew.Dump(a["id"])
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
	})

	router.POST("/urlencoded", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(map[string]string{"message": "b"})
	})

	router.POST("/body", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		decoder := json.NewDecoder(r.Body)
		type user struct {
			Name string `json:"name" validate:"required"`
			Age  int    `json:"age"`
		}

		u := user{}
		err := decoder.Decode(&u)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid request body"})
			return
		}

		err = validator.New().Struct(u)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid request body"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(map[string]string{"message": u.Name})
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
