CREATE TABLE trigger_test (
    id bigint PRIMARY KEY,
    val text,
    updated_at timestamp without time zone
);

CREATE TABLE log_table (
    msg text
);

CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_before_update
    BEFORE UPDATE ON trigger_test
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

CREATE OR REPLACE FUNCTION log_insert()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO log_table(msg) VALUES ('inserted ' || NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_after_insert
    AFTER INSERT ON trigger_test
    FOR EACH ROW
    EXECUTE FUNCTION log_insert();
