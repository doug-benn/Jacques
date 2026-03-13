-- Test fixture for Row Level Security (RLS) policies
-- Covers: ENABLE ROW LEVEL SECURITY, and various CREATE POLICY actions

CREATE TABLE rls_test (
    id bigint NOT NULL,
    owner_id bigint NOT NULL,
    val text
);

ALTER TABLE ONLY rls_test ADD PRIMARY KEY (id);

-- Enable RLS
ALTER TABLE rls_test ENABLE ROW LEVEL SECURITY;

-- 1. Simple SELECT policy
CREATE POLICY select_all ON rls_test FOR SELECT USING (true);

-- 2. Policy with expression (e.g. owner check)
CREATE POLICY owner_all ON rls_test FOR ALL TO public USING (owner_id = 1);
