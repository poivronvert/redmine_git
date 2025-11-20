package gitlab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"colosscious.com/github-sync/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIssue(t *testing.T) {
	// Mock GitLab API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 驗證請求
		assert.Equal(t, "/api/v4/projects/123/issues", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "test-token", r.Header.Get("PRIVATE-TOKEN"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// 驗證 payload
		var req CreateIssueRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "Test Issue", req.Title)
		assert.Equal(t, "Test description", req.Description)
		assert.Equal(t, []string{"bug", "from-redmine"}, req.Labels)

		// 返回 mock issue
		response := Issue{
			IID:         456,
			Title:       req.Title,
			Description: req.Description,
			State:       "opened",
			WebURL:      "https://gitlab.com/namespace/project/-/issues/456",
			Labels:      req.Labels,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		apiVersion: "v4",
		client:     &http.Client{},
	}

	// 測試 CreateIssue
	req := CreateIssueRequest{
		Title:       "Test Issue",
		Description: "Test description",
		Labels:      []string{"bug", "from-redmine"},
	}

	issue, err := client.CreateIssue("123", req)
	require.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, 456, issue.IID)
	assert.Equal(t, "Test Issue", issue.Title)
	assert.Equal(t, "https://gitlab.com/namespace/project/-/issues/456", issue.WebURL)
}

func TestCreateIssueWithNamespaceProject(t *testing.T) {
	// 測試使用 namespace/project 格式
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// HTTP server 會自動解碼 URL，所以這裡收到的是解碼後的路徑
		// 但實際 GitLab API 會接收 URL encoded 的 project ID
		assert.Contains(t, r.URL.Path, "/api/v4/projects/")
		assert.Contains(t, r.URL.Path, "/issues")
		assert.Equal(t, "POST", r.Method)

		response := Issue{
			IID:    789,
			Title:  "Test",
			WebURL: "https://gitlab.com/namespace/project/-/issues/789",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		apiVersion: "v4",
		client:     &http.Client{},
	}

	req := CreateIssueRequest{
		Title: "Test",
	}

	issue, err := client.CreateIssue("namespace/project", req)
	require.NoError(t, err)
	assert.Equal(t, 789, issue.IID)
}

func TestCreateIssueError(t *testing.T) {
	// Mock server 返回錯誤
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"401 Unauthorized"}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "invalid-token",
		baseURL:    server.URL,
		apiVersion: "v4",
		client:     &http.Client{},
	}

	req := CreateIssueRequest{
		Title: "Test",
	}

	_, err := client.CreateIssue("123", req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestUpdateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/123/issues/456", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var req UpdateIssueRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "Updated Title", req.Title)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"iid":456,"title":"Updated Title"}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		apiVersion: "v4",
		client:     &http.Client{},
	}

	req := UpdateIssueRequest{
		Title: "Updated Title",
	}

	err := client.UpdateIssue("123", 456, req)
	assert.NoError(t, err)
}

func TestCloseIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/123/issues/456", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var req UpdateIssueRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "close", req.StateEvent)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"iid":456,"state":"closed"}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		apiVersion: "v4",
		client:     &http.Client{},
	}

	err := client.CloseIssue("123", 456)
	assert.NoError(t, err)
}

func TestAddComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/123/issues/456/notes", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req CreateNoteRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "Test comment", req.Body)

		response := Note{
			ID:   789,
			Body: req.Body,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		apiVersion: "v4",
		client:     &http.Client{},
	}

	err := client.AddComment("123", 456, "Test comment")
	assert.NoError(t, err)
}

func TestValidateProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := Project{
			ID:                123,
			Name:              "Test Project",
			PathWithNamespace: "namespace/project",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		apiVersion: "v4",
		client:     &http.Client{},
	}

	err := client.ValidateProject("123")
	assert.NoError(t, err)
}

func TestValidateProjectNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"404 Project Not Found"}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		apiVersion: "v4",
		client:     &http.Client{},
	}

	err := client.ValidateProject("999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or no permission")
}

func TestBuildIssueURL(t *testing.T) {
	client := &Client{
		baseURL:    "https://gitlab.com",
		apiVersion: "v4",
	}

	tests := []struct {
		name      string
		projectID string
		issueIID  int
		wantURL   string
	}{
		{
			name:      "namespace/project format",
			projectID: "mygroup/myproject",
			issueIID:  123,
			wantURL:   "https://gitlab.com/mygroup/myproject/-/issues/123",
		},
		{
			name:      "numeric ID",
			projectID: "456",
			issueIID:  789,
			wantURL:   "https://gitlab.com/projects/456/-/issues/789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := client.BuildIssueURL(tt.projectID, tt.issueIID)
			assert.Equal(t, tt.wantURL, url)
		})
	}
}

func TestNormalizeProjectID(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name      string
		projectID string
		want      string
	}{
		{
			name:      "numeric ID",
			projectID: "123",
			want:      "123",
		},
		{
			name:      "namespace/project",
			projectID: "mygroup/myproject",
			want:      "mygroup%2Fmyproject",
		},
		{
			name:      "nested namespace",
			projectID: "group/subgroup/project",
			want:      "group%2Fsubgroup%2Fproject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.normalizeProjectID(tt.projectID)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestNewClient(t *testing.T) {
	cfg := config.GitLabConfig{
		Token:      "test-token",
		BaseURL:    "https://gitlab.example.com",
		APIVersion: "v4",
	}

	client := NewClient(cfg)

	assert.Equal(t, "test-token", client.token)
	assert.Equal(t, "https://gitlab.example.com", client.baseURL)
	assert.Equal(t, "v4", client.apiVersion)
	assert.NotNil(t, client.client)
}

func TestNewClientTrimsTrailingSlash(t *testing.T) {
	cfg := config.GitLabConfig{
		Token:      "test-token",
		BaseURL:    "https://gitlab.example.com/",
		APIVersion: "v4",
	}

	client := NewClient(cfg)

	assert.Equal(t, "https://gitlab.example.com", client.baseURL)
}
