CREATE TYPE public.order_status AS ENUM ('pending', 'processing', 'shipped', 'delivered');

CREATE TABLE public.reservations (
    id bigint PRIMARY KEY,
    room_number text NOT NULL,
    stay_period tstzrange NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.inventory (
    id bigint PRIMARY KEY,
    product_id bigint NOT NULL,
    quantity_range int4range NOT NULL,
    last_updated timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    birth_date date,
    age_range int4range,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.events (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    event_time tsrange NOT NULL,
    duration interval
);

CREATE TABLE public.order_versions (
    id bigint PRIMARY KEY,
    order_id bigint NOT NULL,
    status_range daterange NOT NULL,
    status order_status NOT NULL
);
