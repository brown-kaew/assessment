package expense

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

type Handler interface {
	CreateNewExpense(expense *Expense) error
	GetExpenseById(id int) (*Expense, error)
	UpdateExpenseById(expense *Expense) error
	GetAllExpenses() ([]Expense, error)
}

type handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB, g *echo.Group) Handler {
	handler := &handler{
		db: db,
	}
	handler.initRoutes(g)
	return handler
}

func (h *handler) initRoutes(g *echo.Group) {
	g.POST("/expenses", h.createNewExpenseHandler())
	g.GET("/expenses/:id", h.getExpenseHandler())
	g.PUT("/expenses/:id", h.updateExpenseHandler())
	g.GET("/expenses", h.getAllExpenseHandler())
}

func (h *handler) createNewExpenseHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var expense Expense
		err := c.Bind(&expense)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
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
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid Id")
		}
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
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		expense.Id, err = strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid Id")
		}

		err = h.UpdateExpenseById(&expense)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, expense)
	}
}

func (h *handler) getAllExpenseHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		expense, err := h.GetAllExpenses()
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

func (h *handler) GetExpenseById(id int) (*Expense, error) {
	stmt, err := h.db.Prepare("SELECT * FROM expenses WHERE id=$1")
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Cannot prepare statment")
	}

	row := stmt.QueryRow(id)

	var expense Expense
	err = row.Scan(&expense.Id, &expense.Title, &expense.Amount, &expense.Note, pq.Array(&expense.Tags))
	if err == sql.ErrNoRows {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Expense not found")
	}
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

func (h *handler) GetAllExpenses() ([]Expense, error) {
	stmt, err := h.db.Prepare("SELECT * FROM expenses")
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Cannot prepare statment")
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Can't query all expenses: "+err.Error())
	}

	var expenses []Expense
	for rows.Next() {
		var expense Expense
		err = rows.Scan(&expense.Id, &expense.Title, &expense.Amount, &expense.Note, pq.Array(&expense.Tags))
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "Can't scan expense: "+err.Error())
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}
