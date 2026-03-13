CREATE TYPE test_enum AS ENUM ('a', 'b', 'c');

CREATE DOMAIN test_domain AS text;

CREATE TABLE type_test (
    id bigint PRIMARY KEY,
    enum_val test_enum,
    arr_val text[],
    json_val json,
    jsonb_val jsonb,
    range_val tstzrange,
    int_range int4range,
    domain_val test_domain
);
