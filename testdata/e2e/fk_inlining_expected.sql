CREATE TABLE public.categories (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    category_id bigint REFERENCES public.categories(id) NOT NULL,
    name text NOT NULL
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

CREATE TABLE public.order_items (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES public.orders(id) ON DELETE CASCADE NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL
);

CREATE TABLE public.accounts (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES public.accounts(id) ON UPDATE SET NULL
);

CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL
);

CREATE TABLE public.comments (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) NOT NULL,
    content text NOT NULL
);

CREATE TABLE public.order_shipments (
    order_id bigint NOT NULL,
    shipment_id bigint NOT NULL,
    tracking_number text,
    PRIMARY KEY (order_id, shipment_id)
);

CREATE TABLE public.queues (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    parent_id bigint REFERENCES queues(id)
);

CREATE TABLE public.set_default_child (
    id bigint PRIMARY KEY,
    parent_id bigint DEFAULT 1 REFERENCES public.set_default_child(id) ON DELETE SET DEFAULT
);

CREATE TABLE public.set_null_child (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES public.set_null_child(id) ON DELETE SET NULL
);

CREATE TABLE public.employee_hierarchy (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    manager_id bigint REFERENCES public.employee_hierarchy(id)
);

CREATE TABLE public.match_full_a (
    id bigint PRIMARY KEY
);

CREATE TABLE public.match_full_b (
    id bigint PRIMARY KEY,
    ref_id bigint REFERENCES public.match_full_a(id) MATCH FULL NOT NULL
);

CREATE TABLE public.match_partial_a (
    id bigint PRIMARY KEY
);

CREATE TABLE public.match_partial_b (
    id bigint PRIMARY KEY,
    ref_id bigint REFERENCES public.match_partial_a(id) MATCH PARTIAL
);

CREATE TABLE public.no_action_child (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES public.no_action_child(id) ON DELETE NO ACTION
);

CREATE TABLE public.departments (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    manager_id bigint REFERENCES public.employees(id)
);

CREATE TABLE public.employees (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    department_id bigint REFERENCES public.departments(id)
);

ALTER TABLE public.accounts
    ADD CONSTRAINT accounts_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.accounts(id) ON UPDATE SET NULL;

ALTER TABLE public.order_shipments
    ADD CONSTRAINT order_shipments_order_fkey FOREIGN KEY (order_id, shipment_id) REFERENCES public.orders(id, id);

ALTER TABLE public.queues
    ADD CONSTRAINT queues_parent_fkey FOREIGN KEY (parent_id) REFERENCES queues(id) NOT VALID;

ALTER TABLE public.set_default_child
    ADD CONSTRAINT set_default_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.set_default_child(id) ON DELETE SET DEFAULT;

ALTER TABLE public.set_null_child
    ADD CONSTRAINT set_null_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.set_null_child(id) ON DELETE SET NULL;

ALTER TABLE public.employee_hierarchy
    ADD CONSTRAINT employee_hierarchy_manager_fkey FOREIGN KEY (manager_id) REFERENCES public.employee_hierarchy(id);

ALTER TABLE public.no_action_child
    ADD CONSTRAINT no_action_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.no_action_child(id) ON DELETE NO ACTION;
