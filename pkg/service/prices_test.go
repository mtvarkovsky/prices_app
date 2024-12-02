package service

import (
	"context"
	"prices/pkg/models"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestPrices(t *testing.T) *Prices {
	ctrl := gomock.NewController(t)
	log := zap.NewNop()
	repo := NewMockRepository(ctrl)
	prcs := NewPrices(log, repo)

	return prcs
}

func TestPrices_Get(t *testing.T) {
	prcs := newTestPrices(t)
	repo := prcs.repo.(*MockRepository)
	expectedPrice := &models.Price{
		ID:             "test_id_1",
		Price:          decimal.NewFromFloat(3.14),
		ExpirationDate: time.Now(),
	}
	repo.EXPECT().
		Get(gomock.Any(), expectedPrice.ID).
		Return(expectedPrice, nil)
	res, err := prcs.Get(context.Background(), expectedPrice.ID)
	assert.NoError(t, err)
	assert.Equal(t, expectedPrice, res)
}
