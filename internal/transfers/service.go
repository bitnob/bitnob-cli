package transfers

import (
	"context"
	"encoding/json"
	"fmt"

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

type Transfer struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Address       string `json:"address"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Chain         string `json:"chain"`
	Network       string `json:"network"`
	Reference     string `json:"reference,omitempty"`
	CreatedAt     string `json:"created_at"`
	Description   string `json:"description,omitempty"`
}

type Response struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      Transfer     `json:"data"`
	Metadata  ResponseMeta `json:"metadata,omitempty"`
	Timestamp string       `json:"timestamp,omitempty"`
}

type CreateInput struct {
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Chain       string `json:"chain"`
	Reference   string `json:"reference,omitempty"`
	Description string `json:"description,omitempty"`
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{configStore: configStore, authStore: authStore, client: client}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Response, error) {
	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return Response{}, err
	}

	active, err := profile.Active(cfg)
	if err != nil {
		return Response{}, err
	}

	credentials, err := s.authStore.LoadCredentials(ctx, active.Name)
	if err != nil {
		return Response{}, auth.CredentialLoadError(active.Name, err)
	}

	body, err := json.Marshal(input)
	if err != nil {
		return Response{}, err
	}

	responseBody, err := s.client.Do(ctx, "POST", "/api/wallets/transfers", credentials.ClientID, credentials.SecretKey, body)
	if err != nil {
		return Response{}, err
	}

	var response Response
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return Response{}, fmt.Errorf("decode transfer response: %w", err)
	}

	return response, nil
}
