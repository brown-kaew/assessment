package expense

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type handler struct {
	Service Service
}

func (h *handler) InitRoutes(e *echo.Echo) {
	e.POST("/expenses", createNewExpense)
}

func NewHandler(s Service) *handler {
	return &handler{
		Service: s,
	}
}

func createNewExpense(c echo.Context) error {
	var expense Expense
	err := c.Bind(&expense)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, expense)
}
