CREATE SCHEMA app;

CREATE SCHEMA "MySchema";

CREATE TABLE countries (
    id bigint PRIMARY KEY,
    code text NOT NULL,
    name text NOT NULL
);

CREATE TABLE app.users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    country_id bigint REFERENCES countries(id)
);

CREATE TABLE "MySchema".profile (
    id bigint PRIMARY KEY,
    bio text
);