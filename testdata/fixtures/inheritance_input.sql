-- Test fixture for multi-parent inheritance (Experimental Folding)
-- Covers: Multi-parent inheritance preservation gating

CREATE TABLE parent_1 (
    id bigint PRIMARY KEY
);

CREATE TABLE parent_2 (
    id bigint PRIMARY KEY
);

-- Multi-parent inheritance (stripped in default, preserved in experimental)
CREATE TABLE multi_child (
    val text
) INHERITS (parent_1, parent_2);
