# Redmine with GitFetcher & GitHub Sync

整合 Redmine 專案管理系統、GitFetcher 自動拉取服務和 GitHub Issue 同步服務。

## 服務架構

```
┌─────────────┐         ┌──────────────┐         ┌─────────┐
│   Redmine   │◄───────►│  GitFetcher  │◄────────│  GitHub │
│   (13000)   │         │   (13100)    │   pull  │   Repos │
└─────────────┘         └──────────────┘         └─────────┘
       │                                                 ▲
       │                ┌──────────────┐                │
       └───────────────►│ GitHub Sync  │────────────────┘
                        │  (background)│   create issue
                        └──────────────┘
                               │
                        ┌──────────────┐
                        │  PostgreSQL  │
                        │  (5432)      │
                        └──────────────┘
```

## 服務說明

### Redmine (Port: 13000)
專案管理系統，用於 issue tracking 和專案協作。

### GitFetcher (Port: 13100)
自動從 GitHub 拉取 Git repositories 並同步到 Redmine。
- 定時拉取 Git repos
- 提供 Web UI 管理介面
- RESTful API
- 詳細文檔：[plugins/gitfetcher/README.md](plugins/gitfetcher/README.md)

### GitHub/GitLab Sync (背景服務)
自動將 Redmine issues 同步到 GitHub 或 GitLab。
- 單向同步：Redmine → GitHub/GitLab
- 支援多專案、多平台
- 可同時同步到 GitHub 和 GitLab
- 配置檔熱更新
- 詳細文檔：[plugins/github-sync/README.md](plugins/github-sync/README.md)

### PostgreSQL (Port: 5432)
共用資料庫，包含三個獨立 schemas：
- `public` - Redmine 資料
- `gitfetcher` - GitFetcher 資料
- `redmine_github_sync` - GitHub Sync 資料

## 快速開始

### 1. 前置需求

- Docker
- Docker Compose
- SSH keys（用於 Git 存取）
- GitHub Personal Access Token（用於 GitHub Sync）

### 2. 配置服務

#### 2.1 GitFetcher 配置

```bash
cp plugins/gitfetcher/config.example.yaml gitfetcher-config.yaml
```

編輯 `gitfetcher-config.yaml`：

```yaml
repos:
  - name: "my-repo"
    url: "git@github.com:myorg/my-repo.git"
    local_path: "/repos/my-repo"
    interval: "5m"

ssh_key_path: "/tmp/.ssh/id_ed25519"
http_port: 8080
log_path: "./logs"
```

#### 2.2 GitHub/GitLab Sync 配置

```bash
cp github-sync-config.example.yaml github-sync-config.yaml
```

編輯 `github-sync-config.yaml`：

**僅同步到 GitHub（預設）：**
```yaml
redmine:
  url: "http://redmine:3000"
  display_url: "http://192.168.181.245:13000"
  api_key: "your-redmine-api-key"
  projects:
    - identifier: "my-project"
      custom_fields:
        target_repo_id: 11
        github_issue_url_id: 12
      sync_to:
        github: true
        gitlab: false

github:
  token: "ghp_xxxxxxxxxxxx"
  base_url: "https://github.com"
```

**同時同步到 GitHub 和 GitLab：**
```yaml
redmine:
  projects:
    - identifier: "my-project"
      custom_fields:
        target_repo_id: 11
        github_issue_url_id: 12
        target_gitlab_project_id: 13
        gitlab_issue_url_id: 14
      sync_to:
        github: true
        gitlab: true

github:
  token: "ghp_xxxxxxxxxxxx"
  base_url: "https://github.com"

gitlab:
  token: "glpat-xxxxxxxxxxxx"
  base_url: "https://gitlab.com"
  api_version: "v4"
```

**僅同步到 GitLab：**
```yaml
redmine:
  projects:
    - identifier: "my-project"
      custom_fields:
        target_gitlab_project_id: 13
        gitlab_issue_url_id: 14
      sync_to:
        github: false
        gitlab: true

gitlab:
  token: "glpat-xxxxxxxxxxxx"
  base_url: "https://gitlab.com"
```

詳細說明請參考 `github-sync-config.example.yaml` 中的註解。

#### 2.3 配置 SSH Keys

確保 `~/.ssh/id_ed25519` 存在且有權限訪問 Git repositories。

### 3. 啟動服務

```bash
# 啟動所有服務
docker compose up -d

# 查看 logs
docker compose logs -f

# 單獨啟動特定服務
docker compose up -d redmine
docker compose up -d gitfetcher
docker compose up -d github-sync
```

### 4. 訪問服務

- **Redmine**: http://your-server-ip:13000
  - 預設帳號：`admin`
  - 預設密碼：`admin`（首次登入後請修改）

- **GitFetcher Web UI**: http://your-server-ip:13100

## 完整工作流程

### 1. Redmine 設定

#### 建立 Custom Fields
進入 Redmine 管理介面 > 自訂欄位 > 議題：

**GitHub 整合（如需要）：**
1. **目標 GitHub Repo** (List)
   - 選項值：`myorg/backend`, `myorg/frontend` 等

2. **GitHub Issue URL** (Link)
   - 系統自動填入，用於顯示同步結果

**GitLab 整合（如需要）：**
1. **Target GitLab Project** (Text)
   - 填入 GitLab Project ID（推薦）或 `namespace/project` 格式
   - 如何取得 Project ID：進入 GitLab 專案 → 點擊右上角 ⋮ → "Copy project ID"

2. **GitLab Issue URL** (Link)
   - 系統自動填入，用於顯示同步結果

#### 設定 Repository
1. 進入專案設定 > Repositories
2. 新增 Repository：
   - **SCM**: Git
   - **Path**: `/usr/src/redmine/repositories/my-repo`
   - **Identifier**: `my-repo`

### 2. 使用流程

1. **GitFetcher 拉取程式碼**
   - 自動從 GitHub 拉取最新 commits
   - 同步到 Redmine repositories 目錄

2. **Redmine 更新 Changesets**
   - 設定 Cron job 定時執行：
   ```bash
   */10 * * * * curl -s "http://192.168.181.245:13000/sys/fetch_changesets?key=YOUR_API_KEY" > /dev/null 2>&1
   ```

3. **建立 Redmine Issue**
   - 新增 issue
   - 選擇「目標 GitHub Repo」
   - 儲存

4. **自動同步到 GitHub**
   - GitHub Sync 每 5 分鐘檢查一次
   - 自動建立 GitHub issue
   - 回寫 URL 到 Redmine

## 常用操作

### 查看 Logs

```bash
# 所有服務
docker compose logs -f

# 特定服務
docker compose logs -f redmine
docker compose logs -f gitfetcher
docker compose logs -f github-sync
docker compose logs -f postgres_db
```

### 重啟服務

```bash
# 重啟所有服務
docker compose restart

# 重啟特定服務
docker compose restart github-sync
```

### 停止服務

```bash
# 停止所有服務
docker compose down

# 停止並刪除 volumes（注意：會刪除所有資料）
docker compose down -v
```

### 重新編譯

```bash
# 重新編譯並啟動
docker compose up -d --build

# 重新編譯特定服務
docker compose build github-sync
docker compose up -d github-sync
```

## 資料持久化

Docker volumes：
- `postgres-data`: PostgreSQL 資料庫資料
- `redmine-repositories`: Redmine repositories（與 GitFetcher 共用）
- `redmine-files`: Redmine 附件檔案

這些 volumes 會持久化資料，即使容器被刪除也不會遺失。

## 故障排除

### GitFetcher 無法拉取 repository

1. **檢查 SSH key 權限**：
   ```bash
   docker exec -it super_redmine_gitfetcher ssh -T git@github.com
   ```
   預期輸出：`Hi username! You've successfully authenticated`

2. **檢查 repository 路徑**：
   ```bash
   docker exec -it super_redmine_gitfetcher ls -la /repos/
   ```

3. **查看詳細 logs**：
   ```bash
   docker compose logs gitfetcher
   cat plugins/gitfetcher/logs/fetch-$(date +%Y-%m-%d).log
   ```

### Redmine 看不到 Git commits

1. **檢查 repository 路徑**：
   - Redmine 設定路徑：`/usr/src/redmine/repositories/my-repo`
   - 實際掛載目錄：`ls /usr/src/redmine/repositories/`

2. **手動觸發 fetch_changesets**：
   ```bash
   curl "http://localhost:13000/sys/fetch_changesets?key=YOUR_API_KEY"
   ```

3. **檢查 Redmine logs**：
   ```bash
   docker compose logs redmine
   ```

### GitHub Sync 同步失敗

1. **檢查服務狀態**：
   ```bash
   docker compose ps github-sync
   docker compose logs -f github-sync
   ```

2. **檢查資料庫連線**：
   ```bash
   docker compose exec github-sync ping postgres_db
   ```

3. **查看資料庫狀態**：
   ```bash
   docker compose exec postgres_db psql -U redmine -d redmine
   # 在 psql 內執行：
   SELECT * FROM redmine_github_sync.sync_records ORDER BY synced_at DESC LIMIT 10;
   SELECT * FROM redmine_github_sync.sync_errors WHERE resolved = FALSE;
   ```

4. **檢查配置**：
   - Redmine API key 是否正確
   - GitHub token 是否有 `repo` 權限
   - Custom field IDs 是否正確

### 配置熱更新失效

某些編輯器（如 vim）會改變檔案 inode，導致 fsnotify 失效。

解決方法：重啟服務
```bash
docker compose restart github-sync
```

## 專案結構

```
redmine_git/
├── docker-compose.yml              # Docker Compose 配置
├── Dockerfile                      # Redmine Dockerfile
├── gitfetcher-config.yaml          # GitFetcher 配置
├── github-sync-config.yaml         # GitHub Sync 配置
├── plugins/
│   ├── gitfetcher/                 # GitFetcher 服務
│   │   ├── Dockerfile
│   │   ├── README.md
│   │   ├── main.go
│   │   └── ...
│   └── github-sync/                # GitHub Sync 服務
│       ├── Dockerfile
│       ├── README.md
│       ├── cmd/sync/main.go
│       ├── internal/
│       └── ...
└── repos/                          # Git repositories（共用掛載）
```

## 端口配置

- Redmine: 13000 (外部) → 3000 (內部)
- GitFetcher: 13100 (外部) → 8080 (內部)
- PostgreSQL: 5432 (僅內部網路)

## 安全建議

1. **修改預設密碼**：首次登入 Redmine 後立即修改 admin 密碼
2. **保護 API Key**：妥善保管 Redmine API key 和 GitHub token
3. **SSH Key 權限**：使用 deploy keys 或限制 SSH keys 權限
4. **網路隔離**：考慮在 Docker 內部網路運行，透過 reverse proxy 對外
5. **定期備份**：定期備份 Docker volumes

## 授權

本專案基於以下開源專案：
- [Redmine](https://www.redmine.org/) - GPL v2
- [PostgreSQL](https://www.postgresql.org/) - PostgreSQL License
- GitFetcher - 自主開發
- GitHub Sync - 自主開發

## 維護者

CLCS Admin

## 更新記錄

- **2025-11-13**: 新增 GitHub Sync 服務與 display_url 配置
- **2025-11-11**: 初始版本，包含 Redmine 與 GitFetcher
