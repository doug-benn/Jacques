-- Test fixture for ALTER COLUMN patterns
-- Covers: SET DEFAULT, DROP DEFAULT, SET NOT NULL, DROP NOT NULL, SET TYPE

CREATE TABLE column_ext_test (
    id bigint NOT NULL,
    val_def text NOT NULL DEFAULT 'initial',
    val_null text,
    val_type text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY column_ext_test ADD CONSTRAINT column_ext_test_pkey PRIMARY KEY (id);

-- SET DEFAULT
ALTER TABLE column_ext_test ALTER COLUMN created_at SET DEFAULT NOW();

-- DROP DEFAULT
ALTER TABLE column_ext_test ALTER COLUMN val_def DROP DEFAULT;

-- SET NOT NULL
ALTER TABLE column_ext_test ALTER COLUMN val_null SET NOT NULL;

-- DROP NOT NULL (this should pass through as a separate ALTER)
ALTER TABLE column_ext_test ALTER COLUMN val_type DROP NOT NULL;

-- SET TYPE
ALTER TABLE column_ext_test ALTER COLUMN val_type TYPE varchar(100);
