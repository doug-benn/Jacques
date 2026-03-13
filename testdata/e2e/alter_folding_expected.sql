CREATE TABLE folding_basics (
    id bigint PRIMARY KEY,
    sku text NOT NULL UNIQUE,
    user_id bigint NOT NULL,
    order_number text NOT NULL,
    balance numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'active',
    UNIQUE (user_id, order_number),
    CHECK (balance >= 0),
    CHECK (status IN ('active', 'inactive'))
);

CREATE TABLE folding_complex (
    id bigint PRIMARY KEY USING btree,
    price numeric(10,2) NOT NULL,
    discount numeric(10,2) NOT NULL,
    final_price numeric(10,2) NOT NULL,
    code text NOT NULL UNIQUE,
    CHECK (price >= 0 AND price < 10000),
    CHECK (final_price = price - discount)
);

CREATE TABLE folding_deferrable (
    id_defer bigint PRIMARY KEY DEFERRABLE,
    email_defer text NOT NULL UNIQUE DEFERRABLE INITIALLY DEFERRED,
    email_not_defer text NOT NULL UNIQUE
);
