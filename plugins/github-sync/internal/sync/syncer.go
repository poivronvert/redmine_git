package sync

import (
	"fmt"
	"log"
	"strings"

	"colosscious.com/github-sync/internal/config"
	"colosscious.com/github-sync/internal/github"
	"colosscious.com/github-sync/internal/gitlab"
	"colosscious.com/github-sync/internal/platform"
	"colosscious.com/github-sync/internal/redmine"
	"colosscious.com/github-sync/internal/storage"
)

// Syncer 同步器
type Syncer struct {
	config       *config.Config
	redmine      *redmine.Client
	githubClient *github.Client
	gitlabClient *gitlab.Client
	github       platform.IssueTracker
	gitlab       platform.IssueTracker
	storage      *storage.PostgresDB
}

// NewSyncer 建立同步器
func NewSyncer(cfg *config.Config, db *storage.PostgresDB) *Syncer {
	s := &Syncer{
		config:  cfg,
		redmine: redmine.NewClient(cfg.Redmine),
		storage: db,
	}

	// 初始化 GitHub 客戶端和適配器（如果有配置）
	if cfg.GitHub.Token != "" {
		s.githubClient = github.NewClient(cfg.GitHub)
		s.github = platform.NewGitHubAdapter(s.githubClient, cfg.GitHub.BaseURL)
	}

	// 初始化 GitLab 客戶端和適配器（如果有配置）
	if cfg.GitLab.Token != "" {
		s.gitlabClient = gitlab.NewClient(cfg.GitLab)
		s.gitlab = platform.NewGitLabAdapter(s.gitlabClient)
	}

	return s
}

// Run 執行一次同步
func (s *Syncer) Run() error {
	log.Println("Starting sync run...")

	totalSynced := 0
	totalErrors := 0

	// 遍歷所有配置的專案
	for _, project := range s.config.Redmine.Projects {
		synced, errors := s.syncProject(project)
		totalSynced += synced
		totalErrors += errors
	}

	log.Printf("Sync completed: %d issues synced, %d errors", totalSynced, totalErrors)

	// 印出統計資訊
	stats, err := s.storage.GetStats()
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
	} else {
		log.Printf("Stats: Total synced=%d, Today=%d, Unresolved errors=%d",
			stats["total_synced"], stats["today_synced"], stats["unresolved_errors"])
	}

	return nil
}

// syncProject 同步單一專案
func (s *Syncer) syncProject(project config.ProjectConfig) (int, int) {
	log.Printf("Syncing project: %s", project.Identifier)

	synced := 0
	errors := 0

	// 同步到 GitHub（如果啟用）
	if project.SyncTo.GitHub && s.github != nil {
		issues, err := s.redmine.GetNewIssues(
			project.Identifier,
			project.CustomFields.TargetRepoID,
			project.CustomFields.GitHubIssueURLID,
		)

		if err != nil {
			log.Printf("Failed to get GitHub issues for project %s: %v", project.Identifier, err)
			errors++
		} else if len(issues) > 0 {
			log.Printf("Found %d GitHub issues to sync", len(issues))
			for _, issue := range issues {
				if err := s.syncToGitHub(issue, project); err != nil {
					log.Printf("Failed to sync issue #%d to GitHub: %v", issue.ID, err)
					errors++
				} else {
					synced++
				}
			}
		}
	}

	// 同步到 GitLab（如果啟用）
	if project.SyncTo.GitLab && s.gitlab != nil {
		issues, err := s.redmine.GetNewIssues(
			project.Identifier,
			project.CustomFields.TargetGitLabProjectID,
			project.CustomFields.GitLabIssueURLID,
		)

		if err != nil {
			log.Printf("Failed to get GitLab issues for project %s: %v", project.Identifier, err)
			errors++
		} else if len(issues) > 0 {
			log.Printf("Found %d GitLab issues to sync", len(issues))
			for _, issue := range issues {
				if err := s.syncToGitLab(issue, project); err != nil {
					log.Printf("Failed to sync issue #%d to GitLab: %v", issue.ID, err)
					errors++
				} else {
					synced++
				}
			}
		}
	}

	if synced == 0 && errors == 0 {
		log.Printf("No new issues to sync for project %s", project.Identifier)
	}

	return synced, errors
}

// syncToGitHub 同步到 GitHub
func (s *Syncer) syncToGitHub(issue redmine.Issue, project config.ProjectConfig) error {
	// 檢查是否已同步
	isSynced, err := s.storage.IsSynced(issue.ID, "github")
	if err != nil {
		return fmt.Errorf("failed to check sync status: %w", err)
	}
	if isSynced {
		log.Printf("Issue #%d already synced to GitHub, skipping", issue.ID)
		return nil
	}

	// 取得目標 repo
	targetRepo := issue.GetCustomFieldValue(project.CustomFields.TargetRepoID)
	if targetRepo == "" {
		log.Printf("Issue #%d has no target GitHub repo, skipping", issue.ID)
		return nil
	}

	// 驗證 repo 格式
	if !strings.Contains(targetRepo, "/") {
		errMsg := fmt.Sprintf("Invalid repo format '%s', expected 'owner/repo'", targetRepo)
		s.handleError(issue.ID, errMsg, "github")
		return fmt.Errorf("invalid repo format: %s", targetRepo)
	}

	return s.syncToTracker(s.github, issue, project, targetRepo,
		project.CustomFields.GitHubIssueURLID, "github")
}

// syncToGitLab 同步到 GitLab
func (s *Syncer) syncToGitLab(issue redmine.Issue, project config.ProjectConfig) error {
	// 檢查是否已同步
	isSynced, err := s.storage.IsSynced(issue.ID, "gitlab")
	if err != nil {
		return fmt.Errorf("failed to check sync status: %w", err)
	}
	if isSynced {
		log.Printf("Issue #%d already synced to GitLab, skipping", issue.ID)
		return nil
	}

	// 取得目標 project
	targetProject := issue.GetCustomFieldValue(project.CustomFields.TargetGitLabProjectID)
	if targetProject == "" {
		log.Printf("Issue #%d has no target GitLab project, skipping", issue.ID)
		return nil
	}

	return s.syncToTracker(s.gitlab, issue, project, targetProject,
		project.CustomFields.GitLabIssueURLID, "gitlab")
}

// syncToTracker 同步到指定的 Issue Tracker 平台
func (s *Syncer) syncToTracker(
	tracker platform.IssueTracker,
	issue redmine.Issue,
	project config.ProjectConfig,
	targetProject string,
	urlFieldID int,
	platformName string,
) error {
	log.Printf("Syncing issue #%d to %s project: %s", issue.ID, platformName, targetProject)

	// 建立 issue title
	title := fmt.Sprintf(s.config.Sync.TitleFormat, issue.ID, issue.Subject)

	// 準備 issue body
	body := s.buildIssueBody(issue)

	// 創建 issue
	createdIssue, err := tracker.CreateIssue(targetProject, platform.IssueRequest{
		Title:       title,
		Description: body,
		Labels:      s.mapLabels(issue),
	})

	if err != nil {
		s.handleError(issue.ID, fmt.Sprintf("Failed to create %s issue: %v", platformName, err), platformName)
		return fmt.Errorf("failed to create %s issue: %w", platformName, err)
	}

	log.Printf("Created %s issue: %s", platformName, createdIssue.URL)

	// 回寫 URL 到 Redmine
	if err := s.redmine.UpdateCustomField(
		issue.ID,
		urlFieldID,
		createdIssue.URL,
	); err != nil {
		log.Printf("Warning: Failed to update Redmine custom field: %v", err)
	}

	// 記錄到資料庫
	if err := s.storage.RecordSyncWithPlatform(storage.SyncRecord{
		RedmineIssueID:    issue.ID,
		RedmineProject:    project.Identifier,
		GitHubRepo:        targetProject,        // 通用欄位，存放目標專案識別符
		GitHubIssueNumber: 0,                   // 將在 storage 層處理
		GitHubIssueURL:    createdIssue.URL,
		Platform:          platformName,
	}, createdIssue.ID); err != nil {
		return fmt.Errorf("failed to record sync: %w", err)
	}

	log.Printf("✓ Successfully synced Redmine #%d -> %s %s#%s",
		issue.ID, platformName, targetProject, createdIssue.ID)

	return nil
}

// buildIssueBody 建立 issue 的 body
func (s *Syncer) buildIssueBody(issue redmine.Issue) string {
	body := fmt.Sprintf("**From Redmine Issue #%d**\n\n", issue.ID)
	body += fmt.Sprintf("**Project**: %s\n", issue.Project.Name)
	body += fmt.Sprintf("**Tracker**: %s\n", issue.Tracker.Name)
	body += fmt.Sprintf("**Priority**: %s\n", issue.Priority.Name)
	body += fmt.Sprintf("**Author**: %s\n", issue.Author.Name)
	body += fmt.Sprintf("**Created**: %s\n\n", issue.CreatedOn)
	body += "---\n\n"

	if issue.Description != "" {
		body += issue.Description
	} else {
		body += "*No description*"
	}

	body += fmt.Sprintf("\n\n---\n*Synced from Redmine: %s/issues/%d*",
		s.config.Redmine.GetDisplayURL(), issue.ID)

	return body
}

// mapLabels 將 Redmine 的 tracker/priority 對應到 labels
func (s *Syncer) mapLabels(issue redmine.Issue) []string {
	var labels []string

	// Tracker 對應
	switch issue.Tracker.Name {
	case "Bug":
		labels = append(labels, "bug")
	case "Feature":
		labels = append(labels, "enhancement")
	case "Support":
		labels = append(labels, "question")
	}

	// Priority 對應
	switch issue.Priority.Name {
	case "Urgent", "Immediate":
		labels = append(labels, "priority:high")
	case "High":
		labels = append(labels, "priority:medium")
	}

	// 加上來源標籤
	labels = append(labels, "from-redmine")

	return labels
}

// handleError 處理同步錯誤
func (s *Syncer) handleError(issueID int, errorMsg string, platformName string) {
	// 1. 記錄到 log
	if s.config.Sync.OnError.Log {
		log.Printf("Error syncing issue #%d to %s: %s", issueID, platformName, errorMsg)
	}

	// 2. 記錄到資料庫
	if err := s.storage.RecordError(issueID, errorMsg); err != nil {
		log.Printf("Failed to record error to DB: %v", err)
	}

	// 3. 在 Redmine 加註解
	if s.config.Sync.OnError.AddRedmineNote {
		note := fmt.Sprintf("⚠️ %s 同步失敗\n\n錯誤訊息：%s", strings.Title(platformName), errorMsg)
		if err := s.redmine.AddNote(issueID, note); err != nil {
			log.Printf("Failed to add Redmine note: %v", err)
		}
	}
}

// UpdateConfig 更新配置（用於熱更新）
func (s *Syncer) UpdateConfig(cfg *config.Config) {
	s.config = cfg
	s.redmine = redmine.NewClient(cfg.Redmine)

	// 更新 GitHub 客戶端
	if cfg.GitHub.Token != "" {
		s.githubClient = github.NewClient(cfg.GitHub)
		s.github = platform.NewGitHubAdapter(s.githubClient, cfg.GitHub.BaseURL)
	} else {
		s.githubClient = nil
		s.github = nil
	}

	// 更新 GitLab 客戶端
	if cfg.GitLab.Token != "" {
		s.gitlabClient = gitlab.NewClient(cfg.GitLab)
		s.gitlab = platform.NewGitLabAdapter(s.gitlabClient)
	} else {
		s.gitlabClient = nil
		s.gitlab = nil
	}

	log.Println("Syncer config updated")
}
