package repository

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

const postgresDSN = "postgres://rinha_backend:nopassword@localhost:5432/production?sslmode=disable"

type pgTestSuite struct {
	suite.Suite
	pg         *sqlx.DB
	repository *PostgreSQL
}

func (suite *pgTestSuite) SetupSuite() {
	require := suite.Require()

	pg, err := NewPostgreSQL(context.Background(),
		PgConfig{
			DSN: postgresDSN,
		},
	)
	require.NoError(err)
	require.NotNil(pg)

	suite.pg = pg.GetDBInstance()
	suite.repository = &PostgreSQL{db: suite.pg}
}

func (suite *pgTestSuite) TearDownSuite() {
	suite.pg.Close()
}

func TestNewPostgres(t *testing.T) {
	if os.Getenv("ENABLE_INTEGRATION_TEST") != "1" {
		t.Skip("[integration tests] skipping ...")
		return
	}

	t.Parallel()
	suite.Run(t, &pgTestSuite{})
}
