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
