package payouts

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

type Quote struct {
	ID                 string         `json:"id"`
	Status             string         `json:"status"`
	SettlementCurrency string         `json:"settlement_currency,omitempty"`
	QuoteID            string         `json:"quote_id,omitempty"`
	SettlementAmount   any            `json:"settlement_amount,omitempty"`
	BTCRate            any            `json:"btc_rate,omitempty"`
	ExchangeRate       any            `json:"exchange_rate,omitempty"`
	ExpiryTimestamp    any            `json:"expiry_timestamp,omitempty"`
	Amount             any            `json:"amount,omitempty"`
	SatAmount          any            `json:"sat_amount,omitempty"`
	ExpiresInText      string         `json:"expires_in_text,omitempty"`
	QuoteText          string         `json:"quote_text,omitempty"`
	Reference          string         `json:"reference,omitempty"`
	CustomerID         string         `json:"customer_id,omitempty"`
	PaymentReason      string         `json:"payment_reason,omitempty"`
	CallbackURL        string         `json:"callback_url,omitempty"`
	BeneficiaryID      string         `json:"beneficiary_id,omitempty"`
	ClientMetaData     map[string]any `json:"client_meta_data,omitempty"`
	CentAmount         string         `json:"cent_amount,omitempty"`
	CentFees           string         `json:"cent_fees,omitempty"`
	Fees               any            `json:"fees,omitempty"`
	Address            string         `json:"address,omitempty"`
	Source             string         `json:"source,omitempty"`
	FromAsset          string         `json:"from_asset,omitempty"`
	Chain              string         `json:"chain,omitempty"`
	ToCurrency         string         `json:"to_currency,omitempty"`
	PaymentETA         string         `json:"payment_eta,omitempty"`
	Expiry             string         `json:"expiry,omitempty"`
	CreatedAt          string         `json:"created_at,omitempty"`
	UpdatedAt          string         `json:"updated_at,omitempty"`
	Beneficiary        map[string]any `json:"beneficiary,omitempty"`
	BeneficiaryDetails map[string]any `json:"beneficiary_details,omitempty"`
	Trip               map[string]any `json:"trip,omitempty"`
	Company            map[string]any `json:"company,omitempty"`
	AllowedFeatures    map[string]any `json:"allowed_features,omitempty"`
}

type QuoteResponse struct {
	Status    string       `json:"status"`
	Message   string       `json:"message"`
	Data      QuoteData    `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type QuoteData struct {
	Quote Quote `json:"quote"`
}

type QuoteListResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		OffRamps []Quote        `json:"off_ramps"`
		Meta     PaginationMeta `json:"meta,omitempty"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type QuoteFetchResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		OffRamps []Quote `json:"off_ramps"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type PaginationMeta struct {
	Page            int  `json:"page"`
	Take            int  `json:"take"`
	ItemCount       int  `json:"item_count"`
	PageCount       int  `json:"page_count"`
	HasPreviousPage bool `json:"has_previous_page"`
	HasNextPage     bool `json:"has_next_page"`
}

type CreateQuoteInput struct {
	FromAsset        string `json:"from_asset"`
	ToCurrency       string `json:"to_currency"`
	Source           string `json:"source"`
	Chain            string `json:"chain,omitempty"`
	Amount           string `json:"amount,omitempty"`
	SettlementAmount string `json:"settlement_amount,omitempty"`
	PaymentReason    string `json:"payment_reason"`
	Reference        string `json:"reference"`
	ClientMetaData   string `json:"client_meta_data,omitempty"`
	Country          string `json:"country"`
}

type InitializeInput struct {
	CustomerID     string         `json:"customer_id"`
	Beneficiary    map[string]any `json:"beneficiary"`
	Reference      string         `json:"reference"`
	PaymentReason  string         `json:"payment_reason"`
	ClientMetaData string         `json:"client_meta_data,omitempty"`
	CallbackURL    string         `json:"callback_url,omitempty"`
}

type InitializeResponse struct {
	Status    string       `json:"status"`
	Message   string       `json:"message"`
	Data      Quote        `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type BeneficiaryLookupInput struct {
	Country       string `json:"country"`
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	Type          string `json:"type"`
}

type BeneficiaryLookupResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AccountName   string `json:"account_name"`
		AccountNumber string `json:"account_number"`
		BankCode      string `json:"bank_code"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type SimulateDepositInput struct {
	QuoteID string `json:"quote_id"`
	Amount  string `json:"amount"`
	TxHash  string `json:"tx_hash"`
}

type SimulateDepositResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Success bool `json:"success"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type CountriesResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Countries []map[string]any `json:"countries,omitempty"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type CountryRequirementsResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Code        string                 `json:"code"`
		Name        string                 `json:"name"`
		Flag        string                 `json:"flag"`
		DialCode    string                 `json:"dial_code"`
		Destination map[string][]FieldSpec `json:"destination"`
	} `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type FieldSpec struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Const    string   `json:"const,omitempty"`
	Enum     []string `json:"enum,omitempty"`
}

type LimitsResponse struct {
	Status    string       `json:"status"`
	Message   string       `json:"message"`
	Data      []Limit      `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type Limit struct {
	LowerLimit     string `json:"lower_limit"`
	HigherLimit    string `json:"higher_limit"`
	Currency       string `json:"currency"`
	Country        string `json:"country"`
	Rate           string `json:"rate"`
	USDLowerLimit  string `json:"usd_lower_limit"`
	USDHigherLimit string `json:"usd_higher_limit"`
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{configStore: configStore, authStore: authStore, client: client}
}

func (s *Service) CreateQuote(ctx context.Context, input CreateQuoteInput) (QuoteResponse, error) {
	var response QuoteResponse
	err := s.doJSON(ctx, "POST", "/api/payouts/quotes", input, &response)
	return response, err
}

func (s *Service) Initialize(ctx context.Context, quoteID string, input InitializeInput) (InitializeResponse, error) {
	var response InitializeResponse
	err := s.doJSON(ctx, "POST", "/api/payouts/quotes/"+url.PathEscape(quoteID)+"/initialize", input, &response)
	return response, err
}

func (s *Service) BeneficiaryLookup(ctx context.Context, input BeneficiaryLookupInput) (BeneficiaryLookupResponse, error) {
	var response BeneficiaryLookupResponse
	err := s.doJSON(ctx, "POST", "/api/payouts/beneficiary-lookup", input, &response)
	return response, err
}

func (s *Service) SimulateDeposit(ctx context.Context, input SimulateDepositInput) (SimulateDepositResponse, error) {
	var response SimulateDepositResponse
	err := s.doJSON(ctx, "POST", "/api/payouts/simulate-deposit", input, &response)
	return response, err
}

func (s *Service) Finalize(ctx context.Context, quoteID string) (InitializeResponse, error) {
	var response InitializeResponse
	err := s.doJSON(ctx, "POST", "/api/payouts/quotes/"+url.PathEscape(quoteID)+"/finalize", nil, &response)
	return response, err
}

func (s *Service) ListQuotes(ctx context.Context, order string, page int, take int) (QuoteListResponse, error) {
	query := url.Values{}
	if order != "" {
		query.Set("order", order)
	}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if take > 0 {
		query.Set("take", strconv.Itoa(take))
	}
	path := "/api/payouts/"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var response QuoteListResponse
	err := s.doJSON(ctx, "GET", path, nil, &response)
	return response, err
}

func (s *Service) GetQuoteByQuoteID(ctx context.Context, quoteID string) (QuoteFetchResponse, error) {
	var response QuoteFetchResponse
	err := s.doJSON(ctx, "GET", "/api/payouts/quotes/"+url.PathEscape(quoteID), nil, &response)
	return response, err
}

func (s *Service) GetQuoteByID(ctx context.Context, id string) (QuoteFetchResponse, error) {
	var response QuoteFetchResponse
	err := s.doJSON(ctx, "GET", "/api/payouts/fetch/"+url.PathEscape(id), nil, &response)
	return response, err
}

func (s *Service) GetQuoteByReference(ctx context.Context, reference string) (QuoteFetchResponse, error) {
	var response QuoteFetchResponse
	err := s.doJSON(ctx, "GET", "/api/payouts/fetch/reference/"+url.PathEscape(reference), nil, &response)
	return response, err
}

func (s *Service) CountryRequirements(ctx context.Context, countryCode string) (CountryRequirementsResponse, error) {
	var response CountryRequirementsResponse
	err := s.doJSON(ctx, "GET", "/api/payouts/supported-countries/"+url.PathEscape(countryCode)+"/requirements", nil, &response)
	return response, err
}

func (s *Service) SupportedCountries(ctx context.Context) (CountriesResponse, error) {
	var response CountriesResponse
	err := s.doJSON(ctx, "GET", "/api/payouts/supported-countries", nil, &response)
	return response, err
}

func (s *Service) Limits(ctx context.Context) (LimitsResponse, error) {
	var response LimitsResponse
	err := s.doJSON(ctx, "GET", "/api/payouts/limits", nil, &response)
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
		return fmt.Errorf("decode payouts response: %w", err)
	}

	return nil
}
