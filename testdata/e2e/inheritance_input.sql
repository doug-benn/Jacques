-- Test fixture for basic table inheritance
-- Covers: Simple single-parent inheritance, ONLY removal in INHERITS

CREATE TABLE parent_table (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

-- Basic inheritance
CREATE TABLE child_table (
    val text
) INHERITS (parent_table);

-- ONLY removal in INHERITS
CREATE TABLE only_child (
    val text
) INHERITS (ONLY parent_table);
