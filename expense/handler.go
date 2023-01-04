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
	GetExpenseById(id string) (*Expense, error)
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
	e.GET("/expenses/:id", h.getExpenseHandler())
}

func (h *handler) createNewExpenseHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var expense Expense
		err := c.Bind(&expense)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		err = h.CreateNewExpense(&expense)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, expense)
	}
}

func (h *handler) getExpenseHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		expense, err := h.GetExpenseById(id)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, expense)
	}
}

func (h *handler) CreateNewExpense(expense *Expense) error {
	sql := `
	INSERT INTO
		expenses (title, amount, note, tags)
	VALUES
		($1, $2, $3, $4) 
	RETURNING id;
	`
	row := h.db.QueryRow(sql, expense.Title, expense.Amount, expense.Note, pq.Array(&expense.Tags))

	if err := row.Scan(&expense.Id); err != nil {
		return err
	}
	return nil
}

func (h *handler) GetExpenseById(id string) (*Expense, error) {
	stmt, err := h.db.Prepare("SELECT * FROM expenses WHERE id=$1")
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Cannot prepare statment")
	}

	row := stmt.QueryRow(id)
	if row.Err() != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Expense not found")
	}

	var e Expense
	err = row.Scan(&e.Id, &e.Title, &e.Amount, &e.Note, pq.Array(&e.Tags))
	if err != nil {
		return nil, err
	}
	return &e, nil
}
