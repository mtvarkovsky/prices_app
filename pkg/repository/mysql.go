package repository

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // DB driver
	"github.com/nullism/bqb"
	"prices/pkg/config"
	"prices/pkg/errors"
	"prices/pkg/models"
	"time"
)

type (
	mysqlPrices struct {
		db     *sql.DB
		config config.Storage
	}
)

func newMysqlPrices(config config.Storage) (*mysqlPrices, error) {
	db, err := sql.Open(config.Type, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("can't establish connection to the data storage: %w", err)
	}
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxConnections)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't ping data storage: %w", err)
	}

	return &mysqlPrices{
		db:     db,
		config: config,
	}, nil
}

func (r *mysqlPrices) Create(ctx context.Context, prices []*models.Price) error {
	values := bqb.Q()
	for _, price := range prices {
		values.Comma("(?,?,?)", price.ID, price.Price, price.ExpirationDate)
	}
	q := bqb.New(
		`
			INSERT INTO prices (id, price, expiration_date) VALUES ?
-- 			ON DUPLICATE KEY UPDATE 
-- 				price = IF(expiration_date < VALUES(expiration_date), VALUES(price), price),
-- 				expiration_date = IF(expiration_date < VALUES(expiration_date), VALUES(expiration_date), expiration_date)
		`,
		values,
	)
	query, args, err := q.ToMysql()
	if err != nil {
		return fmt.Errorf("can't build create prices query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't execute create prices query: %w", err)
	}

	return nil
}

func (r *mysqlPrices) ImportFile(ctx context.Context, filePath string) error {
	q := bqb.New(fmt.Sprintf(`
		LOAD DATA CONCURRENT LOCAL INFILE '%s'
		IGNORE
		INTO TABLE prices
		FIELDS TERMINATED BY ','
		LINES TERMINATED BY '\n'
		(id,price,expiration_date)
	`, filePath))
	query, args, err := q.ToMysql()
	if err != nil {
		return fmt.Errorf("can't build import prices from file=%s query: %w", filePath, err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't execute import prices from file=%s query: %w", filePath, err)
	}

	return nil
}

func (r *mysqlPrices) Get(ctx context.Context, id string) (*models.Price, error) {
	q := bqb.New(
		`
			SELECT id, price, expiration_date FROM prices
			WHERE id = ?
		`,
		id,
	)
	query, args, err := q.ToMysql()
	if err != nil {
		return nil, fmt.Errorf("can't build get price query: %w", err)
	}

	var price models.Price

	row := r.db.QueryRowContext(ctx, query, args...)
	err = row.Scan(&price.ID, &price.Price, &price.ExpirationDate)
	if err != nil {
		if errors.ErrorIs(err, sql.ErrNoRows) {
			return nil, errors.PriceNotFound
		}
		return nil, fmt.Errorf("can't execute get price query: %w", err)
	}

	return &price, nil
}
