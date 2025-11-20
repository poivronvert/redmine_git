-- Migration: Add multi-platform support to sync_records table
-- Date: 2025-11-20
-- Description: Add platform column and update constraints to support both GitHub and GitLab

-- Add platform column (default to 'github' for existing records)
ALTER TABLE redmine_github_sync.sync_records
ADD COLUMN IF NOT EXISTS platform VARCHAR(20) NOT NULL DEFAULT 'github';

-- Drop the old unique constraint on redmine_issue_id
ALTER TABLE redmine_github_sync.sync_records
DROP CONSTRAINT IF EXISTS sync_records_redmine_issue_id_key;

-- Add new composite unique constraint (redmine_issue_id, platform)
-- This allows the same Redmine issue to be synced to multiple platforms
ALTER TABLE redmine_github_sync.sync_records
ADD CONSTRAINT sync_records_unique_platform
UNIQUE (redmine_issue_id, platform);

-- Create index on platform for faster queries
CREATE INDEX IF NOT EXISTS idx_platform
ON redmine_github_sync.sync_records(platform);

-- Create composite index for common queries
CREATE INDEX IF NOT EXISTS idx_redmine_issue_platform
ON redmine_github_sync.sync_records(redmine_issue_id, platform);

-- Update github_repo column comment to reflect it's now a generic field
COMMENT ON COLUMN redmine_github_sync.sync_records.github_repo IS
'Target project identifier (GitHub: owner/repo, GitLab: project_id or namespace/project)';

COMMENT ON COLUMN redmine_github_sync.sync_records.github_issue_number IS
'Issue number (GitHub: number, GitLab: iid)';

COMMENT ON COLUMN redmine_github_sync.sync_records.platform IS
'Platform name: github or gitlab';
