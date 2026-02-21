-- Test fixture for range types
-- Covers: DATERANGE, TSRANGE, INT4RANGE, TSTZRANGE as column types

CREATE TABLE public.reservations (
    id bigint NOT NULL,
    room_number text NOT NULL,
    stay_period tstzrange NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_pkey PRIMARY KEY (id);

CREATE TABLE public.inventory (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity_range int4range NOT NULL,
    last_updated timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.inventory
    ADD CONSTRAINT inventory_pkey PRIMARY KEY (id);

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    birth_date date,
    age_range int4range,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE TABLE public.events (
    id bigint NOT NULL,
    name text NOT NULL,
    event_time tsrange NOT NULL,
    duration interval
);

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);

CREATE TYPE public.order_status AS ENUM ('pending', 'processing', 'shipped', 'delivered');

CREATE TABLE public.order_versions (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    status_range daterange NOT NULL,
    status order_status NOT NULL
);

ALTER TABLE ONLY public.order_versions
    ADD CONSTRAINT order_versions_pkey PRIMARY KEY (id);
