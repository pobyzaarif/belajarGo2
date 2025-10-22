package inventory

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pobyzaarif/belajarGo2/service/inventory"
)

type Controller struct {
	logger       *slog.Logger
	inventorySvc inventory.Service
}

func NewController(
	logger *slog.Logger,
	s inventory.Service,
) *Controller {
	return &Controller{
		logger:       logger,
		inventorySvc: s,
	}
}

type InventoryRequest struct {
	Code        string `json:"code" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Stock       int    `json:"stock"`
	Description string `json:"description"`
	Status      string `json:"status" validate:"required,oneof=active broken"`
}

func (ctrl *Controller) Create(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]interface{}{"message": "OK", "data": "data"})
}

func (ctrl *Controller) GetAll(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "OK", "data": "data"})
}

func (ctrl *Controller) GetByCode(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "OK", "data": "data"})
}

func (ctrl *Controller) Update(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "OK", "data": "data"})
}

func (ctrl *Controller) Delete(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "OK", "data": "data"})
}
