BEGIN;

-- Drop existing foreign key constraints
ALTER TABLE chatbots DROP CONSTRAINT IF EXISTS chatbots_workspace_id_fkey;
ALTER TABLE chatbots DROP CONSTRAINT IF EXISTS chatbots_organization_id_fkey;

-- Re-add constraints with ON DELETE CASCADE
ALTER TABLE chatbots
    ADD CONSTRAINT chatbots_workspace_id_fkey
    FOREIGN KEY (workspace_id)
    REFERENCES workspaces(id)
    ON DELETE CASCADE;

ALTER TABLE chatbots
    ADD CONSTRAINT chatbots_organization_id_fkey
    FOREIGN KEY (organization_id)
    REFERENCES organizations(id)
    ON DELETE CASCADE;

COMMIT;
