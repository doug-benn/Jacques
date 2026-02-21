CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    price numeric(10,2) NOT NULL,
    discount numeric(10,2) DEFAULT 0,
    final_price numeric(10,2) GENERATED ALWAYS AS (price - discount) STORED
);

CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    first_name text NOT NULL,
    last_name text NOT NULL,
    full_name text GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW(),
    age interval GENERATED ALWAYS AS (updated_at - created_at) STORED
);

CREATE TABLE public.events (
    id bigint PRIMARY KEY,
    data jsonb NOT NULL,
    event_type text NOT NULL,
    is_urgent boolean GENERATED ALWAYS AS (data->>'urgent' = 'true') STORED
);
