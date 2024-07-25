package routes

import (
	"account/handlers"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Register routes in client instances to echo
func RegisterRoutes(e *echo.Echo, c *handlers.Client) {
	apiV1 := e.Group("/api/v1")

	apiV1.POST("/daftar", c.Register)
	apiV1.POST("/tabung", c.Deposit)
	apiV1.POST("/tarik", c.Withdraw)
	apiV1.GET("/saldo/:no_rekening", c.CheckBalance)

	e.GET("/api/version", func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"version": "1.0.0"})
	})
}
