CREATE SCHEMA "MySchema";

CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text
);

CREATE TABLE public.order_items (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) NOT NULL,
    product_id bigint NOT NULL
);

CREATE TABLE public."userProfiles" (
    id bigint PRIMARY KEY,
    "First Name" text NOT NULL,
    "Last Name" text NOT NULL
);

CREATE TABLE "MySchema".products (
    id bigint PRIMARY KEY,
    name text NOT NULL
);
