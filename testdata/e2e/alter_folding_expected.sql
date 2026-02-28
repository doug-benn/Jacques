CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text,
    name text NOT NULL,
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    sku text NOT NULL,
    name text NOT NULL,
    UNIQUE (sku)
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    order_number text NOT NULL,
    UNIQUE (user_id, order_number)
);

CREATE TABLE public.accounts (
    id bigint PRIMARY KEY,
    balance numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'active',
    CHECK (balance >= 0),
    CHECK (status IN ('active', 'inactive', 'suspended'))
);

CREATE TABLE public.inventories (
    id bigint PRIMARY KEY,
    product_id bigint NOT NULL,
    quantity integer NOT NULL,
    reorder_point integer NOT NULL,
    CHECK (quantity >= 0),
    CHECK (reorder_point >= 0),
    CHECK (reorder_point <= quantity)
);

CREATE TABLE public.inline_notnull (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL,
    UNIQUE (email)
);

CREATE TABLE public.complex_check (
    id bigint PRIMARY KEY,
    price numeric(10,2) NOT NULL,
    discount numeric(10,2) NOT NULL,
    final_price numeric(10,2) NOT NULL,
    CHECK (price >= 0 AND price < 10000),
    CHECK (final_price = price - discount)
);

CREATE TABLE public.multi_constraint (
    id bigint PRIMARY KEY,
    code text NOT NULL,
    UNIQUE (code)
);

CREATE TABLE public.deferrable_unique (
    id bigint NOT NULL,
    email text NOT NULL
);

ALTER TABLE ONLY public.deferrable_unique
    ADD CONSTRAINT deferrable_unique_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.deferrable_unique
    ADD CONSTRAINT deferrable_unique_email_key UNIQUE (email) DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE public.using_clause (
    id bigint NOT NULL
);

ALTER TABLE ONLY public.using_clause
    ADD CONSTRAINT using_clause_pkey PRIMARY KEY (id) USING btree;
