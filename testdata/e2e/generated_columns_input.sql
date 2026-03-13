-- Test fixture for generated columns
-- Covers: GENERATED ALWAYS AS (expression) STORED preservation

CREATE TABLE public.generated_test (
    id bigint NOT NULL,
    price numeric(10,2) NOT NULL,
    discount numeric(10,2) DEFAULT 0,
    first_name text,
    last_name text,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW(),
    data jsonb,
    
    -- 1. Numeric expression
    final_price numeric(10,2) GENERATED ALWAYS AS (price - discount) STORED,
    
    -- 2. String concatenation
    full_name text GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED,
    
    -- 3. Date/Interval expression
    age interval GENERATED ALWAYS AS (updated_at - created_at) STORED,
    
    -- 4. JSONB extraction
    is_urgent boolean GENERATED ALWAYS AS (data->>'urgent' = 'true') STORED
);

ALTER TABLE ONLY public.generated_test ADD CONSTRAINT generated_test_pkey PRIMARY KEY (id);
