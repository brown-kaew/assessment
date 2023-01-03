package expense

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type handler struct {
	service Service
}

func NewHandler(s Service) *handler {
	return &handler{
		service: s,
	}
}

func (h *handler) InitRoutes(e *echo.Echo) {
	e.POST("/expenses", h.createNewExpense())
}

func (h *handler) createNewExpense() echo.HandlerFunc {
	return func(c echo.Context) error {
		var expense Expense
		err := c.Bind(&expense)
		if err != nil {
			return err
		}
		h.service.CreateNewExpense(expense)
		return c.JSON(http.StatusCreated, expense)
	}
}
