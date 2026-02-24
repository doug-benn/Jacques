-- Test fixture for ONLY keyword removal
-- Features tested:
--   - ONLY removal: Remove ONLY from ALTER TABLE statements (pass-through statements)
--
-- Input: pg_dump output with ALTER TABLE ONLY statements that don't get folded
-- Expected: Same statements but with ONLY keyword removed

CREATE TABLE public.users (
    id bigint NOT NULL,
    name text
);

-- ALTER statements that pass through unchanged - ONLY should be removed
ALTER TABLE ONLY public.users ADD COLUMN new_col int;

ALTER TABLE ONLY public.users RENAME TO new_users;

ALTER TABLE ONLY public.new_users RENAME COLUMN name TO user_name;
