CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.administrators (
    role text NOT NULL DEFAULT 'admin',
    permissions text[] NOT NULL
);

CREATE TABLE public.moderators (
    department text NOT NULL,
    badge_number text
);

CREATE TABLE public.registered_users (
    email_verified boolean NOT NULL DEFAULT false,
    last_login timestamp without time zone
);

CREATE TABLE public.user_sessions (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) ON DELETE CASCADE NOT NULL,
    token text NOT NULL,
    expires_at timestamp without time zone NOT NULL
);
