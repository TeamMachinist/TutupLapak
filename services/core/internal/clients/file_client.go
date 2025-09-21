package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileMetadataResponse struct {
	ID               uuid.UUID `json:"id"`
	UserID           string    `json:"user_id"`
	FileURI          string    `json:"file_uri"`
	FileThumbnailURI string    `json:"file_thumbnail_uri"`
	CreatedAt        time.Time `json:"created_at"`
}

type FileClientInterface interface {
	GetFileByID(ctx context.Context, fileID uuid.UUID) (*FileMetadataResponse, error)
	GetFilesByIDList(ctx context.Context, fileIDs []string) ([]*FileMetadataResponse, error)
}

type FileClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewFileClient(baseURL string) *FileClient {
	return &FileClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (fc *FileClient) GetFileByID(ctx context.Context, fileID uuid.UUID) (*FileMetadataResponse, error) {
	url := fmt.Sprintf("%s/v1/file/%s", fc.BaseURL, fileID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Tidak ada Authorization header — tanpa auth
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

func (fc *FileClient) GetFilesByIDList(ctx context.Context, fileIDs []string) ([]*FileMetadataResponse, error) {
	if len(fileIDs) == 0 {
		return []*FileMetadataResponse{}, nil
	}

	query := url.Values{}
	joinId := strings.Join(fileIDs, ",")
	query.Set("id", joinId)

	fmt.Println(`print disini saja bos:`, joinId)

	baseURL := fmt.Sprintf("%s/v1/file", fc.BaseURL)
	reqURL := fmt.Sprintf("%s?%s", baseURL, query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var filesResp []*FileMetadataResponse
	if err := json.Unmarshal(body, &filesResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return filesResp, nil
}
