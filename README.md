# Jacques - Schema decontamination

![Jacques decontaminating](/web/jacques_he_is_clean.gif)

_"Voilà. He is clean."_ - Jacques (Finding Nemo)

Jacques decontaminates database schema dumps. Cleaning, folding and consolidating schemas for use in migrations, sqlc, ORMs, and version control.

Currently supports **PostgreSQL**. More databases coming soon.

## Features

- Removes noise: comments, `GRANT`, `REVOKE`, `OWNER TO`, `SET`, session config
- Consolidates `CREATE TABLE` + `ALTER TABLE` statements into self-contained definitions
- Inlines foreign key constraints into column definitions
- Converts sequences to `SERIAL`/`BIGSERIAL` types
- Handles database-specific types: enums, domains, generated columns
- Guaranteed semantic equivalence - cleaned output produces identical schema

## Why?

Database dump tools like `pg_dump` intentionally split table definitions across multiple statements for dependency ordering, kinda ugly and not very readable. Jacques folds them back together into a clean, consolidated and readable schema that's perfect for:

- Migration files
- An sqlc schema
- ORM code generation
- Documentation

## Experimental Folding

A CLI flag to enable foldering transformations that can't be end-to-end tested. Because the semantic equivalence can't be tested these schema transformations are opt-in at your own risk - Please check the output carefully!

## Quick Start

```bash
# Install
go install github.com/doug-benn/Jacques@latest

# Use with file
jacques -i schema.sql -o cleaned.sql

# Or pipe through stdin/stdout
pg_dump --schema-only mydb | jacques > cleaned.sql
```

## Example

**Before** (pg_dump output):

```sql
CREATE TABLE users (
    id bigint NOT NULL,
    email text
);

CREATE SEQUENCE users_id_seq;

ALTER SEQUENCE users_id_seq OWNED BY users.id;

ALTER TABLE users ALTER COLUMN id SET DEFAULT nextval('users_id_seq');

ALTER TABLE ONLY users ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY users ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE users OWNER TO admin;
```

**After** (Jacques output):

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email text UNIQUE
);
```

## CLI Options

| Flag             | Description                      | Default |
| ---------------- | -------------------------------- | ------- |
| `-i`, `--input`  | Input SQL file (`-` for stdin)   | stdin   |
| `-o`, `--output` | Output SQL file (`-` for stdout) | stdout  |
