CREATE SEQUENCE order_ids
    START WITH 1000
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 999999999
    CACHE 10;

CREATE SEQUENCE product_ids
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 1000
    CACHE 5
    CYCLE;

CREATE TABLE orders (
    id bigint PRIMARY KEY DEFAULT nextval('public.order_ids'::regclass),
    customer_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

CREATE TABLE products (
    id bigint PRIMARY KEY DEFAULT nextval('public.product_ids'::regclass),
    name text NOT NULL
);

ALTER SEQUENCE order_ids RESTART WITH 2000;

ALTER SEQUENCE product_ids INCREMENT BY 2;

ALTER SEQUENCE product_ids RESTART WITH 100;
