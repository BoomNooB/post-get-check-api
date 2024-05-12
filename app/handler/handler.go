package handler

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type HandlerInterface interface {
	HealthCheck(ec echo.Context) error
}

type handler struct {
	// service   service.Servicer
	validator *validator.Validate
}

func NewHandler() HandlerInterface {
	validator := validator.New()
	return &handler{validator}
}

func (h *handler) HealthCheck(ec echo.Context) error {
	return ec.String(http.StatusOK, "Service is healthy!")
}
