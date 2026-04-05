package customers

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

type Customer struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	CustomerType string `json:"customer_type,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Email        string `json:"email"`
	Phone        string `json:"phone,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	Blacklist    bool   `json:"blacklist"`
	Metadata     string `json:"metadata,omitempty"`
	IsActive     bool   `json:"is_active,omitempty"`
	CompanyID    string `json:"company_id,omitempty"`
	CreatedBy    string `json:"created_by,omitempty"`
}

type CustomerResponse struct {
	Success   *bool         `json:"success,omitempty"`
	Status    *bool         `json:"status,omitempty"`
	Message   string        `json:"message"`
	Data      Customer      `json:"data"`
	Metadata  *ResponseMeta `json:"metadata,omitempty"`
	Timestamp string        `json:"timestamp,omitempty"`
}

type CustomerListResponse struct {
	Success   *bool            `json:"success,omitempty"`
	Status    *bool            `json:"status,omitempty"`
	Message   string           `json:"message"`
	Data      CustomerListData `json:"data"`
	Metadata  *ResponseMeta    `json:"metadata,omitempty"`
	Timestamp string           `json:"timestamp,omitempty"`
}

type ResponseMeta struct {
	RequestID string `json:"request_id,omitempty"`
}

type CustomerListData struct {
	Customers []Customer       `json:"customers"`
	Meta      CustomerListMeta `json:"meta"`
}

type CustomerListMeta struct {
	Page            int  `json:"page"`
	Take            int  `json:"take"`
	ItemCount       int  `json:"item_count"`
	PageCount       int  `json:"page_count"`
	HasPreviousPage bool `json:"has_previous_page"`
	HasNextPage     bool `json:"has_next_page"`
}

type CreateCustomerInput struct {
	CustomerType string `json:"customer_type"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Email        string `json:"email"`
	Phone        string `json:"phone,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
}

type UpdateCustomerInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

type BlacklistCustomerInput struct {
	Blacklist bool `json:"blacklist"`
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{
		configStore: configStore,
		authStore:   authStore,
		client:      client,
	}
}

func (s *Service) List(ctx context.Context, page int, limit int, blacklist *bool) (CustomerListResponse, error) {
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if blacklist != nil {
		query.Set("blacklist", strconv.FormatBool(*blacklist))
	}

	path := "/api/customers"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := s.do(ctx, "GET", path, nil)
	if err != nil {
		return CustomerListResponse{}, err
	}

	var response CustomerListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return CustomerListResponse{}, fmt.Errorf("decode customers list response: %w", err)
	}

	return response, nil
}

func (s *Service) Get(ctx context.Context, identifier string) (CustomerResponse, error) {
	body, err := s.do(ctx, "GET", "/api/customers/"+url.PathEscape(identifier), nil)
	if err != nil {
		return CustomerResponse{}, err
	}

	var response CustomerResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return CustomerResponse{}, fmt.Errorf("decode customer response: %w", err)
	}

	return response, nil
}

func (s *Service) Create(ctx context.Context, input CreateCustomerInput) (CustomerResponse, error) {
	bodyBytes, err := json.Marshal(input)
	if err != nil {
		return CustomerResponse{}, err
	}

	body, err := s.do(ctx, "POST", "/api/customers", bodyBytes)
	if err != nil {
		return CustomerResponse{}, err
	}

	var response CustomerResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return CustomerResponse{}, fmt.Errorf("decode create customer response: %w", err)
	}

	return response, nil
}

func (s *Service) Update(ctx context.Context, id string, input UpdateCustomerInput) (CustomerResponse, error) {
	bodyBytes, err := json.Marshal(input)
	if err != nil {
		return CustomerResponse{}, err
	}

	body, err := s.do(ctx, "PUT", "/api/customers/"+url.PathEscape(id), bodyBytes)
	if err != nil {
		return CustomerResponse{}, err
	}

	var response CustomerResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return CustomerResponse{}, fmt.Errorf("decode update customer response: %w", err)
	}

	return response, nil
}

func (s *Service) Delete(ctx context.Context, id string) ([]byte, error) {
	return s.do(ctx, "DELETE", "/api/customers/"+url.PathEscape(id), nil)
}

func (s *Service) Blacklist(ctx context.Context, id string, blacklist bool) (CustomerResponse, error) {
	bodyBytes, err := json.Marshal(BlacklistCustomerInput{Blacklist: blacklist})
	if err != nil {
		return CustomerResponse{}, err
	}

	body, err := s.do(ctx, "POST", "/api/customers/blacklist/"+url.PathEscape(id), bodyBytes)
	if err != nil {
		return CustomerResponse{}, err
	}

	var response CustomerResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return CustomerResponse{}, fmt.Errorf("decode blacklist customer response: %w", err)
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
