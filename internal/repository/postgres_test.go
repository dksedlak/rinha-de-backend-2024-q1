package repository

import (
	"context"
	"fmt"
)

func (suite *pgTestSuite) TestDefaultClientBalance() {
	require := suite.Require()
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
			require.Equal(test.ExpectedError, err)

			if err == nil {
				require.NotNil(balance, "balance should have the previous transaction")
				require.Equal(test.ExpectedLimit, balance.Limit)
			}
		})
	}
}
