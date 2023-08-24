package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"prices/pkg/config"
	"prices/pkg/models"
	"prices/pkg/service"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func serveHTTP(
	e *gin.Engine,
	method string,
	url *url.URL,
	body io.Reader,
	headers map[string]string,
	cookies []http.Cookie,
) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, url.String(), body)

	if len(headers) > 0 {
		for header, content := range headers {
			r.Header.Set(header, content)
		}
	}

	if len(cookies) > 0 {
		for _, cookie := range cookies {
			r.AddCookie(&cookie)
		}
	}

	e.ServeHTTP(w, r)
	return w, r
}

func createURL(path string, query string) *url.URL {
	return &url.URL{
		Scheme:   "http",
		Path:     path,
		Host:     "localhost",
		RawQuery: query,
	}
}

func newTestAPI(t *testing.T) (*API, *gin.Engine) {
	ctrl := gomock.NewController(t)
	cfg := &config.APIServer{Port: 8080}
	log := zap.NewNop()
	prices := service.NewMockPrices(ctrl)
	api := &API{
		logger:  log,
		config:  cfg,
		prices:  prices,
		BaseURL: BaseURL,
	}
	e := gin.Default()
	api.RegisterHandlers(e)
	return api, e
}

func TestAPI_GetPromotion(t *testing.T) {
	api, e := newTestAPI(t)
	prcs := api.prices.(*service.MockPrices)

	expectedPrice := &models.Price{
		ID:             "test_id_1",
		Price:          decimal.NewFromFloat(3.14),
		ExpirationDate: time.Now().UTC(),
	}

	expectedPriceData, _ := expectedPrice.Price.Float64()

	expectedResp := Promotion{
		Id:             expectedPrice.ID,
		Price:          expectedPriceData,
		ExpirationDate: expectedPrice.ExpirationDate,
	}

	prcs.EXPECT().
		Get(gomock.Any(), expectedPrice.ID).
		Return(expectedPrice, nil)

	response, _ := serveHTTP(
		e,
		http.MethodGet,
		createURL(fmt.Sprintf("/api/v0/prices/promotions/%s", expectedPrice.ID), ""),
		nil,
		nil,
		nil,
	)

	assert.Equal(t, http.StatusOK, response.Code)

	var respBody Promotion
	err := json.Unmarshal(response.Body.Bytes(), &respBody)
	assert.NoError(t, err)

	assert.Equal(t, expectedResp, respBody)
}
