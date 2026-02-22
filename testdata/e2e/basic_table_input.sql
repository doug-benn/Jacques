-- Test fixture for basic table consolidation
-- Features tested:
--   - Noise removal: SET, OWNER TO, COMMENT
--   - Sequence → SERIAL: sequence converted to BIGSERIAL
--   - ALTER folding: PK, UNIQUE constraints folded into CREATE TABLE
--   - ONLY removal: ALTER TABLE ONLY → ALTER TABLE
--
-- Input: pg_dump --schema-only output with noise and ALTER statements
-- Expected: Clean CREATE TABLE with folded constraints

SET statement_timeout = 0;           -- Noise: removed
SET lock_timeout = 0;               -- Noise: removed

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text,
    created_at timestamp without time zone
);

CREATE SEQUENCE public.users_id_seq  -- Converted to SERIAL type
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;  -- Noise: removed

ALTER TABLE public.users ALTER COLUMN id
    SET DEFAULT nextval('public.users_id_seq'::regclass);  -- Folded into BIGSERIAL

ALTER TABLE ONLY public.users       -- ONLY removed
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);  -- Folded into table

ALTER TABLE ONLY public.users       -- ONLY removed
    ADD CONSTRAINT users_email_key UNIQUE (email);  -- Folded into table

ALTER TABLE public.users OWNER TO testuser;  -- Noise: removed

COMMENT ON TABLE public.users IS 'Application users';  -- Noise: removed
