//go:generate mockgen -source repository.go -destination repository_mock.go -package repository Prices

package repository

import (
	"context"
	"fmt"
	"prices/pkg/config"
	"prices/pkg/models"
)

type (
	Prices interface {
		Create(ctx context.Context, prices []*models.Price) error
		Get(ctx context.Context, id string) (*models.Price, error)
		ImportFile(ctx context.Context, filePath string) error
	}
)

func NewPrices(config config.Storage) (Prices, error) {
	switch config.Type {
	case "mysql":
		return newMysqlPrices(config)
	default:
		return nil, fmt.Errorf("can't create storage, unknown type=%s", config.Type)
	}
}
