package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
)

func (suite *pgTestSuite) TestDefaultClientBalance() {
	ctx := context.Background()

	testGroup := []struct {
		ClientID      int
		ExpectedLimit int64
		ExpectedError error
	}{
		{
			ClientID:      1,
			ExpectedLimit: 100000,
			ExpectedError: nil,
		},
		{
			ClientID:      2,
			ExpectedLimit: 80000,
			ExpectedError: nil,
		},
		{
			ClientID:      3,
			ExpectedLimit: 1000000,
			ExpectedError: nil,
		},
		{
			ClientID:      4,
			ExpectedLimit: 10000000,
			ExpectedError: nil,
		},
		{
			ClientID:      5,
			ExpectedLimit: 500000,
			ExpectedError: nil,
		},
		{
			ClientID:      6,
			ExpectedLimit: 100000,
			ExpectedError: ErrNotFound,
		},
	}

	for _, test := range testGroup {
		suite.Run(fmt.Sprintf("test the default balance of the client: %d", test.ClientID), func() {
			balance, err := suite.repository.getClientBalance(ctx, test.ClientID)
			suite.Assert().Equal(test.ExpectedError, err)

			if err == nil {
				suite.Assert().NotNil(balance, "balance should have the previous transaction")
				suite.Assert().Equal(test.ExpectedLimit, balance.Limit)
			}
		})
	}
}

func (suite *pgTestSuite) TestAddNewTransaction() {
	ctx := context.Background()

	testGroup := []struct {
		ClientID      int
		Transaction   internal.Transaction
		ExpectedError error
	}{
		{
			ClientID: 1,
			Transaction: internal.Transaction{
				Value:       1,
				Type:        internal.TransactionDebit,
				Description: "any description here",
				CreatedAt:   time.Now(),
			},
			ExpectedError: nil,
		},
		{
			ClientID: 1,
			Transaction: internal.Transaction{
				Value:       100000,
				Type:        internal.TransactionDebit,
				Description: "random",
				CreatedAt:   time.Now(),
			},
			ExpectedError: ErrInsufficientLimit,
		},
		{
			ClientID: 1,
			Transaction: internal.Transaction{
				Value:       99999,
				Type:        internal.TransactionDebit,
				Description: "random2 - success",
				CreatedAt:   time.Now(),
			},
			ExpectedError: nil,
		},
		{
			ClientID: 1,
			Transaction: internal.Transaction{
				Value:       1,
				Type:        internal.TransactionDebit,
				Description: "error",
				CreatedAt:   time.Now(),
			},
			ExpectedError: ErrInsufficientLimit,
		},
	}

	for _, test := range testGroup {
		suite.Run(fmt.Sprintf("test the default balance of the client: %d", test.ClientID), func() {
			resume, err := suite.repository.AddNewTransaction(ctx, test.ClientID, test.Transaction)
			suite.Assert().Equal(test.ExpectedError, err)
			if err == nil {
				balance, err := suite.repository.getClientBalance(ctx, test.ClientID)
				suite.Assert().NoError(err)
				suite.Assert().NotEmpty(balance.LastTransactions)
				suite.Assert().Equal(resume.Limit, balance.Limit)
				suite.Assert().Equal(resume.Amount, balance.Amount)
			}
		})
	}
}
