CREATE TABLE fk_parent (
    id bigint PRIMARY KEY
);

CREATE TABLE fk_actions (
    id bigint PRIMARY KEY,
    col_cascade bigint REFERENCES fk_parent(id) ON DELETE CASCADE NOT NULL,
    col_setnull bigint REFERENCES fk_parent(id) ON DELETE SET NULL,
    col_setdef bigint DEFAULT 1 REFERENCES fk_parent(id) ON DELETE SET DEFAULT,
    col_restrict bigint REFERENCES fk_parent(id) ON DELETE RESTRICT NOT NULL,
    col_upd_cascade bigint REFERENCES fk_parent(id) ON UPDATE CASCADE NOT NULL
);

CREATE TABLE fk_circular_a (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES fk_circular_a(id),
    other_id bigint REFERENCES fk_circular_b(id)
);

CREATE TABLE fk_circular_b (
    id bigint PRIMARY KEY,
    other_id bigint REFERENCES fk_circular_a(id)
);

CREATE TABLE fk_complex (
    id bigint NOT NULL,
    sub_id bigint NOT NULL,
    ref_id bigint REFERENCES fk_parent(id),
    match_full_id bigint REFERENCES fk_parent(id) MATCH FULL,
    match_partial_id bigint REFERENCES fk_parent(id) MATCH PARTIAL,
    PRIMARY KEY (id, sub_id)
);

ALTER TABLE fk_circular_a ADD CONSTRAINT self_fkey FOREIGN KEY (parent_id) REFERENCES fk_circular_a(id);

ALTER TABLE fk_complex ADD CONSTRAINT multi_fkey FOREIGN KEY (id, sub_id) REFERENCES fk_complex(id, sub_id);

ALTER TABLE fk_complex ADD CONSTRAINT not_valid_fkey FOREIGN KEY (ref_id) REFERENCES fk_parent(id) NOT VALID;
