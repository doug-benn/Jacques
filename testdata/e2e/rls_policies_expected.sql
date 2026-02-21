CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE,
    name text NOT NULL,
    role text NOT NULL DEFAULT 'user'
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) NOT NULL,
    total numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'pending'
);

CREATE TABLE public.documents (
    id bigint PRIMARY KEY,
    owner_id bigint NOT NULL,
    title text NOT NULL,
    content text,
    visibility text NOT NULL DEFAULT 'private'
);

CREATE TABLE public.accounts (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    balance numeric(10,2) NOT NULL DEFAULT 0
);

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

ALTER TABLE public.documents ENABLE ROW LEVEL SECURITY;

CREATE POLICY documents_owner_policy ON public.documents
    FOR ALL
    USING (owner_id = current_user_id())
    WITH CHECK (owner_id = current_user_id());

CREATE POLICY documents_public_read_policy ON public.documents
    FOR SELECT
    USING (visibility = 'public');

ALTER TABLE public.accounts FORCE ROW LEVEL SECURITY;

CREATE POLICY accounts_policy ON public.accounts
    FOR ALL
    USING (true);
