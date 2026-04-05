package lightning

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

type ResponseMeta struct {
	RequestID string `json:"request_id,omitempty"`
}

type Invoice struct {
	ID          string     `json:"id"`
	PaymentHash string     `json:"payment_hash"`
	Request     string     `json:"request"`
	Preimage    string     `json:"preimage,omitempty"`
	AmountSat   Int64Value `json:"amount_sat"`
	AmountMsat  Int64Value `json:"amount_msat,omitempty"`
	FeeSat      Int64Value `json:"fee_sat,omitempty"`
	FeeMsat     Int64Value `json:"fee_msat,omitempty"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Direction   string     `json:"direction,omitempty"`
	CompanyID   string     `json:"company_id,omitempty"`
	CustomerID  string     `json:"customer_id,omitempty"`
	Reference   string     `json:"reference,omitempty"`
	ExpiresAt   any        `json:"expires_at"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at,omitempty"`
}

type InvoiceResponse struct {
	Success   *bool        `json:"success,omitempty"`
	Status    *bool        `json:"status,omitempty"`
	Message   string       `json:"message"`
	Data      Invoice      `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type InvoiceListData struct {
	Invoices   []Invoice `json:"invoices"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

type InvoiceListResponse struct {
	Success   *bool           `json:"success,omitempty"`
	Status    *bool           `json:"status,omitempty"`
	Message   string          `json:"message"`
	Data      InvoiceListData `json:"data"`
	Metadata  ResponseMeta    `json:"metadata,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

type CreateInvoiceInput struct {
	CustomerID  string `json:"customer_id,omitempty"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Expiry      int64  `json:"expiry,omitempty"`
	Reference   string `json:"reference"`
}

type DecodeInput struct {
	Request string `json:"request"`
}

type DecodeData struct {
	PaymentHash string     `json:"payment_hash"`
	Destination string     `json:"destination"`
	AmountSat   Int64Value `json:"amount_sat"`
	AmountMsat  Int64Value `json:"amount_msat"`
	Description string     `json:"description"`
	ExpiresAt   Int64Value `json:"expires_at"`
	IsExpired   bool       `json:"is_expired,omitempty"`
}

type DecodeResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      DecodeData   `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type InitiatePaymentInput struct {
	Request string `json:"request"`
}

type InitiatePaymentData struct {
	PaymentHash string `json:"payment_hash"`
	Destination string `json:"destination"`
	AmountSat   string `json:"amount_sat"`
	FeeSat      string `json:"fee_sat"`
	TotalSat    string `json:"total_sat"`
	Description string `json:"description"`
	ExpiresAt   string `json:"expires_at"`
	IsExpired   bool   `json:"is_expired"`
	RouteFound  bool   `json:"route_found"`
}

type InitiatePaymentResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	Data      InitiatePaymentData `json:"data"`
	Metadata  ResponseMeta        `json:"metadata,omitempty"`
	Timestamp string              `json:"timestamp,omitempty"`
}

type SendPaymentInput struct {
	CustomerID string `json:"customer_id,omitempty"`
	Request    string `json:"request"`
	Reference  string `json:"reference"`
	MaxFee     int64  `json:"max_fee,omitempty"`
}

type SendPaymentData struct {
	ID            string     `json:"id"`
	PaymentHash   string     `json:"payment_hash"`
	Preimage      string     `json:"preimage"`
	AmountSat     Int64Value `json:"amount_sat"`
	FeeSat        Int64Value `json:"fee_sat"`
	Status        string     `json:"status"`
	CreatedAt     string     `json:"created_at"`
	Reference     string     `json:"reference"`
	TransactionID string     `json:"transaction_id"`
}

type SendPaymentResponse struct {
	Success   bool            `json:"success"`
	Message   string          `json:"message"`
	Data      SendPaymentData `json:"data"`
	Metadata  ResponseMeta    `json:"metadata,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

type VerifyInvoiceInput struct {
	PaymentHash string `json:"payment_hash"`
}

type VerifyInvoiceData struct {
	ID             string     `json:"id"`
	PaymentHash    string     `json:"payment_hash"`
	Preimage       string     `json:"preimage"`
	AmountSat      Int64Value `json:"amount_sat"`
	Status         string     `json:"status"`
	IsPaid         bool       `json:"is_paid"`
	LedgerCredited bool       `json:"ledger_credited"`
	SettledAt      any        `json:"settled_at"`
}

type Int64Value int64

func (v *Int64Value) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "" || text == "null" {
		*v = 0
		return nil
	}

	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		text = text[1 : len(text)-1]
	}

	n, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return fmt.Errorf("parse int64 value %q: %w", text, err)
	}

	*v = Int64Value(n)
	return nil
}

type VerifyInvoiceResponse struct {
	Success   *bool             `json:"success,omitempty"`
	Status    *bool             `json:"status,omitempty"`
	Message   string            `json:"message"`
	Data      VerifyInvoiceData `json:"data"`
	Metadata  ResponseMeta      `json:"metadata,omitempty"`
	Timestamp string            `json:"timestamp,omitempty"`
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{configStore: configStore, authStore: authStore, client: client}
}

func (s *Service) CreateInvoice(ctx context.Context, input CreateInvoiceInput) (InvoiceResponse, error) {
	var response InvoiceResponse
	err := s.doJSON(ctx, "POST", "/api/lightning/invoices", input, &response)
	return response, err
}

func (s *Service) Decode(ctx context.Context, input DecodeInput) (DecodeResponse, error) {
	var response DecodeResponse
	err := s.doJSON(ctx, "POST", "/api/lightning/decode", input, &response)
	return response, err
}

func (s *Service) InitiatePayment(ctx context.Context, input InitiatePaymentInput) (InitiatePaymentResponse, error) {
	var response InitiatePaymentResponse
	err := s.doJSON(ctx, "POST", "/api/lightning/payments/initiate", input, &response)
	return response, err
}

func (s *Service) SendPayment(ctx context.Context, input SendPaymentInput) (SendPaymentResponse, error) {
	var response SendPaymentResponse
	err := s.doJSON(ctx, "POST", "/api/lightning/payments/send", input, &response)
	return response, err
}

func (s *Service) GetInvoice(ctx context.Context, id string) (InvoiceResponse, error) {
	var response InvoiceResponse
	err := s.doJSON(ctx, "GET", "/api/lightning/invoices/"+url.PathEscape(id), nil, &response)
	return response, err
}

func (s *Service) ListInvoices(ctx context.Context, page int, pageSize int, status string, direction string) (InvoiceListResponse, error) {
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		query.Set("page_size", strconv.Itoa(pageSize))
	}
	if status != "" {
		query.Set("status", status)
	}
	if direction != "" {
		query.Set("direction", direction)
	}

	path := "/api/lightning/invoices"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var response InvoiceListResponse
	err := s.doJSON(ctx, "GET", path, nil, &response)
	return response, err
}

func (s *Service) VerifyInvoice(ctx context.Context, input VerifyInvoiceInput) (VerifyInvoiceResponse, error) {
	var response VerifyInvoiceResponse
	err := s.doJSON(ctx, "POST", "/api/lightning/invoices/verify", input, &response)
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
		return fmt.Errorf("decode lightning response: %w", err)
	}

	return nil
}
