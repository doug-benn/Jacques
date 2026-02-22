CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    name text NOT NULL
);

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_select_policy ON public.users
    FOR SELECT
    USING (true);
