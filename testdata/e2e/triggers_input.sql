-- Test fixture for triggers
-- Covers: Trigger function preservation and CREATE TRIGGER inlining/pass-through

CREATE TABLE trigger_test (
    id bigint NOT NULL,
    val text,
    updated_at timestamp without time zone
);

ALTER TABLE ONLY trigger_test ADD PRIMARY KEY (id);

-- Trigger Function
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- BEFORE UPDATE Trigger
CREATE TRIGGER trg_before_update
    BEFORE UPDATE ON trigger_test
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

-- AFTER INSERT Trigger
CREATE OR REPLACE FUNCTION log_insert()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO log_table(msg) VALUES ('inserted ' || NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE log_table (
    msg text
);

CREATE TRIGGER trg_after_insert
    AFTER INSERT ON trigger_test
    FOR EACH ROW
    EXECUTE FUNCTION log_insert();
