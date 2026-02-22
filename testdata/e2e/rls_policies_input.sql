-- Test fixture for Row Level Security (RLS) policies
-- Covers: RLS enable, SELECT policy

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_select_policy ON public.users
    FOR SELECT
    USING (true);
