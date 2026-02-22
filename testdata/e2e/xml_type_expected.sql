CREATE TABLE public.documents (
    id bigint PRIMARY KEY,
    title text NOT NULL,
    content xml,
    metadata xml,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    specifications xml,
    category text NOT NULL
);
