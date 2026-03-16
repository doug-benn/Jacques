CREATE TABLE noise_test (
    id bigint NOT NULL,
    val text
);

CREATE TABLE drop_test (
    id bigint PRIMARY KEY
);

CREATE TABLE string_noise (
    val text DEFAULT 'SET timeout; -- not a comment; ONLY stay'
);

DROP TABLE IF EXISTS drop_test;

DROP VIEW IF EXISTS drop_view;

CREATE VIEW drop_view AS SELECT id FROM drop_test;

CREATE FUNCTION func_with_noise() RETURNS void AS $$
BEGIN
    -- Internal comments and SET MUST be preserved
    SET statement_timeout = 3600;
    PERFORM 1;
END;
$$ LANGUAGE plpgsql;
