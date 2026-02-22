-- Test fixture for exclusion constraints
-- Exclusion constraints are passed through as ALTER TABLE statements
-- Uses btree (standard PostgreSQL) - no extension required

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

CREATE TABLE public.users (
    id bigint NOT NULL,
    name text NOT NULL,
    email text NOT NULL
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE TABLE public.reservations (
    id bigint NOT NULL,
    room_id bigint NOT NULL,
    booking_date date NOT NULL
);

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_no_double_booking EXCLUDE USING btree (room_id WITH =, booking_date WITH =);
