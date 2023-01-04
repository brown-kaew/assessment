package expense_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/brown-kaew/assessment/expense"
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

	// Arrange
	expectId := "1"
	mock.ExpectQuery("INSERT INTO expenses").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectId))
	e := &expense.Expense{
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}
	handler := expense.NewHandler(db)

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

	// Arrange
	expectId := "1"
	mock.ExpectPrepare("SELECT \\* FROM expenses.*").ExpectQuery().WithArgs(expectId).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
			AddRow(expectId, "strawberry smoothie", "79", "night market promotion discount 10 bath", `{"food","beverage"}`))
	handler := expense.NewHandler(db)

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
