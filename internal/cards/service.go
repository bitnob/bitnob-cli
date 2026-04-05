package cards

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

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

type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type SpendingLimits struct {
	SingleTransaction string   `json:"single_transaction,omitempty"`
	Daily             string   `json:"daily,omitempty"`
	Weekly            string   `json:"weekly,omitempty"`
	Monthly           string   `json:"monthly,omitempty"`
	AllowedCategories []string `json:"allowed_categories,omitempty"`
	BlockedCategories []string `json:"blocked_categories,omitempty"`
	AllowedMerchants  []string `json:"allowed_merchants,omitempty"`
	BlockedMerchants  []string `json:"blocked_merchants,omitempty"`
}

type CurrentSpending struct {
	DailySpent   string `json:"daily_spent,omitempty"`
	WeeklySpent  string `json:"weekly_spent,omitempty"`
	MonthlySpent string `json:"monthly_spent,omitempty"`
}

type Card struct {
	ID              string          `json:"id"`
	CompanyID       string          `json:"company_id,omitempty"`
	CustomerID      string          `json:"customer_id"`
	Name            string          `json:"name,omitempty"`
	CardType        string          `json:"card_type"`
	CardBrand       string          `json:"card_brand"`
	LastFour        string          `json:"last_four"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	Balance         string          `json:"balance"`
	SpendingLimits  SpendingLimits  `json:"spending_limits,omitempty"`
	CurrentSpending CurrentSpending `json:"current_spending,omitempty"`
	BillingAddress  Address         `json:"billing_address,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at,omitempty"`
}

type CardData struct {
	Card Card `json:"card"`
}

type Response struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    CardData `json:"data"`
}

type CardListItem struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Name       string `json:"name,omitempty"`
	CardType   string `json:"card_type"`
	CardBrand  string `json:"card_brand"`
	LastFour   string `json:"last_four"`
	Currency   string `json:"currency"`
	Status     string `json:"status"`
	Balance    string `json:"balance"`
	CreatedAt  string `json:"created_at"`
}

type ListData struct {
	Cards  []CardListItem `json:"cards"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit,omitempty"`
	Offset int            `json:"offset,omitempty"`
}

type ListResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    ListData `json:"data"`
}

type CardDetails struct {
	CardID         string  `json:"card_id"`
	CardNumber     string  `json:"card_number"`
	CVV            string  `json:"cvv"`
	ExpiryMonth    string  `json:"expiry_month"`
	ExpiryYear     string  `json:"expiry_year"`
	CardholderName string  `json:"cardholder_name"`
	BillingAddress Address `json:"billing_address"`
}

type DetailsResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    CardDetails `json:"data"`
}

type BalanceOperationData struct {
	CardID        string `json:"card_id"`
	Operation     string `json:"operation"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	NewBalance    string `json:"new_balance"`
	TransactionID string `json:"transaction_id"`
	Reference     string `json:"reference"`
	CreatedAt     string `json:"created_at"`
}

type BalanceOperationResponse struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Data    BalanceOperationData `json:"data"`
}

type FreezeData struct {
	CardID     string `json:"card_id"`
	Status     string `json:"status"`
	FrozenAt   string `json:"frozen_at,omitempty"`
	UnfrozenAt string `json:"unfrozen_at,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type FreezeResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Data    FreezeData `json:"data"`
}

type TerminateData struct {
	CardID           string `json:"card_id"`
	Status           string `json:"status"`
	RemainingBalance string `json:"remaining_balance"`
	BalanceReturned  bool   `json:"balance_returned"`
	TerminatedAt     string `json:"terminated_at"`
	Reason           string `json:"reason,omitempty"`
}

type TerminateResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Data    TerminateData `json:"data"`
}

type SpendingLimitsUpdateData struct {
	CardID         string         `json:"card_id"`
	Operation      string         `json:"operation"`
	SpendingLimits SpendingLimits `json:"spending_limits"`
	UpdatedAt      string         `json:"updated_at"`
}

type SpendingLimitsUpdateResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Data    SpendingLimitsUpdateData `json:"data"`
}

type CreateInput struct {
	CustomerID     string         `json:"customer_id"`
	Name           string         `json:"name"`
	CardType       string         `json:"card_type"`
	CardBrand      string         `json:"card_brand"`
	Currency       string         `json:"currency"`
	BillingAddress Address        `json:"billing_address"`
	SpendingLimits SpendingLimits `json:"spending_limits,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedBy      string         `json:"created_by,omitempty"`
}

type FundWithdrawInput struct {
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description,omitempty"`
	Reference   string `json:"reference"`
}

type ReasonInput struct {
	Reason string `json:"reason,omitempty"`
}

type CustomerCardsResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Cards []CardListItem `json:"cards"`
		Total int            `json:"total"`
	} `json:"data"`
}

type ListFilters struct {
	Status     string
	CardType   string
	CustomerID string
	Limit      int
	Offset     int
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{configStore: configStore, authStore: authStore, client: client}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Response, error) {
	if input.CreatedBy == "" {
		credentials, err := s.activeCredentials(ctx)
		if err != nil {
			return Response{}, err
		}
		input.CreatedBy = credentials.ClientID
	}

	var response Response
	err := s.doJSON(ctx, "POST", "/api/cards", input, &response)
	return response, err
}

func (s *Service) Get(ctx context.Context, cardID string) (Response, error) {
	var response Response
	err := s.doJSON(ctx, "GET", "/api/cards/"+url.PathEscape(cardID), nil, &response)
	return response, err
}

func (s *Service) GetDetails(ctx context.Context, cardID string) (DetailsResponse, error) {
	var response DetailsResponse
	err := s.doJSON(ctx, "GET", "/api/cards/"+url.PathEscape(cardID)+"/details", nil, &response)
	return response, err
}

func (s *Service) List(ctx context.Context, filters ListFilters) (ListResponse, error) {
	query := url.Values{}
	if filters.Status != "" {
		query.Set("status", filters.Status)
	}
	if filters.CardType != "" {
		query.Set("card_type", filters.CardType)
	}
	if filters.CustomerID != "" {
		query.Set("customer_id", filters.CustomerID)
	}
	if filters.Limit > 0 {
		query.Set("limit", strconv.Itoa(filters.Limit))
	}
	if filters.Offset > 0 {
		query.Set("offset", strconv.Itoa(filters.Offset))
	}

	path := "/api/cards"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var response ListResponse
	err := s.doJSON(ctx, "GET", path, nil, &response)
	return response, err
}

func (s *Service) Fund(ctx context.Context, cardID string, input FundWithdrawInput) (BalanceOperationResponse, error) {
	var response BalanceOperationResponse
	err := s.doJSON(ctx, "POST", "/api/cards/"+url.PathEscape(cardID)+"/fund", input, &response)
	return response, err
}

func (s *Service) Withdraw(ctx context.Context, cardID string, input FundWithdrawInput) (BalanceOperationResponse, error) {
	var response BalanceOperationResponse
	err := s.doJSON(ctx, "POST", "/api/cards/"+url.PathEscape(cardID)+"/withdraw", input, &response)
	return response, err
}

func (s *Service) Freeze(ctx context.Context, cardID string, input ReasonInput) (FreezeResponse, error) {
	var response FreezeResponse
	err := s.doJSON(ctx, "POST", "/api/cards/"+url.PathEscape(cardID)+"/freeze", input, &response)
	return response, err
}

func (s *Service) Unfreeze(ctx context.Context, cardID string) (FreezeResponse, error) {
	var response FreezeResponse
	err := s.doJSON(ctx, "POST", "/api/cards/"+url.PathEscape(cardID)+"/unfreeze", nil, &response)
	return response, err
}

func (s *Service) Terminate(ctx context.Context, cardID string, input ReasonInput) (TerminateResponse, error) {
	var response TerminateResponse
	err := s.doJSON(ctx, "POST", "/api/cards/"+url.PathEscape(cardID)+"/terminate", input, &response)
	return response, err
}

func (s *Service) UpdateSpendingLimits(ctx context.Context, cardID string, input SpendingLimits) (SpendingLimitsUpdateResponse, error) {
	var response SpendingLimitsUpdateResponse
	err := s.doJSON(ctx, "PUT", "/api/cards/"+url.PathEscape(cardID)+"/spending-limits", input, &response)
	return response, err
}

func (s *Service) GetCustomerCards(ctx context.Context, customerID string) (CustomerCardsResponse, error) {
	var response CustomerCardsResponse
	err := s.doJSON(ctx, "GET", "/api/customers/"+url.PathEscape(customerID)+"/cards", nil, &response)
	return response, err
}

func (s *Service) doJSON(ctx context.Context, method string, path string, payload any, out any) error {
	credentials, err := s.activeCredentials(ctx)
	if err != nil {
		return err
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
		return fmt.Errorf("decode cards response: %w", err)
	}

	return nil
}

func (s *Service) activeCredentials(ctx context.Context) (auth.Credentials, error) {
	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return auth.Credentials{}, err
	}

	active, err := profile.Active(cfg)
	if err != nil {
		return auth.Credentials{}, err
	}

	credentials, err := s.authStore.LoadCredentials(ctx, active.Name)
	if err != nil {
		return auth.Credentials{}, auth.CredentialLoadError(active.Name, err)
	}

	return credentials, nil
}
