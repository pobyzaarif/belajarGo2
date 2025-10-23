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

	jwtMiddleware := middleware.JWTMiddleware(jwtSecret)
	userNAdmin := middleware.RBACMiddleware([]string{"user", "admin"})
	adminOnly := middleware.RBACMiddleware([]string{"admin"})
	superadminOnly := middleware.RBACMiddleware([]string{"superadmin"})

	// user endpoint
	userEndpoint := e.Group("/users")
	userEndpoint.POST("/register", ctrlUser.Register)
	userEndpoint.POST("/login", ctrlUser.Login)

	// inventory endpoint
	inventoryEndpoint := e.Group("/inventories", jwtMiddleware)
	inventoryEndpoint.GET("", ctrlInv.GetAll, userNAdmin)
	inventoryEndpoint.GET("/:code", ctrlInv.GetByCode, userNAdmin)
	inventoryEndpoint.POST("", ctrlInv.Create, adminOnly)
	inventoryEndpoint.PUT("/:code", ctrlInv.Update, adminOnly)
	inventoryEndpoint.DELETE("/:code", ctrlInv.Delete, superadminOnly)
}
