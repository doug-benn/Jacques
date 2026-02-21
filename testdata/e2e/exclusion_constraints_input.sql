-- Test fixture for exclusion constraints
-- Exclusion constraints are passed through as ALTER TABLE statements

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

CREATE TABLE public.calendar (
    id bigint NOT NULL,
    title text NOT NULL,
    exclude_range date NOT NULL
);

ALTER TABLE ONLY public.calendar
    ADD CONSTRAINT calendar_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.calendar
    ADD CONSTRAINT calendar_exclude_date EXCLUDE USING gist (exclude_range WITH &&);

CREATE TABLE public.reservations (
    id bigint NOT NULL,
    room_id bigint NOT NULL,
    start_time timestamp NOT NULL,
    end_time timestamp NOT NULL
);

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_no_overlap EXCLUDE USING btree (room_id WITH =, tstzrange(start_time, end_time) WITH &&);
