//go:build unit

package expense

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setUp(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return db, mock, func() { db.Close() }
}

func TestCreateNewExpense(t *testing.T) {
	db, mock, teardown := setUp(t)
	defer teardown()
	handler := NewHandler(db, echo.New().Group(""))

	// Arrange
	expectId := 1
	mock.ExpectQuery("INSERT INTO expenses").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectId))
	e := &Expense{
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}

	// Act
	err := handler.CreateNewExpense(e)

	// Assert
	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
	assert.Equal(t, expectId, e.Id)
	assert.Equal(t, "strawberry smoothie", e.Title)
	assert.Equal(t, float64(79), e.Amount)
	assert.Equal(t, "night market promotion discount 10 bath", e.Note)
	assert.Equal(t, []string{"food", "beverage"}, e.Tags)
}

func TestGetExpenseById(t *testing.T) {
	db, mock, teardown := setUp(t)
	defer teardown()
	handler := NewHandler(db, echo.New().Group(""))

	// Arrange
	expectId := 1
	mock.ExpectPrepare("SELECT \\* FROM expenses.*").ExpectQuery().WithArgs(expectId).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
			AddRow(expectId, "strawberry smoothie", "79", "night market promotion discount 10 bath", `{"food","beverage"}`))

	// Act
	e, err := handler.GetExpenseById(expectId)

	// Assert
	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
	assert.Equal(t, expectId, e.Id)
	assert.Equal(t, "strawberry smoothie", e.Title)
	assert.Equal(t, float64(79), e.Amount)
	assert.Equal(t, "night market promotion discount 10 bath", e.Note)
	assert.Equal(t, []string{"food", "beverage"}, e.Tags)
}

func TestUpdateExpenseById(t *testing.T) {
	db, mock, teardown := setUp(t)
	defer teardown()
	handler := NewHandler(db, echo.New().Group(""))

	// Arrange
	expectId := 1
	e := &Expense{
		Id:     expectId,
		Title:  "strawberry smoothie",
		Amount: 88,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}
	mock.ExpectPrepare("UPDATE expenses.*").ExpectExec().
		WithArgs(expectId, e.Title, e.Amount, e.Note, pq.Array(&e.Tags)).
		WillReturnResult(driver.RowsAffected(1))

	// Act
	err := handler.UpdateExpenseById(e)

	// Assert
	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
	assert.Equal(t, expectId, e.Id)
	assert.Equal(t, "strawberry smoothie", e.Title)
	assert.Equal(t, float64(88), e.Amount)
	assert.Equal(t, "night market promotion discount 10 bath", e.Note)
	assert.Equal(t, []string{"food", "beverage"}, e.Tags)
}

func TestGetAllExpenses(t *testing.T) {
	db, mock, teardown := setUp(t)
	defer teardown()
	handler := NewHandler(db, echo.New().Group(""))

	// Arrange
	mock.ExpectPrepare("SELECT \\* FROM expenses.*").ExpectQuery().
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
			AddRow("1", "strawberry smoothie", "79", "night market promotion discount 10 bath", `{"food","beverage"}`).
			AddRow("2", "MaMa", "5", "No money", `{"food"}`))

	// Act
	expenses, err := handler.GetAllExpenses()

	// Assert
	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(expenses))
	e := expenses[0]
	assert.Equal(t, 1, e.Id)
	assert.Equal(t, "strawberry smoothie", e.Title)
	assert.Equal(t, float64(79), e.Amount)
	assert.Equal(t, "night market promotion discount 10 bath", e.Note)
	assert.Equal(t, []string{"food", "beverage"}, e.Tags)
	e = expenses[1]
	assert.Equal(t, 2, e.Id)
	assert.Equal(t, "MaMa", e.Title)
	assert.Equal(t, float64(5), e.Amount)
	assert.Equal(t, "No money", e.Note)
	assert.Equal(t, []string{"food"}, e.Tags)
}
