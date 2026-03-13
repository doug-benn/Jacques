CREATE SCHEMA app;

CREATE SCHEMA "MySchema";

CREATE TABLE base_table (
    id bigint PRIMARY KEY
);

CREATE TABLE app.app_table (
    id bigint PRIMARY KEY,
    base_id bigint REFERENCES base_table(id)
);

CREATE TABLE "MySchema".profile (
    id bigint PRIMARY KEY,
    bio text
);
