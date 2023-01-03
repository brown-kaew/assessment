package expense

import "database/sql"

type Service interface {
	CreateNewExpense(expense Expense)
}

type service struct {
	db *sql.DB
}

func (s *service) CreateNewExpense(expense Expense) {

}

func NewService(db *sql.DB) Service {
	return &service{
		db: db,
	}
}
