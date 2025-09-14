package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/teammachinist/tutuplapak/internal"
)

type FileMetadataResponse struct {
	ID               uuid.UUID `json:"id"`
	UserID           string    `json:"user_id"`
	FileURI          string    `json:"file_uri"`
	FileThumbnailURI string    `json:"file_thumbnail_uri"`
	CreatedAt        time.Time `json:"created_at"`
}
type FileClientInterface interface {
	GetFileByID(ctx context.Context, fileID uuid.UUID, userID string) (*FileMetadataResponse, error) // ← tambahkan userID
}

type FileClient struct {
	BaseURL    string
	HTTPClient *http.Client
	JWTService *internal.JWTService
}

func NewFileClient(baseURL string, jwtService *internal.JWTService) *FileClient {
	return &FileClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		JWTService: jwtService,
	}
}

func (fc *FileClient) GetFileByID(ctx context.Context, fileID uuid.UUID, userID string) (*FileMetadataResponse, error) {

	// Generate JWT token untuk request — gunakan userID dari produk!
	token, err := fc.JWTService.GenerateToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/file/%s", fc.BaseURL, fileID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := fc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("file not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var fileResp FileMetadataResponse
	if err := json.Unmarshal(body, &fileResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &fileResp, nil
}
