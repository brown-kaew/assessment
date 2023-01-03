package expense

import (
	"database/sql"

	"github.com/lib/pq"
)

type Service interface {
	CreateNewExpense(expense *Expense) (*Expense, error)
}

type service struct {
	db *sql.DB
}

func (s *service) CreateNewExpense(expense *Expense) (*Expense, error) {
	createSql := `
	INSERT INTO
		expenses (title, amount, note, tags)
	VALUES
		($1, $2, $3, $4) 
	RETURNING id;
	`
	row := s.db.QueryRow(createSql, expense.Title, expense.Amount, expense.Note, pq.Array(&expense.Tags))

	if err := row.Scan(&expense.Id); err != nil {
		return expense, err
	}
	return expense, nil
}

func NewService(db *sql.DB) Service {
	return &service{
		db: db,
	}
}
