CREATE TABLE customers (
    id bigint NOT NULL,
    name text NOT NULL,
    contact contact_info,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id bigint NOT NULL,
    email email NOT NULL,
    phone phone_number,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    quantity positive_int NOT NULL,
    status order_status NOT NULL DEFAULT 'pending'
);
