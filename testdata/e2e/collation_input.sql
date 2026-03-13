-- Test fixture for collation
-- Covers: COLLATE clause preservation in column definitions

CREATE TABLE collation_test (
    id bigint NOT NULL,
    name_c text COLLATE "C" NOT NULL,
    name_default text COLLATE "default",
    name_none text
);

ALTER TABLE ONLY collation_test ADD PRIMARY KEY (id);
ALTER TABLE ONLY collation_test ADD CONSTRAINT collation_test_name_c_key UNIQUE (name_c);
