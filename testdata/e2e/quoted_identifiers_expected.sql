CREATE SCHEMA "MySchema";

CREATE TABLE users (
    id bigint PRIMARY KEY,
    email text
);

CREATE TABLE order_items (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES users(id) NOT NULL,
    product_id bigint NOT NULL
);

CREATE TABLE "userProfiles" (
    id bigint PRIMARY KEY,
    "First Name" text NOT NULL,
    "Last Name" text NOT NULL
);

CREATE TABLE "MySchema".products (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE "MySchema"."Orders" (
    "OrderID" bigint NOT NULL,
    "CustomerID" bigint NOT NULL,
    "Order Date" timestamp without time zone NOT NULL,
    PRIMARY KEY ("OrderID", "CustomerID")
);
