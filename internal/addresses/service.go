package addresses

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

type Timestamp struct {
	Seconds int64 `json:"seconds"`
	Nanos   int64 `json:"nanos"`
}

type Address struct {
	ID         string    `json:"id"`
	Address    string    `json:"address"`
	CompanyID  string    `json:"company_id,omitempty"`
	CustomerID string    `json:"customer_id,omitempty"`
	Label      string    `json:"label,omitempty"`
	Chain      string    `json:"chain"`
	Status     string    `json:"status"`
	Reference  string    `json:"reference,omitempty"`
	CreatedAt  Timestamp `json:"created_at"`
}

type ListData struct {
	Addresses  []Address `json:"addresses"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	Total      int       `json:"total"`
	TotalPages int       `json:"total_pages"`
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
	Data      Address      `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type CreateInput struct {
	Chain     string `json:"chain"`
	Label     string `json:"label,omitempty"`
	Reference string `json:"reference,omitempty"`
}

type ValidateInput struct {
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

type ValidateData struct {
	Valid   bool   `json:"valid"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

type ValidateResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      ValidateData `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type Token struct {
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

type SupportedChain struct {
	Chain       string  `json:"chain"`
	NativeToken Token   `json:"nativeToken"`
	Stablecoins []Token `json:"stablecoins"`
}

type SupportedChainsData struct {
	Chains []SupportedChain `json:"chains"`
}

type SupportedChainsResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    SupportedChainsData `json:"data"`
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{configStore: configStore, authStore: authStore, client: client}
}

func (s *Service) List(ctx context.Context, page int, limit int) (ListResponse, error) {
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	path := "/api/addresses"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var response ListResponse
	err := s.doJSON(ctx, "GET", path, nil, &response)
	return response, err
}

func (s *Service) Get(ctx context.Context, id string) (Response, error) {
	var response Response
	err := s.doJSON(ctx, "GET", "/api/addresses/"+url.PathEscape(id), nil, &response)
	return response, err
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Response, error) {
	var response Response
	err := s.doJSON(ctx, "POST", "/api/addresses", input, &response)
	return response, err
}

func (s *Service) Validate(ctx context.Context, input ValidateInput) (ValidateResponse, error) {
	var response ValidateResponse
	err := s.doJSON(ctx, "POST", "/api/addresses/validate", input, &response)
	return response, err
}

func (s *Service) SupportedChains(ctx context.Context) (SupportedChainsResponse, error) {
	var response SupportedChainsResponse
	err := s.doJSON(ctx, "GET", "/api/stablecoins/supported-chains", nil, &response)
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
		return fmt.Errorf("decode addresses response: %w", err)
	}

	return nil
}
