package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"colosscious.com/github-sync/internal/config"
)

// Client GitLab API 客戶端
type Client struct {
	token      string
	baseURL    string
	apiVersion string
	client     *http.Client
}

// Issue GitLab issue 結構
type Issue struct {
	IID         int    `json:"iid"`           // Internal ID (用於 API 操作)
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	WebURL      string `json:"web_url"`
	Labels      []string `json:"labels,omitempty"`
}

// CreateIssueRequest 創建 issue 的請求
type CreateIssueRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}

// UpdateIssueRequest 更新 issue 的請求
type UpdateIssueRequest struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	StateEvent  string   `json:"state_event,omitempty"` // "close" or "reopen"
}

// Note GitLab note (comment) 結構
type Note struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// CreateNoteRequest 創建 note 的請求
type CreateNoteRequest struct {
	Body string `json:"body"`
}

// Project GitLab project 結構（用於驗證）
type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

// NewClient 建立 GitLab 客戶端
func NewClient(cfg config.GitLabConfig) *Client {
	return &Client{
		token:      cfg.Token,
		baseURL:    strings.TrimSuffix(cfg.BaseURL, "/"),
		apiVersion: cfg.APIVersion,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// normalizeProjectID 規範化 Project ID
// 支援兩種格式：
// 1. 數字 ID: "12345"
// 2. Namespace/Project: "mygroup/myproject" (會進行 URL encode)
func (c *Client) normalizeProjectID(projectID string) string {
	// 如果是純數字，直接返回
	if _, err := strconv.Atoi(projectID); err == nil {
		return projectID
	}

	// 如果是 namespace/project 格式，進行 URL encode
	return url.PathEscape(projectID)
}

// buildAPIURL 構建 API URL
func (c *Client) buildAPIURL(path string) string {
	// path 應該以 / 開頭，例如 "/projects/123/issues"
	return fmt.Sprintf("%s/api/%s%s", c.baseURL, c.apiVersion, path)
}

// doRequest 執行 HTTP 請求的通用方法
func (c *Client) doRequest(method, url string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// 設置請求頭
	req.Header.Set("PRIVATE-TOKEN", c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// CreateIssue 在指定專案創建 issue
func (c *Client) CreateIssue(projectID string, req CreateIssueRequest) (*Issue, error) {
	normalizedID := c.normalizeProjectID(projectID)
	endpoint := c.buildAPIURL(fmt.Sprintf("/projects/%s/issues", normalizedID))

	respBody, statusCode, err := c.doRequest("POST", endpoint, req)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(respBody))
	}

	var issue Issue
	if err := json.Unmarshal(respBody, &issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

// UpdateIssue 更新指定的 issue
func (c *Client) UpdateIssue(projectID string, issueIID int, req UpdateIssueRequest) error {
	normalizedID := c.normalizeProjectID(projectID)
	endpoint := c.buildAPIURL(fmt.Sprintf("/projects/%s/issues/%d", normalizedID, issueIID))

	respBody, statusCode, err := c.doRequest("PUT", endpoint, req)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", statusCode, string(respBody))
	}

	return nil
}

// CloseIssue 關閉指定的 issue
func (c *Client) CloseIssue(projectID string, issueIID int) error {
	return c.UpdateIssue(projectID, issueIID, UpdateIssueRequest{
		StateEvent: "close",
	})
}

// AddComment 在指定 issue 添加評論 (使用 Notes API)
func (c *Client) AddComment(projectID string, issueIID int, comment string) error {
	normalizedID := c.normalizeProjectID(projectID)
	endpoint := c.buildAPIURL(fmt.Sprintf("/projects/%s/issues/%d/notes", normalizedID, issueIID))

	req := CreateNoteRequest{Body: comment}

	respBody, statusCode, err := c.doRequest("POST", endpoint, req)
	if err != nil {
		return err
	}

	if statusCode != http.StatusCreated {
		return fmt.Errorf("API returned status %d: %s", statusCode, string(respBody))
	}

	return nil
}

// ValidateProject 驗證專案是否存在且有權限存取
func (c *Client) ValidateProject(projectID string) error {
	normalizedID := c.normalizeProjectID(projectID)
	endpoint := c.buildAPIURL(fmt.Sprintf("/projects/%s", normalizedID))

	respBody, statusCode, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	if statusCode == http.StatusNotFound {
		return fmt.Errorf("project '%s' not found or no permission", projectID)
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", statusCode, string(respBody))
	}

	return nil
}

// BuildIssueURL 構建 issue 的 Web URL
func (c *Client) BuildIssueURL(projectID string, issueIID int) string {
	// GitLab issue URL 格式：https://gitlab.com/namespace/project/-/issues/123
	// 如果 projectID 是數字，我們需要先獲取 project 資訊來構建正確的 URL
	// 為了簡化，這裡假設 projectID 是 namespace/project 格式
	// 如果是數字 ID，URL 仍然可用但不是最佳實踐

	normalizedID := c.normalizeProjectID(projectID)

	// 如果是 URL encoded 的 namespace/project，需要 decode
	decodedID, err := url.PathUnescape(normalizedID)
	if err != nil {
		decodedID = normalizedID
	}

	// 如果是純數字 ID，我們無法直接構建正確的 URL
	// 這種情況下返回一個通用格式，但建議使用 namespace/project
	if _, err := strconv.Atoi(projectID); err == nil {
		// 純數字 ID 的情況，返回一個提示性的 URL
		return fmt.Sprintf("%s/projects/%s/-/issues/%d", c.baseURL, projectID, issueIID)
	}

	// namespace/project 格式
	return fmt.Sprintf("%s/%s/-/issues/%d", c.baseURL, decodedID, issueIID)
}

// GetProject 獲取專案詳細資訊
func (c *Client) GetProject(projectID string) (*Project, error) {
	normalizedID := c.normalizeProjectID(projectID)
	endpoint := c.buildAPIURL(fmt.Sprintf("/projects/%s", normalizedID))

	respBody, statusCode, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", statusCode, string(respBody))
	}

	var project Project
	if err := json.Unmarshal(respBody, &project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}
