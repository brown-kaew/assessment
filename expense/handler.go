package expense

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetHandler(e *echo.Echo) {
	e.POST("/expenses", createNewExpense)
}

func createNewExpense(c echo.Context) error {
	var expense Expense
	err := c.Bind(&expense)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, expense)
}
