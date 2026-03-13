CREATE TABLE rls_test (
    id bigint PRIMARY KEY,
    owner_id bigint NOT NULL,
    val text
);

ALTER TABLE rls_test ENABLE ROW LEVEL SECURITY;

CREATE POLICY select_all ON rls_test FOR SELECT USING (true);

CREATE POLICY owner_all ON rls_test FOR ALL TO public USING (owner_id = 1);
