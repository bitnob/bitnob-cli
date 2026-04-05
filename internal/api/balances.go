package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type BalancesResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      BalancesData `json:"data"`
	Metadata  Metadata     `json:"metadata"`
	Timestamp string       `json:"timestamp"`
}

type BalancesData struct {
	CompanyID string    `json:"company_id"`
	Accounts  []Account `json:"accounts"`
}

type Account struct {
	AccountID                 string `json:"account_id"`
	AccountNumber             string `json:"account_number"`
	Currency                  string `json:"currency"`
	LedgerBalance             string `json:"ledger_balance"`
	AvailableBalance          string `json:"available_balance"`
	LedgerBalanceFormatted    string `json:"ledger_balance_formatted"`
	AvailableBalanceFormatted string `json:"available_balance_formatted"`
	CreatedAt                 string `json:"created_at"`
}

type Metadata struct {
	RequestID string `json:"request_id"`
}

type CurrencyBalanceResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	Data      CurrencyBalanceData `json:"data"`
	Metadata  Metadata            `json:"metadata"`
	Timestamp string              `json:"timestamp"`
}

type CurrencyBalanceData struct {
	AccountNumber             string `json:"account_number"`
	Currency                  string `json:"currency"`
	LedgerBalance             string `json:"ledger_balance"`
	AvailableBalance          string `json:"available_balance"`
	LedgerBalanceFormatted    string `json:"ledger_balance_formatted"`
	AvailableBalanceFormatted string `json:"available_balance_formatted"`
	AsOf                      string `json:"as_of"`
	CompanyID                 string `json:"company_id"`
}

func (c *Client) GetBalances(ctx context.Context, clientID string, secretKey string) (BalancesResponse, error) {
	body, err := c.Do(ctx, "GET", "/api/balances", clientID, secretKey, nil)
	if err != nil {
		return BalancesResponse{}, err
	}

	var response BalancesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return BalancesResponse{}, fmt.Errorf("decode balances response: %w", err)
	}

	return response, nil
}

func (c *Client) VerifyBalances(ctx context.Context, clientID string, secretKey string) error {
	_, err := c.Do(ctx, "GET", "/api/balances", clientID, secretKey, nil)
	return err
}

func (c *Client) GetBalanceByCurrency(ctx context.Context, clientID string, secretKey string, currency string) (CurrencyBalanceResponse, error) {
	body, err := c.Do(ctx, "GET", "/api/balances/"+strings.ToUpper(strings.TrimSpace(currency)), clientID, secretKey, nil)
	if err != nil {
		return CurrencyBalanceResponse{}, err
	}

	var response CurrencyBalanceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return CurrencyBalanceResponse{}, fmt.Errorf("decode currency balance response: %w", err)
	}

	return response, nil
}

func (c *Client) GetWallets(ctx context.Context, clientID string, secretKey string) (BalancesResponse, error) {
	body, err := c.Do(ctx, "GET", "/api/wallets", clientID, secretKey, nil)
	if err != nil {
		return BalancesResponse{}, err
	}

	var response BalancesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return BalancesResponse{}, fmt.Errorf("decode wallets response: %w", err)
	}

	return response, nil
}
