CREATE TABLE exclude_test (
    id bigint PRIMARY KEY,
    room_id bigint NOT NULL,
    booking_date date NOT NULL,
    val text,
    CONSTRAINT no_double_booking EXCLUDE USING btree (room_id WITH =, booking_date WITH =),
    CONSTRAINT partial_exclude EXCLUDE USING btree (room_id WITH =) WHERE (val IS NOT NULL)
);
