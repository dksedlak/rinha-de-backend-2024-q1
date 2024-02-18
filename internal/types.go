package internal

import "time"

type TransactionType string

const (
	TransactionDebit  TransactionType = "d"
	TransactionCredit TransactionType = "c"
)

type Transaction struct {
	Value       uint64
	Type        TransactionType
	Description string
	CreatedAt   time.Time
}

type BankStatement struct {
	Total            uint64
	Date             time.Time
	Limit            uint64
	LastTransactions []Transaction
}
