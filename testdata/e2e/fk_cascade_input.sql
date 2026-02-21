-- Test fixture for FK cascade actions
-- Covers: ON DELETE, ON UPDATE, MATCH FULL/PARTIAL
-- Tables ordered so dependencies come first

-- Base tables (no dependencies)
CREATE TABLE public.users (
    id bigint NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE TABLE public.categories (
    id bigint NOT NULL,
    name text NOT NULL,
    parent_id bigint
);

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);

CREATE TABLE public.accounts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    balance numeric(10,2) NOT NULL DEFAULT 0
);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

-- products (referenced by order_items)
CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    category_id bigint
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_category_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE;

-- posts (referenced by comments)
CREATE TABLE public.posts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    title text NOT NULL
);

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- orders (referenced by order_items)
CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

-- order_items (depends on orders and products)
CREATE TABLE public.order_items (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL DEFAULT 1
);

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_product_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE SET DEFAULT;

-- comments (depends on posts and users)
CREATE TABLE public.comments (
    id bigint NOT NULL,
    post_id bigint,
    user_id bigint,
    content text NOT NULL
);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_post_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

-- transactions (depends on accounts)
CREATE TABLE public.transactions (
    id bigint NOT NULL,
    account_id bigint NOT NULL,
    amount numeric(10,2) NOT NULL,
    type text NOT NULL
);

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_account_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id) ON DELETE CASCADE ON UPDATE RESTRICT;

-- user_profiles (depends on users)
CREATE TABLE public.user_profiles (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    bio text
);

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) MATCH FULL;
