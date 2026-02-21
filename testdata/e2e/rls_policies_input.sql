-- Test fixture for RLS policies
-- Covers: ENABLE ROW LEVEL SECURITY, CREATE POLICY

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL,
    role text NOT NULL DEFAULT 'user'
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_select_policy ON public.users
    FOR SELECT
    USING (true);

CREATE POLICY users_insert_policy ON public.users
    FOR INSERT
    WITH CHECK (true);

CREATE POLICY users_update_policy ON public.users
    FOR UPDATE
    USING (true);

CREATE POLICY users_delete_policy ON public.users
    FOR DELETE
    USING (true);

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'pending'
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE public.orders ENABLE ROW LEVEL SECURITY;

CREATE POLICY orders_select_policy ON public.orders
    FOR SELECT
    USING (user_id = current_user_id());

CREATE POLICY orders_insert_policy ON public.orders
    FOR INSERT
    WITH CHECK (user_id = current_user_id());

CREATE POLICY orders_update_policy ON public.orders
    FOR UPDATE
    USING (user_id = current_user_id());

CREATE TABLE public.documents (
    id bigint NOT NULL,
    owner_id bigint NOT NULL,
    title text NOT NULL,
    content text,
    visibility text NOT NULL DEFAULT 'private'
);

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);

ALTER TABLE public.documents ENABLE ROW LEVEL SECURITY;

CREATE POLICY documents_owner_policy ON public.documents
    FOR ALL
    USING (owner_id = current_user_id())
    WITH CHECK (owner_id = current_user_id());

CREATE POLICY documents_public_read_policy ON public.documents
    FOR SELECT
    USING (visibility = 'public');

CREATE TABLE public.accounts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    balance numeric(10,2) NOT NULL DEFAULT 0
);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

ALTER TABLE public.accounts FORCE ROW LEVEL SECURITY;

CREATE POLICY accounts_policy ON public.accounts
    FOR ALL
    USING (true);
