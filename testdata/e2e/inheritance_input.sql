-- Test fixture for basic table inheritance (E2E testable)
-- Covers: Simple parent-child inheritance
-- Note: Complex inheritance still requires ExperimentalFolding

-- Parent table
CREATE TABLE public.users (
    id bigint NOT NULL,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

-- Child table: administrators
CREATE TABLE public.administrators (
    role text NOT NULL DEFAULT 'admin'
) INHERITS (public.users);

ALTER TABLE ONLY public.administrators
    ADD CONSTRAINT administrators_pkey PRIMARY KEY (id);
