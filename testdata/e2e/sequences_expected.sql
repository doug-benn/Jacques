CREATE SEQUENCE shared_seq START WITH 1 INCREMENT BY 1 NO MAXVALUE CACHE 1;

CREATE SEQUENCE cached_seq CACHE 100;

CREATE TABLE shared_table_1 (
    id bigint  DEFAULT nextval('shared_seq'::regclass) PRIMARY KEY
);

CREATE TABLE shared_table_2 (
    id bigint  DEFAULT nextval('shared_seq'::regclass) PRIMARY KEY
);

CREATE TABLE serial_types (
    id_big BIGSERIAL PRIMARY KEY,
    id_reg SERIAL,
    id_small SMALLSERIAL
);

CREATE TABLE edge_case_table (
    id_start BIGSERIAL PRIMARY KEY,
    id_unowned BIGSERIAL,
    id_bounded BIGSERIAL
);

CREATE TABLE cached_table (
    id bigint  DEFAULT nextval('cached_seq') PRIMARY KEY
);
