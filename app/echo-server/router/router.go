package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pobyzaarif/belajarGo2/app/echo-server/controller/inventory"
	"github.com/pobyzaarif/belajarGo2/app/echo-server/controller/user"
	"github.com/pobyzaarif/belajarGo2/app/echo-server/middleware"
)

func RegisterPath(
	e *echo.Echo,
	jwtSecret string,
	ctrlInv *inventory.Controller,
	ctrlUser *user.Controller,
) {
	// Setup routes
	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "pong",
		})
	})

	// Init JWT
	jwtMiddleware := middleware.JWTMiddleware(jwtSecret)

	// Init ACL
	userNAdminAccess := middleware.ACLMiddleware(map[string]bool{
		"admin": true,
		"user":  true,
	})
	adminAccess := middleware.ACLMiddleware(map[string]bool{
		"admin": true,
	})
	superadminAccess := middleware.ACLMiddleware(map[string]bool{
		"superadmin": true,
	})

	// User endpoint
	userEndpoint := e.Group("/users")
	userEndpoint.POST("/register", ctrlUser.Register)
	userEndpoint.POST("/login", ctrlUser.Login)
	userEndpoint.GET("/email-verification/:code", ctrlUser.VerifyEmail)

	// Inventory endpoint
	inventoryEndpoint := e.Group("/inventories", jwtMiddleware)
	inventoryEndpoint.GET("", ctrlInv.GetAll, userNAdminAccess)
	inventoryEndpoint.GET("/:code", ctrlInv.GetByCode, userNAdminAccess)
	inventoryEndpoint.POST("", ctrlInv.Create, adminAccess)
	inventoryEndpoint.PUT("/:code", ctrlInv.Update, adminAccess)
	inventoryEndpoint.DELETE("/:code", ctrlInv.Delete, superadminAccess)

	// Explore endpoint
	echoJWT := middleware.JwtEchoMiddleware(jwtSecret) // poc: rafly
	exploreEndpoint := e.Group("/explore", echoJWT)
	exploreEndpoint.GET("/rafly", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"message": "halim"}) // poc: halim
	})
}
