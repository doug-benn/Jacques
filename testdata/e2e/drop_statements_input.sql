-- Test fixture for DROP statements with IF EXISTS (E2E testable)
-- DROP statements work with E2E because both DROP and DROP IF EXISTS
-- produce the same final schema on an empty database

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
