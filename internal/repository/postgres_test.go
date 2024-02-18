package repository

import (
	"context"
	"time"

	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
)

func (suite *pgTestSuite) TestCreateTransaction() {
	require := suite.Require()
	clientID := 1
	ctx := context.Background()
	input := internal.Transaction{
		Value:       1000,
		Type:        internal.TransactionCredit,
		Description: "any random description here",
		CreatedAt:   time.Now().UTC(),
	}

	suite.Run("Creates a new transaction with success", func() {
		err := suite.repository.CreateTransaction(ctx, clientID, input)
		require.NoError(err)

		statements, err := suite.repository.GetBankStatements(ctx, clientID)
		require.NoError(err)
		require.NotNil(statements, "statements should have the previous transaction")

		require.NotEmpty(statements.LastTransactions)
	})
}
