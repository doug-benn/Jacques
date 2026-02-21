-- Test fixture for multiple schemas
-- Covers: multiple schemas, cross-schema FKs, schema-qualified names
-- Note: Tables are in dependency order (referenced tables before referencing tables)

-- Create schemas
CREATE SCHEMA app;
CREATE SCHEMA audit;
CREATE SCHEMA finance;

-- Schema: public (shared reference table) - must be first since others reference it
CREATE TABLE public.countries (
    id bigint NOT NULL,
    code text NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT countries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT countries_code_key UNIQUE (code);

-- Schema: app (application tables)
CREATE TABLE app.users (
    id bigint NOT NULL,
    email text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY app.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY app.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

-- Add country_id column and FK to app.users
ALTER TABLE app.users ADD COLUMN country_id bigint;

ALTER TABLE ONLY app.users
    ADD CONSTRAINT users_country_fkey FOREIGN KEY (country_id) REFERENCES public.countries(id);

-- Schema: app (orders in app schema)
CREATE TABLE app.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY app.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

-- Cross-schema FK: app.orders.user_id -> app.users.id
ALTER TABLE ONLY app.orders
    ADD CONSTRAINT orders_user_fkey FOREIGN KEY (user_id) REFERENCES app.users(id);

-- Schema: audit (audit logs)
CREATE TABLE audit.audit_logs (
    id bigint NOT NULL,
    user_id bigint,
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id bigint,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY audit.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);

-- Cross-schema FK: audit.audit_logs.user_id -> app.users.id
ALTER TABLE ONLY audit.audit_logs
    ADD CONSTRAINT audit_logs_user_fkey FOREIGN KEY (user_id) REFERENCES app.users(id);

-- Schema: finance (financial data)
CREATE TABLE finance.invoices (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    amount numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY finance.invoices
    ADD CONSTRAINT invoices_pkey PRIMARY KEY (id);

-- Cross-schema FK: finance.invoices.order_id -> app.orders.id
ALTER TABLE ONLY finance.invoices
    ADD CONSTRAINT invoices_order_fkey FOREIGN KEY (order_id) REFERENCES app.orders(id);
