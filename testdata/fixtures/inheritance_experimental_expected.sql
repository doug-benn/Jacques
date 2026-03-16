CREATE TABLE parent_1 (
    id bigint PRIMARY KEY
);

CREATE TABLE parent_2 (
    id bigint PRIMARY KEY
);

CREATE TABLE multi_child (
    val text
) INHERITS (parent_1, parent_2);
