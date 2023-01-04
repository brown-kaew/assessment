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
	UpdateExpenseById(expense *Expense) error
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
	e.PUT("/expenses/:id", h.updateExpenseHandler())
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

func (h *handler) updateExpenseHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var expense Expense
		err := c.Bind(&expense)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		expense.Id = c.Param("id")

		err = h.UpdateExpenseById(&expense)
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

	var expense Expense
	err = row.Scan(&expense.Id, &expense.Title, &expense.Amount, &expense.Note, pq.Array(&expense.Tags))
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (h *handler) UpdateExpenseById(expense *Expense) error {
	stmt, err := h.db.Prepare(`
	UPDATE expenses
	SET
		title=$2,
		amount=$3,
		note=$4,
		tags=$5
	WHERE id=$1
	`)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cannot prepare statment")
	}

	var res sql.Result
	res, err = stmt.Exec(expense.Id, expense.Title, expense.Amount, expense.Note, pq.Array(&expense.Tags))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cannot update expense: "+err.Error())
	}

	row, err := res.RowsAffected()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cannot update expense: "+err.Error())
	}
	if row == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "Expense not found")
	}
	return nil
}
