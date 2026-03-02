CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name text NOT NULL
);
