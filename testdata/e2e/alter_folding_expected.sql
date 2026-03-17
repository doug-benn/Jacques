CREATE TABLE folding_basics (
    id bigint PRIMARY KEY,
    sku text NOT NULL UNIQUE,
    user_id bigint NOT NULL,
    order_number text NOT NULL,
    balance numeric(10,2) NOT NULL,
    id_defer bigint PRIMARY KEY DEFERRABLE,
    email_defer text NOT NULL UNIQUE DEFERRABLE INITIALLY DEFERRED,
    UNIQUE (user_id, order_number),
    CHECK (balance >= 0)
);

CREATE TABLE folding_exclude (
    id bigint PRIMARY KEY USING btree,
    room_id bigint NOT NULL,
    booking_date date NOT NULL,
    val text,
    CONSTRAINT no_double_booking EXCLUDE USING btree (room_id WITH =, booking_date WITH =),
    CONSTRAINT partial_exclude EXCLUDE USING btree (room_id WITH =) WHERE (val IS NOT NULL)
);

CREATE TABLE fk_parent (
    id bigint PRIMARY KEY
);

CREATE TABLE fk_child (
    id bigint PRIMARY KEY REFERENCES fk_child(id),
    parent_id bigint REFERENCES fk_parent(id),
    col_cascade bigint REFERENCES fk_parent(id) ON DELETE CASCADE NOT NULL,
    col_match_full bigint REFERENCES fk_parent(id) MATCH FULL,
    col_set_null bigint REFERENCES fk_parent(id) ON DELETE SET NULL ON UPDATE SET NULL,
    col_set_default bigint REFERENCES fk_parent(id) ON DELETE SET DEFAULT ON UPDATE SET DEFAULT
);

CREATE TABLE parent_table (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE child_table (
    val text
) INHERITS (parent_table);

CREATE TABLE only_child (
    val text
) INHERITS (parent_table);

CREATE TABLE info (
    id bigint PRIMARY KEY,
    source bigint,
    destination bigint REFERENCES info(id)
);

ALTER TABLE info ADD CONSTRAINT "self ref for source" FOREIGN KEY (source) REFERENCES info(id) NOT VALID;
