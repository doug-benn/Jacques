-- Test fixture for FK remaining actions not covered by other fixtures
-- Covers: ON DELETE SET DEFAULT, ON DELETE RESTRICT, ON DELETE NO ACTION,
--          ON UPDATE CASCADE, ON UPDATE SET NULL

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

-- Table with ON DELETE SET DEFAULT
CREATE TABLE public.accounts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    account_type text NOT NULL DEFAULT 'basic'
);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

-- FK with ON DELETE SET DEFAULT
ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET DEFAULT;

-- Table with ON DELETE RESTRICT (self-referential)
CREATE TABLE public.categories (
    id bigint NOT NULL,
    parent_id bigint,
    name text NOT NULL
);

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);

-- FK with ON DELETE RESTRICT (prevents deletion of parent)
ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.categories(id) ON DELETE RESTRICT;

-- Table with ON DELETE NO ACTION - references categories (which already exists)
CREATE TABLE public.products (
    id bigint NOT NULL,
    category_id bigint NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

-- FK with ON DELETE NO ACTION
ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_category_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE NO ACTION;

-- orders table must come BEFORE order_items since order_items references it
CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

-- Table with ON UPDATE CASCADE - references orders (which already exists)
CREATE TABLE public.order_items (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    product_id bigint NOT NULL
);

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);

-- FK with ON UPDATE CASCADE
ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON UPDATE CASCADE;

-- Table with ON UPDATE SET NULL - references orders (which already exists)
CREATE TABLE public.shipments (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    tracking_number text
);

ALTER TABLE ONLY public.shipments
    ADD CONSTRAINT shipments_pkey PRIMARY KEY (id);

-- FK with ON UPDATE SET NULL
ALTER TABLE ONLY public.shipments
    ADD CONSTRAINT shipments_order_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON UPDATE SET NULL;

-- addresses table - used for FK reference test
CREATE TABLE public.addresses (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    address_line1 text NOT NULL,
    city text NOT NULL
);

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT addresses_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT addresses_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);
