-- Test fixture for exclusion constraints
-- Covers: EXCLUDE USING ... folding into CREATE TABLE

CREATE TABLE exclude_test (
    id bigint NOT NULL,
    room_id bigint NOT NULL,
    booking_date date NOT NULL,
    val text
);

ALTER TABLE ONLY exclude_test ADD CONSTRAINT exclude_test_pkey PRIMARY KEY (id);

-- Simple exclusion constraint using btree (default)
ALTER TABLE ONLY exclude_test
    ADD CONSTRAINT no_double_booking EXCLUDE USING btree (room_id WITH =, booking_date WITH =);

-- Exclusion with WHERE clause
ALTER TABLE ONLY exclude_test
    ADD CONSTRAINT partial_exclude EXCLUDE USING btree (room_id WITH =) WHERE (val IS NOT NULL);
