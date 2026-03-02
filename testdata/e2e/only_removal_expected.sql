CREATE TABLE users (
    id bigint NOT NULL,
    name text,
    new_col int
);

ALTER TABLE users RENAME TO new_users;

ALTER TABLE new_users RENAME COLUMN name TO user_name;
