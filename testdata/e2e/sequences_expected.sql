CREATE SEQUENCE shared_seq START WITH 1 INCREMENT BY 1 NO MAXVALUE CACHE 1;

CREATE SEQUENCE cached_seq CACHE 100;

CREATE TABLE shared_table_1 (
    id bigint NOT NULL DEFAULT nextval('shared_seq'::regclass),
    PRIMARY KEY (id)
);

CREATE TABLE shared_table_2 (
    id bigint NOT NULL DEFAULT nextval('shared_seq'::regclass),
    PRIMARY KEY (id)
);

CREATE TABLE serial_types (
    id_big BIGSERIAL,
    id_reg SERIAL,
    id_small SMALLSERIAL,
    PRIMARY KEY (id_big)
);

CREATE TABLE edge_case_table (
    id_start BIGSERIAL,
    id_unowned BIGSERIAL,
    id_bounded BIGSERIAL,
    PRIMARY KEY (id_start)
);

CREATE TABLE cached_table (
    id bigint NOT NULL DEFAULT nextval('cached_seq'),
    PRIMARY KEY (id)
);
