//go:generate mockgen -source prices.go -destination prices_mock.go -package api Service

package api

import (
	"context"
	"net/http"
	"prices/pkg/config"
	"prices/pkg/errors"
	"prices/pkg/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	BaseURL = "/api/v0/prices"
)

type (
	Service interface {
		Get(ctx context.Context, id string) (*models.Price, error)
	}

	API struct {
		ServerInterface

		BaseURL string
		logger  *zap.Logger
		config  *config.APIServer

		prices Service
	}
)

var (
	options = GinServerOptions{
		BaseURL: BaseURL,
	}
)

func NewAPI(config *config.APIServer, logger *zap.Logger, srv Service) *API {
	log := logger.Named("PricesAPI")
	api := &API{
		BaseURL: BaseURL,
		config:  config,
		logger:  log,
		prices:  srv,
	}
	return api
}

func (api *API) RegisterHandlers(e *gin.Engine) {
	RegisterHandlersWithOptions(e, api, options)
}

func (api *API) mapErrorToStatus(err error) int {
	if errors.ErrorIs(err, errors.ErrPriceNotFound) {
		return http.StatusNotFound
	}
	if errors.ErrorIs(err, errors.ErrInternal) {
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}

func (api *API) priceToResponse(price *models.Price) Promotion {
	priceData, _ := price.Price.Float64()
	return Promotion{
		Id:             price.ID,
		Price:          priceData,
		ExpirationDate: price.ExpirationDate,
	}
}

// GetPromotion (GET /promotions/{promotion_id})
func (api *API) GetPromotion(c *gin.Context, id PromotionId) {
	price, err := api.prices.Get(c, id)
	if err != nil {
		c.AbortWithStatus(api.mapErrorToStatus(err))
		return
	}
	c.JSON(http.StatusOK, api.priceToResponse(price))
}
