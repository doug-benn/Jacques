-- Test fixture for custom types (gated - requires ExperimentalFolding)
-- Covers: COMPOSITE types, DOMAIN with CHECK constraints

-- 1. COMPOSITE type (Gated)
CREATE TYPE address_info AS (
    city text,
    street text
);

-- 2. DOMAIN with CHECK (Gated)
CREATE DOMAIN public.positive_val AS integer
CHECK (VALUE > 0);

-- Table using gated types
CREATE TABLE public.custom_type_test (
    id bigint NOT NULL,
    addr address_info,
    qty public.positive_val
);

ALTER TABLE ONLY public.custom_type_test ADD PRIMARY KEY (id);
