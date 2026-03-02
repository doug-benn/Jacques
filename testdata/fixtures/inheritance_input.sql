-- Test fixture for table inheritance
-- Features tested:
--   - INHERITS clause: Parent-child table relationships
--   - Parent tables: Base tables with inherited columns
--   - Child tables: Tables inheriting from parents
--   - Additional child columns: Child-specific columns beyond parent
--   - FK to parent: References to parent table
--   - Primary keys on children: Separate PKs on child tables
--   - Multi-parent inheritance: Inherit from multiple parents
--
-- Input: pg_dump output with inherited tables
-- Expected: Clean inheritance output
--
-- Note: Multi-parent inheritance requires --experimental-folding

-- Parent table
CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

-- Child table: administrators
CREATE TABLE public.administrators (
    role text NOT NULL DEFAULT 'admin',
    permissions text[] NOT NULL
) INHERITS (public.users);

ALTER TABLE ONLY public.administrators
    ADD CONSTRAINT administrators_pkey PRIMARY KEY (id);

-- Child table: moderators
CREATE TABLE public.moderators (
    department text NOT NULL,
    badge_number text
) INHERITS (public.users);

ALTER TABLE ONLY public.moderators
    ADD CONSTRAINT moderators_pkey PRIMARY KEY (id);

-- Child table: regular users (with additional profile fields)
CREATE TABLE public.registered_users (
    email_verified boolean NOT NULL DEFAULT false,
    last_login timestamp without time zone
) INHERITS (public.users);

ALTER TABLE ONLY public.registered_users
    ADD CONSTRAINT registered_users_pkey PRIMARY KEY (id);

-- Table referencing parent
CREATE TABLE public.user_sessions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    token text NOT NULL,
    expires_at timestamp without time zone NOT NULL
);

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- Multi-parent inheritance (requires ExperimentalFolding)
CREATE TABLE public.accounts (
    id bigint NOT NULL,
    username text NOT NULL
);

ALTER TABLE ONLY public.accounts ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

CREATE TABLE public.profiles (
    id bigint NOT NULL,
    bio text
);

ALTER TABLE ONLY public.profiles ADD CONSTRAINT profiles_pkey PRIMARY KEY (id);

CREATE TABLE public.user_profiles (
    avatar_url text
) INHERITS (public.accounts, public.profiles);

ALTER TABLE ONLY public.user_profiles ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (id);
