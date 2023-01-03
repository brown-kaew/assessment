package expense

import (
	"database/sql"
	"log"
)

type Service interface {
	CreateNewExpense(expense Expense)
}

type service struct {
	db *sql.DB
}

func (s *service) CreateNewExpense(expense Expense) {
	log.Printf("CreateNewExpense %v", expense)
}

func NewService(db *sql.DB) Service {
	return &service{
		db: db,
	}
}
