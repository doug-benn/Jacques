-- Test fixture for types
-- Covers: Enum types, array types, JSON/JSONB, range types, basic DOMAIN types

-- 1. Enum Types
CREATE TYPE test_enum AS ENUM ('a', 'b', 'c');

-- 2. Basic DOMAIN Types (without CHECK)
CREATE DOMAIN public.test_domain AS text;

-- Consolidated table for type testing
CREATE TABLE public.type_test (
    id bigint NOT NULL,
    -- Enum and Array
    enum_val test_enum,
    arr_val text[],
    -- JSON/JSONB
    json_val json,
    jsonb_val jsonb,
    -- Range types
    range_val tstzrange,
    int_range int4range,
    -- Domain type
    domain_val public.test_domain
);

ALTER TABLE ONLY public.type_test ADD PRIMARY KEY (id);
