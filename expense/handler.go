package expense

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

type Handler interface {
	InitRoutes(e *echo.Echo)
	CreateNewExpense(expense *Expense) error
}

type handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) Handler {
	return &handler{
		db: db,
	}
}

func (h *handler) InitRoutes(e *echo.Echo) {
	e.POST("/expenses", h.createNewExpenseHandler())
}

func (h *handler) createNewExpenseHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var expense Expense
		err := c.Bind(&expense)
		if err != nil {
			return err
		}
		err = h.CreateNewExpense(&expense)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, expense)
	}
}

func (h *handler) CreateNewExpense(expense *Expense) error {
	createSql := `
	INSERT INTO
		expenses (title, amount, note, tags)
	VALUES
		($1, $2, $3, $4) 
	RETURNING id;
	`
	row := h.db.QueryRow(createSql, expense.Title, expense.Amount, expense.Note, pq.Array(&expense.Tags))

	if err := row.Scan(&expense.Id); err != nil {
		return err
	}
	return nil
}
