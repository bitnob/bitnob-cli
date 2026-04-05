package beneficiaries

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

type Destination struct {
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
}

func (d Destination) MarshalJSON() ([]byte, error) {
	payload := struct {
		AccountName   string `json:"account_name"`
		AccountNumber string `json:"account_number"`
		BankCode      string `json:"bank_code"`
	}{
		AccountName:   d.AccountName,
		AccountNumber: d.AccountNumber,
		BankCode:      d.BankCode,
	}

	inner, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(inner))
}

func (d *Destination) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "" || text == "null" {
		*d = Destination{}
		return nil
	}

	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		var encoded string
		if err := json.Unmarshal(data, &encoded); err != nil {
			return err
		}
		encoded = strings.TrimSpace(encoded)
		if encoded == "" || encoded == "{}" {
			*d = Destination{}
			return nil
		}
		return json.Unmarshal([]byte(encoded), d)
	}

	type alias Destination
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*d = Destination(decoded)
	return nil
}

type Beneficiary struct {
	ID            string      `json:"id"`
	CompanyID     string      `json:"company_id"`
	Type          string      `json:"type"`
	Name          string      `json:"name"`
	Destination   Destination `json:"destination"`
	IsBlacklisted bool        `json:"is_blacklisted"`
	IsActive      bool        `json:"is_active"`
	CreatedBy     string      `json:"created_by,omitempty"`
	CreatedAt     string      `json:"created_at,omitempty"`
	UpdatedAt     string      `json:"updated_at,omitempty"`
}

type ResponseMeta struct {
	RequestID string `json:"request_id,omitempty"`
}

type Response struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      Beneficiary  `json:"data"`
	Metadata  ResponseMeta `json:"metadata"`
	Timestamp string       `json:"timestamp"`
}

type ListData struct {
	Beneficiaries []Beneficiary `json:"beneficiaries"`
	Total         int           `json:"total"`
	Limit         int           `json:"limit"`
	Offset        int           `json:"offset"`
}

type ListResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      ListData     `json:"data"`
	Metadata  ResponseMeta `json:"metadata"`
	Timestamp string       `json:"timestamp"`
}

type DeleteData struct {
	Success bool `json:"success"`
}

type DeleteResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      DeleteData   `json:"data"`
	Metadata  ResponseMeta `json:"metadata"`
	Timestamp string       `json:"timestamp"`
}

type CreateInput struct {
	CompanyID   string      `json:"company_id"`
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Destination Destination `json:"destination"`
	CreatedBy   string      `json:"created_by,omitempty"`
}

type UpdateInput struct {
	Name          string      `json:"name"`
	Destination   Destination `json:"destination"`
	IsBlacklisted *bool       `json:"is_blacklisted,omitempty"`
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{
		configStore: configStore,
		authStore:   authStore,
		client:      client,
	}
}

func (s *Service) List(ctx context.Context, limit int, offset int) (ListResponse, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}

	path := "/api/beneficiaries"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := s.do(ctx, "GET", path, nil)
	if err != nil {
		return ListResponse{}, err
	}

	var response ListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return ListResponse{}, fmt.Errorf("decode beneficiaries list response: %w", err)
	}

	return response, nil
}

func (s *Service) Get(ctx context.Context, id string) (Response, error) {
	body, err := s.do(ctx, "GET", "/api/beneficiaries/"+url.PathEscape(id), nil)
	if err != nil {
		return Response{}, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, fmt.Errorf("decode beneficiary response: %w", err)
	}

	return response, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Response, error) {
	bodyBytes, err := json.Marshal(input)
	if err != nil {
		return Response{}, err
	}

	body, err := s.do(ctx, "POST", "/api/beneficiaries", bodyBytes)
	if err != nil {
		return Response{}, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, fmt.Errorf("decode create beneficiary response: %w", err)
	}

	return response, nil
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Response, error) {
	bodyBytes, err := json.Marshal(input)
	if err != nil {
		return Response{}, err
	}

	body, err := s.do(ctx, "PUT", "/api/beneficiaries/"+url.PathEscape(id), bodyBytes)
	if err != nil {
		return Response{}, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, fmt.Errorf("decode update beneficiary response: %w", err)
	}

	return response, nil
}

func (s *Service) Delete(ctx context.Context, id string) (DeleteResponse, error) {
	body, err := s.do(ctx, "DELETE", "/api/beneficiaries/"+url.PathEscape(id), nil)
	if err != nil {
		return DeleteResponse{}, err
	}

	var response DeleteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return DeleteResponse{}, fmt.Errorf("decode delete beneficiary response: %w", err)
	}

	return response, nil
}

func (s *Service) do(ctx context.Context, method string, path string, body []byte) ([]byte, error) {
	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return nil, err
	}

	active, err := profile.Active(cfg)
	if err != nil {
		return nil, err
	}

	credentials, err := s.authStore.LoadCredentials(ctx, active.Name)
	if err != nil {
		return nil, auth.CredentialLoadError(active.Name, err)
	}

	return s.client.Do(ctx, method, path, credentials.ClientID, credentials.SecretKey, body)
}
