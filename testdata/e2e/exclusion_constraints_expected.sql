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

CREATE TABLE public.reservations (
    id bigint PRIMARY KEY,
    room_id bigint NOT NULL,
    booking_date date NOT NULL,
    CONSTRAINT reservations_no_double_booking EXCLUDE USING btree (room_id WITH =, booking_date WITH =)
);
