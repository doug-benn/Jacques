CREATE TABLE column_ext_test (
    id bigint PRIMARY KEY,
    val_def text NOT NULL DEFAULT 'initial',
    val_null text NOT NULL,
    val_type varchar(100),
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);
