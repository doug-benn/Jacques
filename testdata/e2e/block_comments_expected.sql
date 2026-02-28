-- Test fixture for block comment removal
CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    PRIMARY KEY (id)
);

-- Table for products
CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    PRIMARY KEY (id)
);
