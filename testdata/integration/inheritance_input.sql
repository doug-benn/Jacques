-- Test fixture for table inheritance
-- Covers: INHERITS clause, child tables, parent-child relationships

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
