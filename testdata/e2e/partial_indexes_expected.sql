CREATE TABLE index_test (
    id bigint PRIMARY KEY,
    val text NOT NULL UNIQUE,
    status text NOT NULL,
    data jsonb
);

CREATE INDEX idx_partial ON index_test(status) WHERE status = 'active';

CREATE INDEX idx_include ON index_test(status) INCLUDE (data);

CREATE INDEX idx_expr ON index_test((lower(val)));

CREATE UNIQUE INDEX idx_val_include ON index_test(val) INCLUDE (status);

CREATE INDEX idx_desc ON index_test(val DESC);

CREATE INDEX idx_nulls_first ON index_test(val NULLS FIRST);

CREATE INDEX idx_collate ON index_test(val COLLATE "en_US");
