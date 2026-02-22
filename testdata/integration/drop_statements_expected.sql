CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL
);

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL
);

DROP TABLE IF EXISTS public.users;

DROP TABLE IF EXISTS public.orders;

DROP INDEX IF EXISTS idx_users_email;

DROP INDEX IF EXISTS idx_orders_user;

CREATE INDEX idx_users_email ON public.users(email);

CREATE INDEX idx_orders_user ON public.orders(user_id);
