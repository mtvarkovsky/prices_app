//go:generate mockgen -source prices.go -destination repository_mock.go -package service Repository

package service

import (
	"context"
	"prices/pkg/errors"
	"prices/pkg/models"

	"go.uber.org/zap"
)

type (
	Repository interface {
		CreateMany(ctx context.Context, prices []*models.Price) error
		Get(ctx context.Context, id string) (*models.Price, error)
		ImportFile(ctx context.Context, filePath string) error
	}

	Prices struct {
		logger *zap.Logger
		repo   Repository
	}
)

func NewPrices(logger *zap.Logger, repo Repository) *Prices {
	log := logger.Named("PricesService")
	p := &Prices{
		logger: log,
		repo:   repo,
	}
	return p
}

func (p *Prices) Get(ctx context.Context, id string) (*models.Price, error) {
	price, err := p.repo.Get(ctx, id)
	if err != nil {
		if !errors.ErrorIs(err, errors.ErrPriceNotFound) {
			p.logger.Sugar().Errorf("can't get price, id=%s: (%s)", id, err.Error())
			return nil, errors.ErrInternal
		}
		return nil, err
	}
	return price, nil
}
