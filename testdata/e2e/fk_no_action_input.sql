-- Test fixture for FK with NO ACTION
-- Covers: ON DELETE NO ACTION, ON UPDATE NO ACTION

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

-- products (referenced by order_items)
CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

CREATE TABLE public.posts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    title text NOT NULL,
    content text
);

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE NO ACTION;

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE NO ACTION;

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
    ADD CONSTRAINT order_items_product_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE NO ACTION ON UPDATE NO ACTION;

CREATE TABLE public.comments (
    id bigint NOT NULL,
    post_id bigint,
    user_id bigint,
    content text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_post_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE NO ACTION;

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE NO ACTION ON UPDATE NO ACTION;
