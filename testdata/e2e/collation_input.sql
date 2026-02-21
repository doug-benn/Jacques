-- Test fixture for collation
-- Covers: COLLATE clause in column definitions

CREATE TABLE public.users (
    id bigint NOT NULL,
    name text COLLATE "C" NOT NULL,
    email text COLLATE "default" NOT NULL,
    bio text
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    description text COLLATE "C",
    notes text
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);
