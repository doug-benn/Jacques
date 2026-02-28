CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE,
    status text NOT NULL DEFAULT 'active',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_active ON public.users(status) WHERE status = 'active';

CREATE INDEX idx_users_email ON public.users(email);

CREATE INDEX idx_users_pending_verified ON public.users(status, email) WHERE status IN ('pending', 'verified');

CREATE INDEX idx_users_status_include ON public.users(status) INCLUDE (email, created_at);

CREATE UNIQUE INDEX idx_users_email_covering ON public.users(email) INCLUDE (status);

CREATE INDEX idx_users_email_lower ON public.users((lower(email)));

CREATE INDEX idx_users_status_active ON public.users((status = 'active')) WHERE status = 'active';

CREATE INDEX idx_users_created_year ON public.users((EXTRACT(YEAR FROM created_at)));

CREATE INDEX idx_users_name_lower ON public.users((lower(name)));
