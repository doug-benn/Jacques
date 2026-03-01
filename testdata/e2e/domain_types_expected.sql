CREATE DOMAIN public.email AS text;

CREATE DOMAIN public.status AS text;

CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email public.email NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    status public.status NOT NULL DEFAULT 'pending'
);
