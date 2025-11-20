package platform

import (
	"fmt"
	"strconv"

	"colosscious.com/github-sync/internal/github"
)

// GitHubAdapter 將 GitHub 客戶端適配為 IssueTracker 介面
type GitHubAdapter struct {
	client  *github.Client
	baseURL string
}

// NewGitHubAdapter 創建 GitHub 適配器
func NewGitHubAdapter(client *github.Client, baseURL string) *GitHubAdapter {
	return &GitHubAdapter{
		client:  client,
		baseURL: baseURL,
	}
}

// CreateIssue 創建 Issue
func (a *GitHubAdapter) CreateIssue(project string, req IssueRequest) (*Issue, error) {
	githubReq := github.CreateIssueRequest{
		Title:  req.Title,
		Body:   req.Description,
		Labels: req.Labels,
	}

	issue, err := a.client.CreateIssue(project, githubReq)
	if err != nil {
		return nil, err
	}

	return &Issue{
		ID:    strconv.Itoa(issue.Number),
		Title: issue.Title,
		URL:   issue.HTMLURL,
		State: issue.State,
	}, nil
}

// UpdateIssue 更新 Issue
func (a *GitHubAdapter) UpdateIssue(project string, issueID string, req IssueRequest) error {
	issueNumber, err := strconv.Atoi(issueID)
	if err != nil {
		return fmt.Errorf("invalid GitHub issue number: %s", issueID)
	}

	githubReq := github.CreateIssueRequest{
		Title:  req.Title,
		Body:   req.Description,
		State:  req.State,
		Labels: req.Labels,
	}

	return a.client.UpdateIssue(project, issueNumber, githubReq)
}

// CloseIssue 關閉 Issue
func (a *GitHubAdapter) CloseIssue(project string, issueID string) error {
	issueNumber, err := strconv.Atoi(issueID)
	if err != nil {
		return fmt.Errorf("invalid GitHub issue number: %s", issueID)
	}

	return a.client.CloseIssue(project, issueNumber)
}

// AddComment 添加評論
func (a *GitHubAdapter) AddComment(project string, issueID string, comment string) error {
	issueNumber, err := strconv.Atoi(issueID)
	if err != nil {
		return fmt.Errorf("invalid GitHub issue number: %s", issueID)
	}

	return a.client.AddComment(project, issueNumber, comment)
}

// ValidateProject 驗證專案
func (a *GitHubAdapter) ValidateProject(project string) error {
	return a.client.ValidateRepo(project)
}

// BuildIssueURL 構建 Issue URL
func (a *GitHubAdapter) BuildIssueURL(project string, issueID string) string {
	issueNumber, err := strconv.Atoi(issueID)
	if err != nil {
		return ""
	}

	return github.BuildIssueURL(a.baseURL, project, issueNumber)
}

// GetPlatformName 取得平台名稱
func (a *GitHubAdapter) GetPlatformName() string {
	return "github"
}
