CREATE TABLE users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    name text NOT NULL
);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_select_policy ON users
    FOR SELECT
    USING (true);
