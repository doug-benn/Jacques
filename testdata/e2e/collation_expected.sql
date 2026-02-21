CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    name text COLLATE "C" NOT NULL,
    email text COLLATE "default" NOT NULL UNIQUE,
    bio text
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) NOT NULL,
    description text COLLATE "C",
    notes text
);
