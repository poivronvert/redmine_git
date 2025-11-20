package platform

import (
	"fmt"
	"strconv"

	"colosscious.com/github-sync/internal/gitlab"
)

// GitLabAdapter 將 GitLab 客戶端適配為 IssueTracker 介面
type GitLabAdapter struct {
	client *gitlab.Client
}

// NewGitLabAdapter 創建 GitLab 適配器
func NewGitLabAdapter(client *gitlab.Client) *GitLabAdapter {
	return &GitLabAdapter{
		client: client,
	}
}

// CreateIssue 創建 Issue
func (a *GitLabAdapter) CreateIssue(project string, req IssueRequest) (*Issue, error) {
	gitlabReq := gitlab.CreateIssueRequest{
		Title:       req.Title,
		Description: req.Description,
		Labels:      req.Labels,
	}

	issue, err := a.client.CreateIssue(project, gitlabReq)
	if err != nil {
		return nil, err
	}

	return &Issue{
		ID:    strconv.Itoa(issue.IID),
		Title: issue.Title,
		URL:   issue.WebURL,
		State: issue.State,
	}, nil
}

// UpdateIssue 更新 Issue
func (a *GitLabAdapter) UpdateIssue(project string, issueID string, req IssueRequest) error {
	issueIID, err := strconv.Atoi(issueID)
	if err != nil {
		return fmt.Errorf("invalid GitLab issue IID: %s", issueID)
	}

	gitlabReq := gitlab.UpdateIssueRequest{
		Title:       req.Title,
		Description: req.Description,
		Labels:      req.Labels,
	}

	// 如果指定了狀態，設置 state_event
	if req.State == "closed" {
		gitlabReq.StateEvent = "close"
	} else if req.State == "open" || req.State == "opened" {
		gitlabReq.StateEvent = "reopen"
	}

	return a.client.UpdateIssue(project, issueIID, gitlabReq)
}

// CloseIssue 關閉 Issue
func (a *GitLabAdapter) CloseIssue(project string, issueID string) error {
	issueIID, err := strconv.Atoi(issueID)
	if err != nil {
		return fmt.Errorf("invalid GitLab issue IID: %s", issueID)
	}

	return a.client.CloseIssue(project, issueIID)
}

// AddComment 添加評論
func (a *GitLabAdapter) AddComment(project string, issueID string, comment string) error {
	issueIID, err := strconv.Atoi(issueID)
	if err != nil {
		return fmt.Errorf("invalid GitLab issue IID: %s", issueID)
	}

	return a.client.AddComment(project, issueIID, comment)
}

// ValidateProject 驗證專案
func (a *GitLabAdapter) ValidateProject(project string) error {
	return a.client.ValidateProject(project)
}

// BuildIssueURL 構建 Issue URL
func (a *GitLabAdapter) BuildIssueURL(project string, issueID string) string {
	issueIID, err := strconv.Atoi(issueID)
	if err != nil {
		return ""
	}

	return a.client.BuildIssueURL(project, issueIID)
}

// GetPlatformName 取得平台名稱
func (a *GitLabAdapter) GetPlatformName() string {
	return "gitlab"
}
