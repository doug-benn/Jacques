CREATE TYPE test_enum AS ENUM ('a', 'b', 'c');

CREATE DOMAIN test_domain AS text;

CREATE TABLE type_test (
    id bigint PRIMARY KEY,
    enum_val test_enum,
    arr_val text[],
    jsonb_val jsonb,
    range_val tstzrange,
    domain_val test_domain,
    uuid_val uuid,
    inet_val inet,
    cidr_val cidr,
    macaddr_val macaddr,
    collate_c text COLLATE "C" NOT NULL UNIQUE,
    collate_def text COLLATE "default",
    price numeric(10,2) NOT NULL DEFAULT 0,
    discount numeric(10,2) DEFAULT 0,
    final_price numeric(10,2) GENERATED ALWAYS AS (price - discount) STORED
);
