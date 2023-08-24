package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	Price struct {
		ID             string          `db:"id"`
		Price          decimal.Decimal `db:"price"`
		ExpirationDate time.Time       `db:"expiration_date"`
	}
)
