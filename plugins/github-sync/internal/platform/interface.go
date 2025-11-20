package platform

// IssueTracker 定義統一的 Issue 追蹤平台介面
// 支援 GitHub, GitLab 等不同的 Issue 追蹤系統
type IssueTracker interface {
	// CreateIssue 在指定專案中創建 Issue
	// project: GitHub 格式為 "owner/repo", GitLab 格式為 Project ID 或 "namespace/project"
	CreateIssue(project string, req IssueRequest) (*Issue, error)

	// UpdateIssue 更新指定的 Issue
	// project: 專案識別符
	// issueID: GitHub 為 number (字串格式), GitLab 為 iid (字串格式)
	UpdateIssue(project string, issueID string, req IssueRequest) error

	// CloseIssue 關閉指定的 Issue
	// project: 專案識別符
	// issueID: Issue 識別符
	CloseIssue(project string, issueID string) error

	// AddComment 在指定 Issue 添加評論
	// project: 專案識別符
	// issueID: Issue 識別符
	// comment: 評論內容
	AddComment(project string, issueID string, comment string) error

	// ValidateProject 驗證專案是否存在且有權限存取
	// project: 專案識別符
	ValidateProject(project string) error

	// BuildIssueURL 構建 Issue 的 Web URL
	// project: 專案識別符
	// issueID: Issue 識別符
	BuildIssueURL(project string, issueID string) string

	// GetPlatformName 取得平台名稱 (github, gitlab 等)
	GetPlatformName() string
}

// IssueRequest 創建或更新 Issue 的請求結構
type IssueRequest struct {
	Title       string   // Issue 標題
	Description string   // Issue 描述內容
	Labels      []string // 標籤列表
	State       string   // 狀態 (open, closed 等) - 更新時使用
}

// Issue 表示一個 Issue 的回傳資料
type Issue struct {
	ID      string // GitHub: number (轉字串), GitLab: iid (轉字串)
	Title   string // Issue 標題
	URL     string // Issue 的 Web URL
	State   string // Issue 狀態 (open, closed 等)
}
