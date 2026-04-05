package api

import (
	"context"
	"encoding/json"
	"fmt"
)

type WhoAmIResponse struct {
	Authenticated   bool            `json:"authenticated"`
	AuthMethod      string          `json:"auth_method"`
	ClientID        string          `json:"client_id"`
	ClientName      string          `json:"client_name"`
	ActiveCompanyID string          `json:"active_company_id"`
	Environment     string          `json:"environment"`
	Permissions     []string        `json:"permissions"`
	Active          bool            `json:"active"`
	Metadata        map[string]any  `json:"metadata"`
	RateLimit       WhoAmIRateLimit `json:"rate_limit"`
	Timestamp       string          `json:"timestamp"`
}

type WhoAmIRateLimit struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerHour   int `json:"requests_per_hour"`
	BurstSize         int `json:"burst_size"`
}

func (c *Client) GetWhoAmI(ctx context.Context, clientID string, secretKey string) (WhoAmIResponse, error) {
	body, err := c.Do(ctx, "GET", "/api/whoami", clientID, secretKey, nil)
	if err != nil {
		return WhoAmIResponse{}, err
	}

	var response WhoAmIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return WhoAmIResponse{}, fmt.Errorf("decode whoami response: %w", err)
	}

	return response, nil
}
