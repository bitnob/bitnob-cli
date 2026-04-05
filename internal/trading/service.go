package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/bitnob/bitnob-cli/internal/api"
	"github.com/bitnob/bitnob-cli/internal/auth"
	"github.com/bitnob/bitnob-cli/internal/config"
	"github.com/bitnob/bitnob-cli/internal/profile"
)

type Service struct {
	configStore *config.Store
	authStore   *auth.Store
	client      *api.Client
}

type ResponseMeta struct {
	RequestID string `json:"request_id,omitempty"`
}

type Exchange struct {
	SendQuantity    string `json:"send_quantity"`
	SendCurrency    string `json:"send_currency"`
	ReceiveQuantity string `json:"receive_quantity"`
	ReceiveCurrency string `json:"receive_currency"`
}

type Quote struct {
	ID                string   `json:"id"`
	BaseCurrency      string   `json:"base_currency"`
	QuoteCurrency     string   `json:"quote_currency"`
	Side              string   `json:"side"`
	Quantity          string   `json:"quantity"`
	Price             string   `json:"price"`
	ExpiresAt         string   `json:"expires_at"`
	CreatedAt         string   `json:"created_at"`
	OriginalQuantity  string   `json:"original_quantity"`
	ConsumedQuantity  string   `json:"consumed_quantity"`
	RemainingQuantity string   `json:"remaining_quantity"`
	IsExhausted       bool     `json:"is_exhausted"`
	Exchange          Exchange `json:"exchange"`
}

type QuoteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Quote Quote `json:"quote"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type CreateQuoteInput struct {
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
	Side          string `json:"side"`
	Quantity      string `json:"quantity"`
}

type Fill struct {
	ID        string `json:"id"`
	OrderID   string `json:"order_id"`
	Quantity  string `json:"quantity"`
	CreatedAt string `json:"created_at"`
}

type QuoteConsumption struct {
	QuoteID           string `json:"quote_id"`
	ConsumedQuantity  string `json:"consumed_quantity"`
	RemainingQuantity string `json:"remaining_quantity"`
}

type Order struct {
	ID                string `json:"id"`
	BaseCurrency      string `json:"base_currency"`
	QuoteCurrency     string `json:"quote_currency"`
	Side              string `json:"side"`
	OrderType         string `json:"order_type"`
	Quantity          string `json:"quantity"`
	Price             string `json:"price"`
	Status            string `json:"status"`
	Exchange          string `json:"exchange,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	FilledQuantity    string `json:"filled_quantity"`
	RemainingQuantity string `json:"remaining_quantity"`
	QuoteID           string `json:"quote_id,omitempty"`
}

type OrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Order            Order            `json:"order"`
		Fills            []Fill           `json:"fills,omitempty"`
		QuoteConsumption QuoteConsumption `json:"quote_consumption,omitempty"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type OrdersListResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Orders []Order `json:"orders"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type CreateOrderInput struct {
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
	Side          string `json:"side"`
	Quantity      string `json:"quantity"`
	Price         string `json:"price"`
	QuoteID       string `json:"quote_id"`
	Reference     string `json:"reference,omitempty"`
}

type PricesListResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Prices []Price `json:"prices"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type Price struct {
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
	Price         string `json:"price"`
	AsOf          string `json:"as_of"`
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{configStore: configStore, authStore: authStore, client: client}
}

func (s *Service) CreateQuote(ctx context.Context, input CreateQuoteInput) (QuoteResponse, error) {
	var response QuoteResponse
	err := s.doJSON(ctx, "POST", "/api/trading/quotes", input, &response)
	return response, err
}

func (s *Service) CreateOrder(ctx context.Context, input CreateOrderInput) (OrderResponse, error) {
	var response OrderResponse
	err := s.doJSON(ctx, "POST", "/api/trading/orders", input, &response)
	return response, err
}

func (s *Service) ListOrders(ctx context.Context) (OrdersListResponse, error) {
	var response OrdersListResponse
	err := s.doJSON(ctx, "GET", "/api/trading/orders", nil, &response)
	return response, err
}

func (s *Service) GetOrder(ctx context.Context, id string) (OrderResponse, error) {
	var response OrderResponse
	err := s.doJSON(ctx, "GET", "/api/trading/orders/"+url.PathEscape(id), nil, &response)
	return response, err
}

func (s *Service) ListPrices(ctx context.Context) (PricesListResponse, error) {
	var response PricesListResponse
	err := s.doJSON(ctx, "GET", "/api/trading/prices", nil, &response)
	return response, err
}

func (s *Service) doJSON(ctx context.Context, method string, path string, payload any, out any) error {
	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return err
	}

	active, err := profile.Active(cfg)
	if err != nil {
		return err
	}

	credentials, err := s.authStore.LoadCredentials(ctx, active.Name)
	if err != nil {
		return auth.CredentialLoadError(active.Name, err)
	}

	var body []byte
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	responseBody, err := s.client.Do(ctx, method, path, credentials.ClientID, credentials.SecretKey, body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(responseBody, out); err != nil {
		return fmt.Errorf("decode trading response: %w", err)
	}

	return nil
}
