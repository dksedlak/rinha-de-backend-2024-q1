package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

const postgresDSN = "postgres://rinha_backend:nopassword@localhost:5432/production?sslmode=disable"

type pgTestSuite struct {
	suite.Suite

	pgMain       *sqlx.DB
	pgTest       *sqlx.DB
	testDatabase string

	repository internal.Repository
}

func (suite *pgTestSuite) SetupSuite() {
	require := suite.Require()

	pgMain, err := NewPostgreSQL(context.Background(),
		PgConfig{
			DSN: postgresDSN,
		},
	)
	require.NoError(err)
	require.NotNil(pgMain)

	suite.pgMain = pgMain.GetDBInstance()
}

func (suite *pgTestSuite) TearDownSuite() {
	suite.pgMain.Close()
}

func (suite *pgTestSuite) createTestDatabase() *sqlx.DB {
	require := suite.Require()

	hash := sha256.Sum256([]byte(suite.T().Name()))
	suite.testDatabase = "test_" + hex.EncodeToString(hash[:])[:32] + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)

	_, err := suite.pgMain.ExecContext(context.Background(), "CREATE DATABASE "+suite.testDatabase)
	require.NoError(err)

	dsn := postgresDSN[0:strings.LastIndex(postgresDSN, "/")+1] +
		suite.testDatabase +
		postgresDSN[strings.LastIndex(postgresDSN, "?"):]

	dbClient, err := NewPostgreSQL(context.Background(), PgConfig{DSN: dsn})
	require.NoError(err)

	return dbClient.GetDBInstance()
}

func (suite *pgTestSuite) SetupTest() {
	suite.pgTest = suite.createTestDatabase()
	suite.repository = &PostgreSQL{db: suite.pgTest}
}

func (suite *pgTestSuite) TearDownTest() {
	suite.pgTest.Close()

	_, err := suite.pgMain.Exec(
		fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", suite.testDatabase),
	)
	suite.Require().NoError(err)
}

func TestNewPostgres(t *testing.T) {
	if os.Getenv("ENABLE_INTEGRATION_TEST") != "1" {
		t.Skip("[integration tests] skipping ...")
		return
	}

	t.Parallel()
	suite.Run(t, &pgTestSuite{})
}
