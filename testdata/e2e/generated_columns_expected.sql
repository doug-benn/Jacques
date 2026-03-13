CREATE TABLE generated_test (
    id bigint PRIMARY KEY,
    price numeric(10,2) NOT NULL,
    discount numeric(10,2) DEFAULT 0,
    first_name text,
    last_name text,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW(),
    data jsonb,
    final_price numeric(10,2) GENERATED ALWAYS AS (price - discount) STORED,
    full_name text GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED,
    age interval GENERATED ALWAYS AS (updated_at - created_at) STORED,
    is_urgent boolean GENERATED ALWAYS AS (data->>'urgent' = 'true') STORED
);
