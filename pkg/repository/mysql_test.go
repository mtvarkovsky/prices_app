package repository

import (
	"context"
	"fmt"
	"prices/pkg/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newTestMysqlPrices(t *testing.T) (*MySQLPrices, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	repo := &MySQLPrices{
		db: db,
	}
	return repo, mock
}

func TestMysqlPrices_CreateMany(t *testing.T) {
	repo, mock := newTestMysqlPrices(t)
	now := time.Now()
	testData := []*models.Price{
		{
			ID:             "test_id_1",
			Price:          decimal.NewFromFloat(3.14),
			ExpirationDate: now.AddDate(0, 0, 1),
		},
		{
			ID:             "test_id_2",
			Price:          decimal.NewFromFloat(2.71828),
			ExpirationDate: now.AddDate(0, 0, 2),
		},
	}
	expectedQuery := `
			INSERT INTO prices (id, price, expiration_date) VALUES
			(?,?,?),(?,?,?)
			ON DUPLICATE KEY UPDATE
				id = id
		`

	mock.ExpectExec(expectedQuery).WithArgs(
		testData[0].ID, testData[0].Price, testData[0].ExpirationDate,
		testData[1].ID, testData[1].Price, testData[1].ExpirationDate,
	).WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.CreateMany(context.Background(), testData)
	assert.NoError(t, err)
}

func TestMysqlPrices_ImportFile(t *testing.T) {
	repo, mock := newTestMysqlPrices(t)
	testPath := "test/test.csv"
	expectedQuery := fmt.Sprintf(`
		LOAD DATA CONCURRENT LOCAL INFILE '%s'
		IGNORE
		INTO TABLE prices
		FIELDS TERMINATED BY ','
		LINES TERMINATED BY '\n'
		(id,price,expiration_date)
	`, testPath)

	mock.ExpectExec(expectedQuery).WithArgs().WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.ImportFile(context.Background(), testPath)
	assert.NoError(t, err)
}

func TestMysqlPrices_Get(t *testing.T) {
	repo, mock := newTestMysqlPrices(t)
	expectedPrice := &models.Price{
		ID:             "test_id_1",
		Price:          decimal.NewFromFloat(3.14),
		ExpirationDate: time.Now(),
	}

	expectedQuery := `
			SELECT id, price, expiration_date FROM prices
			WHERE id = ?
		`

	mock.ExpectQuery(expectedQuery).
		WithArgs(expectedPrice.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "price", "expiration_date"}).
				AddRow(expectedPrice.ID, expectedPrice.Price, expectedPrice.ExpirationDate),
		)

	res, err := repo.Get(context.Background(), expectedPrice.ID)
	assert.NoError(t, err)
	assert.Equal(t, expectedPrice, res)
}
