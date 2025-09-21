package authz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client for making HTTP calls to auth service
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ValidateTokenRequest for internal API
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateTokenResponse from internal API
type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Error  string `json:"error,omitempty"`
}

// UserInfo for user profile requests
type UserInfo struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Email             string `json:"email,omitempty"`
	Phone             string `json:"phone,omitempty"`
	BankAccountName   string `json:"bank_account_name,omitempty"`
	BankAccountHolder string `json:"bank_account_holder,omitempty"`
}

// ValidateTokenHTTP - Alternative to local validation, uses HTTP call
func (c *AuthClient) ValidateTokenHTTP(token string) (*ValidateTokenResponse, error) {
	req := ValidateTokenRequest{Token: token}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/internal/validate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result ValidateTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetUserInfo - Get user profile from auth service
func (c *AuthClient) GetUserInfo(userID string) (*UserInfo, error) {
	resp, err := c.httpClient.Get(
		fmt.Sprintf("%s/internal/user/%s", c.baseURL, userID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user not found")
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &userInfo, nil
}

type UpdateUserAuthRequest struct {
	Type  string `json:"type"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// About link phone/email
// Let's use single endpoint, single handler (with
// Need your help to edit the handler/internal and service/internal, I can do the main, repo, and query level.

func (c *AuthClient) UpdateUserAuth(userAuthID, phone, email string) error {
	req := UpdateUserAuthRequest{
		Phone: phone,
		Email: email,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("PUT",
		fmt.Sprintf("%s/internal/userauth/%s", c.baseURL, userAuthID),
		bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update user auth, status: %d", resp.StatusCode)
	}

	return nil
}
