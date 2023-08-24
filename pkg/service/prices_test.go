package service

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"prices/pkg/models"
	"prices/pkg/repository"
	"testing"
	"time"
)

func newTestPrices(t *testing.T) *prices {
	ctrl := gomock.NewController(t)
	log := zap.NewNop()
	repo := repository.NewMockPrices(ctrl)
	prcs := NewPrices(log, repo)

	p := prcs.(*prices)
	return p
}

func TestPrices_Get(t *testing.T) {
	prcs := newTestPrices(t)
	repo := prcs.repo.(*repository.MockPrices)
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
