package internal

import "time"

type TransactionType string

const (
	TransactionDebit  TransactionType = "d"
	TransactionCredit TransactionType = "c"
)

type Transaction struct {
	Value       int64
	Type        TransactionType
	Description string
	CreatedAt   time.Time
}

type Resume struct {
	Amount int64
	Limit  int64
}

type BankStatement struct {
	Amount           int64
	Date             time.Time
	Limit            int64
	LastTransactions []Transaction
}
