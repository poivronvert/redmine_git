# CLAUDE.md - AI Assistant Guide for Redmine Git Integration

## Overview

This repository integrates Redmine project management with Git repository synchronization and GitHub issue management through a microservices architecture. It consists of three main services that work together to provide automated Git synchronization and bidirectional issue management between Redmine and GitHub.

**Project Purpose**: Provide automated Git repository synchronization and GitHub issue integration for Redmine installations.

**Primary Language**: Go (services), Ruby on Rails (Redmine)

**Deployment**: Docker Compose with shared PostgreSQL database

## Architecture

### Service Overview

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

### Components

1. **Redmine** (Port: 13000)
   - Project management and issue tracking
   - Built on official Redmine Docker image
   - Custom plugins: Mermaid Macro
   - Shares repository volume with GitFetcher

2. **GitFetcher** (Port: 13100)
   - Go microservice for automated Git repository fetching
   - Web UI for status monitoring and configuration
   - Hot-reload configuration support
   - RESTful API for manual triggers

3. **GitHub Sync** (Background service)
   - One-way synchronization: Redmine → GitHub
   - Polls Redmine for new issues
   - Creates corresponding GitHub issues
   - Uses dedicated PostgreSQL schema for tracking

4. **PostgreSQL** (Port: 5432, internal only)
   - Shared database with isolated schemas:
     - `public` - Redmine data
     - `gitfetcher` - GitFetcher data (if needed)
     - `redmine_github_sync` - GitHub Sync tracking

## Directory Structure

```
redmine_git/
├── .claude/                        # Claude Code configuration
├── .git/                          # Git repository
├── .gitignore                     # Git ignore patterns
├── Dockerfile                     # Redmine container build
├── docker-compose.yml             # Multi-service orchestration
├── gitfetcher-config.yaml         # GitFetcher configuration
├── github-sync-config.example.yaml # GitHub Sync config template
├── README.md                      # Project documentation (Chinese)
├── CLAUDE.md                      # This file - AI assistant guide
│
├── plugins/
│   ├── gitfetcher/               # Git repository fetcher service
│   │   ├── config/               # Configuration management
│   │   │   ├── config.go
│   │   │   └── config_test.go
│   │   ├── fetcher/              # Git fetch logic
│   │   │   ├── fetcher.go
│   │   │   └── fetcher_test.go
│   │   ├── scheduler/            # Task scheduling & hot-reload
│   │   │   ├── scheduler.go
│   │   │   └── scheduler_test.go
│   │   ├── web/                  # HTTP server & API
│   │   │   ├── handler.go
│   │   │   ├── handler_test.go
│   │   │   └── templates/
│   │   │       └── index.html
│   │   ├── main.go               # Entry point
│   │   ├── Dockerfile
│   │   ├── Makefile              # Build & test commands
│   │   ├── go.mod                # Go dependencies
│   │   ├── go.sum
│   │   ├── README.md             # GitFetcher documentation
│   │   └── config.example.yaml
│   │
│   └── github-sync/              # GitHub issue synchronization
│       ├── cmd/
│       │   └── sync/
│       │       └── main.go       # Entry point
│       ├── internal/
│       │   ├── config/           # Viper-based config with hot-reload
│       │   │   ├── config.go
│       │   │   └── config_test.go
│       │   ├── github/           # GitHub API client
│       │   │   ├── client.go
│       │   │   └── client_test.go
│       │   ├── redmine/          # Redmine API client
│       │   │   ├── client.go
│       │   │   └── client_test.go
│       │   ├── storage/          # PostgreSQL persistence
│       │   │   └── postgres.go
│       │   └── sync/             # Core synchronization logic
│       │       ├── scheduler.go
│       │       ├── syncer.go
│       │       └── syncer_test.go
│       ├── Dockerfile
│       ├── go.mod                # Go dependencies
│       ├── go.sum
│       └── README.md             # GitHub Sync documentation
│
└── ssh_keys/                     # SSH keys for Git access (gitignored)
```

## Tech Stack

### GitFetcher
- **Language**: Go 1.23+
- **Web Framework**: Gin (HTTP server)
- **Config**: go-yaml/v3
- **File Watching**: fsnotify (hot-reload)
- **Container**: Alpine Linux + Git + OpenSSH
- **Testing**: Go testing framework + httptest

### GitHub Sync
- **Language**: Go 1.23+
- **Config**: Viper (hot-reload support)
- **Database**: PostgreSQL with lib/pq driver
- **APIs**: Redmine REST API, GitHub REST API
- **Testing**: testify framework

### Redmine
- **Base Image**: redmine:latest (Debian/Bookworm)
- **Ruby**: Bundled with official image
- **Database**: PostgreSQL 15 Alpine
- **Plugins**: Mermaid Macro

## Development Workflows

### Setting Up Development Environment

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd redmine_git
   ```

2. **Prepare configuration files**:
   ```bash
   cp plugins/gitfetcher/config.example.yaml gitfetcher-config.yaml
   cp github-sync-config.example.yaml github-sync-config.yaml
   ```

3. **Set up SSH keys** (for GitFetcher):
   ```bash
   mkdir -p ssh_keys
   cp ~/.ssh/id_ed25519 ssh_keys/
   chmod 600 ssh_keys/id_ed25519
   ```

4. **Start all services**:
   ```bash
   docker compose up -d
   ```

### Working with GitFetcher

**Local development**:
```bash
cd plugins/gitfetcher

# Run tests
make test

# Run with verbose output
make test-verbose

# Check coverage
make test-coverage

# Run with race detector
make test-race

# Format code
make fmt

# Run all checks
make check

# Build binary
make build

# Run locally
go run main.go -config ../../gitfetcher-config.yaml
```

**Testing individual modules**:
```bash
make test-config    # Test config package only
make test-fetcher   # Test fetcher package only
make test-scheduler # Test scheduler package only
make test-web       # Test web package only
```

### Working with GitHub Sync

**Local development**:
```bash
cd plugins/github-sync

# Run tests
go test ./...

# Verbose tests
go test -v ./...

# Run with environment variables
POSTGRES_HOST=localhost \
POSTGRES_PORT=5432 \
POSTGRES_DB=redmine \
POSTGRES_USER=redmine \
POSTGRES_PASSWORD=redmine \
go run cmd/sync/main.go -config ../../github-sync-config.yaml
```

### Docker Workflows

**Build and run specific services**:
```bash
# Build all services
docker compose build

# Build specific service
docker compose build gitfetcher
docker compose build github-sync

# Start specific service
docker compose up -d gitfetcher

# View logs
docker compose logs -f gitfetcher
docker compose logs -f github-sync

# Restart service
docker compose restart github-sync

# Stop all
docker compose down

# Stop and remove volumes (CAUTION: deletes data)
docker compose down -v
```

**Exec into containers**:
```bash
docker compose exec redmine bash
docker compose exec gitfetcher sh
docker compose exec github-sync sh
docker compose exec postgres_db psql -U redmine -d redmine
```

## Testing Conventions

### GitFetcher Testing

- **Coverage Target**: 85%+ overall
- **Test Files**: Located alongside source files with `_test.go` suffix
- **Test Pattern**: Table-driven tests preferred
- **Mocking**: Use interfaces for external dependencies
- **Race Detection**: Always run `make test-race` before commits

**Example test structure**:
```go
func TestFetchRepository(t *testing.T) {
    tests := []struct {
        name    string
        repo    Repository
        wantErr bool
    }{
        {
            name: "valid repository",
            repo: Repository{Name: "test", URL: "git@github.com:user/repo.git"},
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### GitHub Sync Testing

- **Framework**: testify for assertions
- **Test Files**: `*_test.go` alongside source
- **Mocking**: Mock HTTP clients for API testing
- **Database**: Use test transactions where possible

## Git Conventions

### Branch Naming

Follow this pattern for Claude Code sessions:
- Format: `claude/claude-md-{session-id}`
- Example: `claude/claude-md-mi6yoo0r47cxak9q-015GVgAEJo7JrezS684V3sy3`

**Important**: Always develop on the designated Claude branch, not on main/master.

### Commit Messages

Follow conventional commits:
- `feat:` - New features
- `fix:` - Bug fixes
- `chore:` - Maintenance tasks (deps, config)
- `refactor:` - Code restructuring
- `test:` - Test additions/modifications
- `docs:` - Documentation changes

**Examples from this repo**:
- `feat: Enhance GitFetcher with clone and fetch functionality`
- `fix: Update fetch method in Scheduler to use URL from status`
- `chore: Update .gitignore to include gitfetcher-config.local.yaml`

### Git Operations

**Pushing changes**:
```bash
# Always use -u flag for new branches
git push -u origin claude/claude-md-{session-id}
```

**Network retry policy**:
- Retry up to 4 times with exponential backoff (2s, 4s, 8s, 16s)
- Applies to: git push, git fetch, git pull

**Fetching updates**:
```bash
# Prefer specific branch fetches
git fetch origin claude/claude-md-{session-id}
```

## Configuration Management

### Hot-Reload Support

Both GitFetcher and GitHub Sync support configuration hot-reload:

**GitFetcher**:
- Uses fsnotify to watch `config.yaml`
- Web UI config editor available at http://localhost:13100
- Changes apply without container restart

**GitHub Sync**:
- Uses Viper with fsnotify
- Automatically reloads on file changes
- Note: Some editors (vim) may break inode watching - restart container if needed

### Configuration Files

**gitfetcher-config.yaml**:
```yaml
repos:
  - name: "project-name"
    url: "git@github.com:org/repo.git"
    local_path: "/repos/project-name.git"
    interval: "5m"  # Go duration: 5s, 10m, 1h, etc.

ssh_key_path: "/root/.ssh/id_ed25519"
http_port: 8080
log_path: "./logs"
```

**github-sync-config.yaml**:
```yaml
redmine:
  url: "http://redmine:3000"              # Internal API URL
  display_url: "http://192.168.1.100:13000"  # External display URL
  api_key: "your-redmine-api-key"
  projects:
    - identifier: "my-project"
      custom_fields:
        target_repo_id: 10      # Custom field ID for repo selection
        github_issue_url_id: 11 # Custom field ID for GitHub URL

github:
  token: "ghp_xxxxxxxxxxxx"
  base_url: "https://github.com"

sync:
  interval: "5m"
  title_format: "[Redmine #%d] %s"
```

### Environment Variables

**GitHub Sync** requires these env vars:
- `POSTGRES_HOST` - PostgreSQL hostname (default: localhost)
- `POSTGRES_PORT` - PostgreSQL port (default: 5432)
- `POSTGRES_DB` - Database name (default: redmine)
- `POSTGRES_USER` - Database user
- `POSTGRES_PASSWORD` - Database password
- `POSTGRES_SCHEMA` - Schema name (default: redmine_github_sync)
- `CONFIG_PATH` - Config file path (default: ./config.yaml)

## Important Conventions for AI Assistants

### Code Style

1. **Go Code**:
   - Always run `go fmt` before committing
   - Use `make check` for comprehensive checks
   - Follow Go naming conventions (MixedCaps, not snake_case)
   - Keep functions small and focused
   - Add comments for exported functions/types

2. **Error Handling**:
   - Never ignore errors
   - Wrap errors with context: `fmt.Errorf("context: %w", err)`
   - Log errors with appropriate severity

3. **Testing**:
   - Write tests for all new features
   - Maintain 85%+ coverage for GitFetcher
   - Use table-driven tests
   - Test error paths, not just happy paths

### File Operations

1. **Always prefer editing** existing files over creating new ones
2. **Never create** markdown files unless explicitly requested
3. **Use Read tool** before editing any file
4. **Check file paths** exist before referencing them

### Docker Operations

1. **Don't restart** containers unnecessarily
2. **Use docker compose** commands, not raw docker commands
3. **Check logs** before assuming failures
4. **Be aware** of shared volumes between Redmine and GitFetcher

### Database Operations

1. **Schema isolation**: Each service uses its own schema
2. **Never modify** Redmine's `public` schema directly
3. **Use migrations** for schema changes (not implemented yet, but future requirement)
4. **Test database** operations with transactions

### Configuration Changes

1. **Test locally** before updating production configs
2. **Validate YAML** syntax before saving
3. **Document** all config changes in commit messages
4. **Remember** hot-reload may not work with all editors

## Common Tasks

### Adding a New Repository to GitFetcher

**Option 1: Web UI**
1. Navigate to http://localhost:13100
2. Click "Edit Configuration"
3. Add new repository entry
4. Save (auto-reloads)

**Option 2: Manual**
1. Edit `gitfetcher-config.yaml`
2. Add repository to `repos` array
3. Clone the repository manually first:
   ```bash
   git clone --mirror git@github.com:org/repo.git repos/repo-name.git
   ```
4. Save config (auto-reloads)

### Adding a New Redmine Project to GitHub Sync

1. Create custom fields in Redmine (if not exists):
   - "Target GitHub Repo" (List type)
   - "GitHub Issue URL" (Link type)

2. Edit `github-sync-config.yaml`:
   ```yaml
   redmine:
     projects:
       - identifier: "new-project"
         custom_fields:
           target_repo_id: 10    # Your field ID
           github_issue_url_id: 11
   ```

3. Config auto-reloads, check logs:
   ```bash
   docker compose logs -f github-sync
   ```

### Debugging Failed Synchronization

**GitFetcher**:
```bash
# Check service status
docker compose ps gitfetcher

# View logs
docker compose logs -f gitfetcher

# Check SSH connectivity
docker compose exec gitfetcher ssh -T git@github.com

# View detailed fetch logs
docker compose exec gitfetcher cat /app/logs/fetch-$(date +%Y-%m-%d).log

# Manually trigger fetch via API
curl -X POST http://localhost:13100/api/fetch/repo-name
```

**GitHub Sync**:
```bash
# Check service status
docker compose ps github-sync

# View logs
docker compose logs -f github-sync

# Check database records
docker compose exec postgres_db psql -U redmine -d redmine -c \
  "SELECT * FROM redmine_github_sync.sync_records ORDER BY synced_at DESC LIMIT 10;"

# Check errors
docker compose exec postgres_db psql -U redmine -d redmine -c \
  "SELECT * FROM redmine_github_sync.sync_errors WHERE resolved = FALSE;"
```

### Running Tests in Docker

**GitFetcher**:
```bash
# Run tests in clean container
docker compose exec gitfetcher make test

# With coverage
docker compose exec gitfetcher make test-coverage
```

**GitHub Sync**:
```bash
# Run tests
docker compose exec github-sync go test ./...

# Verbose
docker compose exec github-sync go test -v ./...
```

## Security Considerations

1. **API Keys & Tokens**:
   - Never commit `github-sync-config.yaml` with real tokens
   - Use `.gitignore` for local configs
   - Rotate keys periodically

2. **SSH Keys**:
   - Use deploy keys with read-only access when possible
   - Set proper permissions: `chmod 600`
   - Mount as read-only in containers (`:ro`)

3. **Database**:
   - Change default passwords in production
   - Keep PostgreSQL on internal network only
   - Regular backups of volumes

4. **Network**:
   - Only expose necessary ports (13000, 13100)
   - Use reverse proxy for production
   - Enable HTTPS/TLS in production

## Troubleshooting Guide

### GitFetcher Issues

**Problem**: SSH authentication fails
- **Solution**: Check SSH key permissions, test with `ssh -T git@github.com`

**Problem**: Repository not found
- **Solution**: Ensure repository is cloned as bare: `git clone --mirror <url>`

**Problem**: Hot-reload not working
- **Solution**: Restart container or use Web UI editor instead of vim

### GitHub Sync Issues

**Problem**: Issues not syncing
- **Check**: "Target GitHub Repo" field is filled in Redmine issue
- **Check**: GitHub token has `repo` permissions
- **Check**: Redmine API key is valid
- **Check**: Service logs for errors

**Problem**: Database connection failed
- **Solution**: Verify PostgreSQL is running, check environment variables

**Problem**: Config hot-reload not working
- **Solution**: Some editors break fsnotify, restart container

### General Docker Issues

**Problem**: Port already in use
- **Solution**: Stop conflicting service or change port in docker-compose.yml

**Problem**: Permission denied on volumes
- **Solution**: Check file ownership, may need to chown to appropriate UID

**Problem**: Out of disk space
- **Solution**: Clean up: `docker system prune -a`, check volume sizes

## Resources

- **Main README**: [README.md](README.md) - Project overview (Chinese)
- **GitFetcher Docs**: [plugins/gitfetcher/README.md](plugins/gitfetcher/README.md)
- **GitHub Sync Docs**: [plugins/github-sync/README.md](plugins/github-sync/README.md)
- **Redmine Documentation**: https://www.redmine.org/projects/redmine/wiki
- **Docker Compose Reference**: https://docs.docker.com/compose/

## Version Information

- **Go**: 1.23+
- **PostgreSQL**: 15 Alpine
- **Redmine**: Latest (from official Docker image)
- **Docker Compose**: v2+ (uses `docker compose`, not `docker-compose`)

## Development Philosophy

This project follows **MVP + KISS** principles:
- **MVP (Minimum Viable Product)**: Focus on core functionality first
- **KISS (Keep It Simple, Stupid)**: Simple, reliable, maintainable code
- **Testing**: High test coverage ensures reliability
- **Documentation**: Comprehensive docs in both Chinese and English
- **Automation**: Hot-reload, auto-sync, minimal manual intervention

## Notes for AI Assistants

1. **Read before editing**: Always use the Read tool before modifying files
2. **Test changes**: Run appropriate tests after code changes
3. **Follow conventions**: Respect existing code style and patterns
4. **Documentation**: The README files are in Chinese - this is intentional
5. **Commit messages**: Use conventional commits format
6. **Branch discipline**: Always work on designated Claude branches
7. **Volume awareness**: Remember Redmine and GitFetcher share the repositories volume
8. **Schema isolation**: Respect database schema boundaries
9. **Hot-reload**: Take advantage of config hot-reload when testing
10. **Logs are your friend**: Check logs before assuming something is broken

---

**Last Updated**: 2025-11-20
**Maintained by**: CLCS Admin & AI Assistants
