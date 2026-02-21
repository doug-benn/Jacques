CREATE TABLE public.users (
    id BIGSERIAL PRIMARY KEY,
    email text UNIQUE,
    created_at timestamp without time zone
);
