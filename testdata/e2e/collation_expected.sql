CREATE TABLE collation_test (
    id bigint PRIMARY KEY,
    name_c text COLLATE "C" NOT NULL UNIQUE,
    name_default text COLLATE "default",
    name_none text
);
