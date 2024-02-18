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

type BankStatement struct {
	Total            int64
	Date             time.Time
	Limit            int64
	LastTransactions []Transaction
}
