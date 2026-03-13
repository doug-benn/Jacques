-- Test fixture for types and advanced column properties
-- Covers: Enum types, array types, JSON/JSONB, range types, basic DOMAIN types,
--         Generated columns, and Collation clauses.

-- 1. Types & Domains
CREATE TYPE test_enum AS ENUM ('a', 'b', 'c');
CREATE DOMAIN public.test_domain AS text;

-- Consolidated table for type and column testing
CREATE TABLE public.type_test (
    id bigint NOT NULL,
    -- Basic types
    enum_val test_enum,
    arr_val text[],
    jsonb_val jsonb,
    range_val tstzrange,
    domain_val public.test_domain,
    
    -- 2. Collation
    collate_c text COLLATE "C" NOT NULL,
    collate_def text COLLATE "default",
    
    -- 3. Generated Columns
    price numeric(10,2) NOT NULL DEFAULT 0,
    discount numeric(10,2) DEFAULT 0,
    final_price numeric(10,2) GENERATED ALWAYS AS (price - discount) STORED
);

ALTER TABLE ONLY public.type_test ADD PRIMARY KEY (id);
ALTER TABLE ONLY public.type_test ADD CONSTRAINT type_test_collate_key UNIQUE (collate_c);
