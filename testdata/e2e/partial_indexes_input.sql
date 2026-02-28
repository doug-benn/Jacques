-- Test fixture for partial indexes and index options
-- Covers: CREATE INDEX with WHERE clause, INCLUDE columns, expression indexes

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

-- Partial index (only index active users)
CREATE INDEX idx_users_active ON public.users(status) WHERE status = 'active';

-- Regular index on email
CREATE INDEX idx_users_email ON public.users(email);

-- Partial index with multiple conditions
CREATE INDEX idx_users_pending_verified ON public.users(status, email) WHERE status IN ('pending', 'verified');

-- Index with INCLUDE columns (covering index)
CREATE INDEX idx_users_status_include ON public.users(status) INCLUDE (email, created_at);

-- Unique index with INCLUDE
CREATE UNIQUE INDEX idx_users_email_covering ON public.users(email) INCLUDE (status);

-- ============================================
-- Expression indexes (indexes on expressions/functions)
-- ============================================

-- Edge case: Index on lowercased column (case-insensitive search)
CREATE INDEX idx_users_email_lower ON public.users((lower(email)));

-- Edge case: Index on expression
CREATE INDEX idx_users_status_active ON public.users((status = 'active')) WHERE status = 'active';

-- Edge case: Index on function result
CREATE INDEX idx_users_created_year ON public.users((EXTRACT(YEAR FROM created_at)));

-- Index on concat expression
CREATE INDEX idx_users_name_lower ON public.users((lower(name)));
