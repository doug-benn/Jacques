CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL
);

CREATE TABLE public.calendar (
    id bigint PRIMARY KEY,
    title text NOT NULL,
    exclude_range date NOT NULL,
    CONSTRAINT calendar_exclude_date EXCLUDE USING gist (exclude_range WITH &&)
);

CREATE TABLE public.reservations (
    id bigint PRIMARY KEY,
    room_id bigint NOT NULL,
    start_time timestamp NOT NULL,
    end_time timestamp NOT NULL,
    CONSTRAINT reservations_no_overlap EXCLUDE USING btree (room_id WITH =, tstzrange(start_time, end_time) WITH &&)
);
