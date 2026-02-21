CREATE SCHEMA app;

CREATE SCHEMA audit;

CREATE SCHEMA finance;

CREATE TABLE public.countries (
    id bigint PRIMARY KEY,
    code text NOT NULL UNIQUE,
    name text NOT NULL
);

CREATE TABLE app.users (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    country_id bigint REFERENCES public.countries(id)
);

CREATE TABLE app.orders (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES app.users(id) NOT NULL,
    total numeric(10,2) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE audit.audit_logs (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES app.users(id),
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id bigint,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE finance.invoices (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES app.orders(id) NOT NULL,
    amount numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);
