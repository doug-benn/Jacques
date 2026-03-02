CREATE TABLE categories (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE products (
    id bigint PRIMARY KEY,
    category_id bigint REFERENCES categories(id) NOT NULL,
    name text NOT NULL
);

CREATE TABLE orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

CREATE TABLE order_items (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES orders(id) ON DELETE CASCADE NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL
);

CREATE TABLE accounts (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES accounts(id) ON UPDATE SET NULL
);

CREATE TABLE users (
    id bigint PRIMARY KEY,
    email text NOT NULL
);

CREATE TABLE comments (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES users(id) NOT NULL,
    content text NOT NULL
);

CREATE TABLE order_shipments (
    order_id bigint NOT NULL,
    shipment_id bigint NOT NULL,
    tracking_number text,
    PRIMARY KEY (order_id, shipment_id)
);

CREATE TABLE queues (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    parent_id bigint REFERENCES queues(id)
);

CREATE TABLE set_default_child (
    id bigint PRIMARY KEY,
    parent_id bigint DEFAULT 1 REFERENCES set_default_child(id) ON DELETE SET DEFAULT
);

CREATE TABLE set_null_child (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES set_null_child(id) ON DELETE SET NULL
);

CREATE TABLE employee_hierarchy (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    manager_id bigint REFERENCES employee_hierarchy(id)
);

CREATE TABLE departments (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    manager_id bigint REFERENCES employees(id)
);

CREATE TABLE employees (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    department_id bigint REFERENCES departments(id)
);

CREATE TABLE no_action_child (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES no_action_child(id) ON DELETE NO ACTION
);

CREATE TABLE categories2 (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES categories2(id) ON DELETE RESTRICT,
    name text NOT NULL
);

CREATE TABLE order_items2 (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES orders(id) ON UPDATE CASCADE NOT NULL,
    product_id bigint NOT NULL
);

CREATE TABLE match_full_a (
    id bigint PRIMARY KEY
);

CREATE TABLE match_full_b (
    id bigint PRIMARY KEY,
    ref_id bigint REFERENCES match_full_a(id) MATCH FULL NOT NULL
);

CREATE TABLE match_partial_a (
    id bigint PRIMARY KEY
);

CREATE TABLE match_partial_b (
    id bigint PRIMARY KEY,
    ref_id bigint REFERENCES match_partial_a(id) MATCH PARTIAL
);

ALTER TABLE accounts
    ADD CONSTRAINT accounts_parent_fkey FOREIGN KEY (parent_id) REFERENCES accounts(id) ON UPDATE SET NULL;

ALTER TABLE order_shipments
    ADD CONSTRAINT order_shipments_order_fkey FOREIGN KEY (order_id, shipment_id) REFERENCES orders(id, id);

ALTER TABLE queues
    ADD CONSTRAINT queues_parent_fkey FOREIGN KEY (parent_id) REFERENCES queues(id) NOT VALID;

ALTER TABLE set_default_child
    ADD CONSTRAINT set_default_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES set_default_child(id) ON DELETE SET DEFAULT;

ALTER TABLE set_null_child
    ADD CONSTRAINT set_null_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES set_null_child(id) ON DELETE SET NULL;

ALTER TABLE employee_hierarchy
    ADD CONSTRAINT employee_hierarchy_manager_fkey FOREIGN KEY (manager_id) REFERENCES employee_hierarchy(id);

ALTER TABLE no_action_child
    ADD CONSTRAINT no_action_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES no_action_child(id) ON DELETE NO ACTION;

ALTER TABLE categories2
    ADD CONSTRAINT categories2_parent_fkey FOREIGN KEY (parent_id) REFERENCES categories2(id) ON DELETE RESTRICT;
