-- Test fixture for IF EXISTS on DROP statements
-- This mimics pg_dump --schema-only --clean output where DROP statements exist before CREATE
DROP TABLE public.users;
DROP TABLE public.orders;
DROP INDEX idx_users_email;
DROP INDEX idx_orders_user;

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL
);

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL
);

CREATE INDEX idx_users_email ON public.users(email);
CREATE INDEX idx_orders_user ON public.orders(user_id);
