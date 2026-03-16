CREATE TYPE address_info AS (
    city text,
    street text
);

CREATE DOMAIN positive_val AS integer
CHECK (VALUE > 0);

CREATE TABLE custom_type_test (
    id bigint PRIMARY KEY,
    addr address_info,
    qty positive_val
);
