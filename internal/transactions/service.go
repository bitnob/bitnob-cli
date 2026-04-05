package transactions

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

type ResponseMeta struct {
	RequestID string `json:"request_id,omitempty"`
}

type Transaction struct {
	TransactionID   string         `json:"transaction_id"`
	AccountNumber   string         `json:"account_number"`
	Currency        string         `json:"currency"`
	Type            string         `json:"type"`
	State           string         `json:"state"`
	Amount          string         `json:"amount"`
	Fee             string         `json:"fee"`
	Reference       string         `json:"reference"`
	IdempotencyKey  string         `json:"idempotency_key,omitempty"`
	TradeID         any            `json:"trade_id,omitempty"`
	Metadata        map[string]any `json:"metadata"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
	ValueDate       string         `json:"value_date"`
	AmountFormatted string         `json:"amount_formatted"`
	FeeFormatted    string         `json:"fee_formatted"`
	Side            string         `json:"side"`
}

type ListData struct {
	Transactions   []Transaction `json:"transactions"`
	NextCursor     string        `json:"next_cursor,omitempty"`
	PreviousCursor string        `json:"previous_cursor,omitempty"`
	TotalCount     int           `json:"total_count"`
	HasMore        bool          `json:"has_more"`
}

type ListResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      ListData     `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type Response struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      Transaction  `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type ListFilters struct {
	Page              int
	Limit             int
	Cursor            string
	Reference         string
	Hash              string
	Action            string
	Channel           string
	Type              string
	Status            string
	Currency          string
	BankAccountID     string
	CustomerReference string
	WalletID          string
	Address           string
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{configStore: configStore, authStore: authStore, client: client}
}

func (s *Service) List(ctx context.Context, filters ListFilters) (ListResponse, error) {
	query := url.Values{}
	if filters.Page > 0 {
		query.Set("page", strconv.Itoa(filters.Page))
	}
	if filters.Limit > 0 {
		query.Set("limit", strconv.Itoa(filters.Limit))
	}
	if filters.Cursor != "" {
		query.Set("cursor", filters.Cursor)
	}
	if filters.Reference != "" {
		query.Set("reference", filters.Reference)
	}
	if filters.Hash != "" {
		query.Set("hash", filters.Hash)
	}
	if filters.Action != "" {
		query.Set("action", filters.Action)
	}
	if filters.Channel != "" {
		query.Set("channel", filters.Channel)
	}
	if filters.Type != "" {
		query.Set("type", filters.Type)
	}
	if filters.Status != "" {
		query.Set("status", filters.Status)
	}
	if filters.Currency != "" {
		query.Set("currency", filters.Currency)
	}
	if filters.BankAccountID != "" {
		query.Set("bankAccountId", filters.BankAccountID)
	}
	if filters.CustomerReference != "" {
		query.Set("customerReference", filters.CustomerReference)
	}
	if filters.WalletID != "" {
		query.Set("walletId", filters.WalletID)
	}
	if filters.Address != "" {
		query.Set("address", filters.Address)
	}

	path := "/api/transactions"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var response ListResponse
	err := s.doJSON(ctx, "GET", path, nil, &response)
	return response, err
}

func (s *Service) Get(ctx context.Context, idOrReference string) (Response, error) {
	var response Response
	err := s.doJSON(ctx, "GET", "/api/transactions/"+url.PathEscape(idOrReference), nil, &response)
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
		return fmt.Errorf("decode transactions response: %w", err)
	}

	return nil
}
