ALTER TABLE chatbots DROP COLUMN IF EXISTS organization_id;
ALTER TABLE chatbots DROP COLUMN IF EXISTS workspace_id;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS organizations;
