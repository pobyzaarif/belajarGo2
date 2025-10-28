package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	invCtrl "github.com/pobyzaarif/belajarGo2/app/echo-server/controller/inventory"
	"github.com/pobyzaarif/belajarGo2/app/echo-server/controller/user"
	_ "github.com/pobyzaarif/belajarGo2/app/echo-server/docs"
	"github.com/pobyzaarif/belajarGo2/app/echo-server/router"
	invRepo "github.com/pobyzaarif/belajarGo2/repository/inventory"
	userRepo "github.com/pobyzaarif/belajarGo2/repository/user"
	invSvc "github.com/pobyzaarif/belajarGo2/service/inventory"
	userSvc "github.com/pobyzaarif/belajarGo2/service/user"
	"github.com/pobyzaarif/belajarGo2/util/database"
	cfg "github.com/pobyzaarif/go-config"
	echoSwagger "github.com/swaggo/echo-swagger"
)

var loggerOption = slog.HandlerOptions{AddSource: true}
var logger = slog.New(slog.NewJSONHandler(os.Stdout, &loggerOption))

type Config struct {
	AppHost      string `env:"APP_HOST"`
	AppPort      string `env:"APP_PORT"`
	AppJWTSecret string `env:"APP_JWT_SECRET"`

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

	// Setup server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.CORS())
	e.Use(middleware.LoggerWithConfig(
		middleware.LoggerConfig{
			Skipper: middleware.DefaultSkipper,
			Format: `{"time":"${time_rfc3339_nano}","level":"INFO","id":"${id}","remote_ip":"${remote_ip}",` +
				`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
				`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
				`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
			CustomTimeFormat: "2006-01-02 15:04:05.00000",
		},
	))
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(middleware.Recover())

	// Setup routes
	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "pong",
		})
	})

	e.GET("/swagger/*", echoSwagger.EchoWrapHandler())

	// user
	userRepo := userRepo.NewGormRepository(db)
	userSvc := userSvc.NewService(logger, userRepo, config.AppJWTSecret)
	userCtrl := user.NewController(logger, userSvc)

	// inventory
	inventoryRepo := invRepo.NewGormRepository(db)
	inventorySvc := invSvc.NewService(inventoryRepo)
	inventoryCtrl := invCtrl.NewController(logger, inventorySvc)

	router.RegisterPath(
		e,
		config.AppJWTSecret,
		inventoryCtrl,
		userCtrl,
	)

	// Start server
	address := config.AppHost + ":" + config.AppPort
	go func() {
		if err := e.Start(address); err != http.ErrServerClosed {
			log.Fatal("Failed on http server " + config.AppPort)
		}
	}()

	logger.Info("Api service running in " + address)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// a timeout of 10 seconds to shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Failed to shutting down echo server", "err", err)
	} else {
		logger.Info("Successfully shutting down echo server")
	}
}
