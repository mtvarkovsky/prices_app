package service

import (
	"context"
	"go.uber.org/zap"
	"prices/pkg/errors"
	"prices/pkg/models"
	"prices/pkg/repository"
)

type (
	Error error

	Prices interface {
		Get(ctx context.Context, id string) (*models.Price, error)
	}

	prices struct {
		logger *zap.Logger
		repo   repository.Prices
	}
)

func NewPrices(logger *zap.Logger, repo repository.Prices) Prices {
	log := logger.Named("PricesService")
	p := &prices{
		logger: log,
		repo:   repo,
	}
	return p
}

func (p *prices) Get(ctx context.Context, id string) (*models.Price, error) {
	price, err := p.repo.Get(ctx, id)
	if err != nil {
		if !errors.ErrorIs(err, errors.PriceNotFound) {
			p.logger.Sugar().Errorf("can't get price, id=%s: (%s)", id, err.Error())
			return nil, errors.InternalError
		}
		return nil, err
	}
	return price, nil
}
