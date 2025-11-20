# Database Migrations

This directory contains SQL migration scripts for the Redmine GitHub/GitLab Sync service.

## Migration History

### 001 - Initial Schema (Built-in)
Created automatically by `internal/storage/postgres.go` when the service first starts.
- Creates `sync_records` table
- Creates `sync_errors` table
- Creates indexes

### 002 - Add Platform Support
**File**: `002_add_platform_support.sql`
**Date**: 2025-11-20

Adds multi-platform support to allow syncing to both GitHub and GitLab.

**Changes**:
- Adds `platform` column to `sync_records` table
- Updates unique constraint to `(redmine_issue_id, platform)`
- Adds platform-related indexes
- Updates column comments

## How to Apply Migrations

### Method 1: Manual Execution (Recommended for Production)

```bash
# Connect to your PostgreSQL database
psql -h localhost -U redmine -d redmine

# Execute the migration
\i /path/to/migrations/002_add_platform_support.sql
```

### Method 2: Using psql Command

```bash
psql -h localhost -U redmine -d redmine -f migrations/002_add_platform_support.sql
```

### Method 3: Docker Environment

```bash
docker exec -i super_redmine_postgres psql -U redmine -d redmine < migrations/002_add_platform_support.sql
```

## Rollback

Currently, rollback is not automated. If you need to rollback migration 002:

```sql
-- Remove platform column
ALTER TABLE redmine_github_sync.sync_records
DROP COLUMN IF EXISTS platform;

-- Restore old unique constraint
ALTER TABLE redmine_github_sync.sync_records
ADD CONSTRAINT sync_records_redmine_issue_id_key
UNIQUE (redmine_issue_id);

-- Drop platform indexes
DROP INDEX IF EXISTS redmine_github_sync.idx_platform;
DROP INDEX IF EXISTS redmine_github_sync.idx_redmine_issue_platform;
```

## Verification

After applying migration 002, verify the changes:

```sql
-- Check table structure
\d redmine_github_sync.sync_records

-- Check constraints
SELECT conname, contype FROM pg_constraint
WHERE conrelid = 'redmine_github_sync.sync_records'::regclass;

-- Check indexes
\di redmine_github_sync.*
```

## Notes

- Always backup your database before running migrations
- Test migrations in a development environment first
- The service will continue to work with the old schema (backward compatible)
- After migration, restart the sync service to use the new multi-platform features
