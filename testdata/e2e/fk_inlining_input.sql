-- Test fixture for FK inlining (E2E testable)
-- Covers: Simple FK inlining, ON DELETE CASCADE/SET NULL/SET DEFAULT,
-- self-referential FK, multi-column FK, NOT VALID FK, circular dependencies
-- Note: MATCH FULL/PARTIAL requires ExperimentalFolding

-- Simple FK (should inline)
CREATE TABLE public.categories (
    id bigint NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);

CREATE TABLE public.products (
    id bigint NOT NULL,
    category_id bigint NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_category_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id);

-- FK with ON DELETE CASCADE (should inline)
CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

CREATE TABLE public.order_items (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL
);

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;

-- FK with ON UPDATE SET NULL (should inline)
CREATE TABLE public.accounts (
    id bigint NOT NULL,
    parent_id bigint
);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.accounts(id) ON UPDATE SET NULL;

-- FK with quoted constraint name (should inline)
CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE TABLE public.comments (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    content text NOT NULL
);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT "comments user fk" FOREIGN KEY (user_id) REFERENCES public.users(id);

-- Multi-column FK (cannot inline - passes through as ALTER TABLE)
CREATE TABLE public.order_shipments (
    order_id bigint NOT NULL,
    shipment_id bigint NOT NULL,
    tracking_number text
);

ALTER TABLE ONLY public.order_shipments
    ADD CONSTRAINT order_shipments_pkey PRIMARY KEY (order_id, shipment_id);

ALTER TABLE ONLY public.order_shipments
    ADD CONSTRAINT order_shipments_order_fkey FOREIGN KEY (order_id, shipment_id) REFERENCES public.orders(id, id);

-- FK with NOT VALID (cannot inline - passes through)
CREATE TABLE public.queues (
    id bigint NOT NULL,
    name text NOT NULL,
    parent_id bigint
);

ALTER TABLE ONLY public.queues
    ADD CONSTRAINT queues_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.queues
    ADD CONSTRAINT queues_parent_fkey FOREIGN KEY (parent_id) REFERENCES queues(id) NOT VALID;

-- Edge case: FK with ON DELETE SET DEFAULT (should inline)
CREATE TABLE public.set_default_child (
    id bigint NOT NULL,
    parent_id bigint DEFAULT 1
);

ALTER TABLE ONLY public.set_default_child
    ADD CONSTRAINT set_default_child_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.set_default_child
    ADD CONSTRAINT set_default_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.set_default_child(id) ON DELETE SET DEFAULT;

-- Edge case: FK with ON DELETE SET NULL (should inline)
CREATE TABLE public.set_null_child (
    id bigint NOT NULL,
    parent_id bigint
);

ALTER TABLE ONLY public.set_null_child
    ADD CONSTRAINT set_null_child_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.set_null_child
    ADD CONSTRAINT set_null_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.set_null_child(id) ON DELETE SET NULL;

-- Edge case: Self-referential FK (should inline)
CREATE TABLE public.employee_hierarchy (
    id bigint NOT NULL,
    name text NOT NULL,
    manager_id bigint
);

ALTER TABLE ONLY public.employee_hierarchy
    ADD CONSTRAINT employee_hierarchy_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.employee_hierarchy
    ADD CONSTRAINT employee_hierarchy_manager_fkey FOREIGN KEY (manager_id) REFERENCES public.employee_hierarchy(id);

-- Edge case: Circular FK - employees reference departments, departments reference employees
CREATE TABLE public.departments (
    id bigint NOT NULL,
    name text NOT NULL,
    manager_id bigint
);

ALTER TABLE ONLY public.departments ADD CONSTRAINT departments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_manager_fkey FOREIGN KEY (manager_id) REFERENCES public.employees(id);

CREATE TABLE public.employees (
    id bigint NOT NULL,
    name text NOT NULL,
    department_id bigint
);

ALTER TABLE ONLY public.employees ADD CONSTRAINT employees_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_department_fkey FOREIGN KEY (department_id) REFERENCES public.departments(id);

-- Negative test: FK with ON DELETE NO ACTION (should pass through)
CREATE TABLE public.no_action_child (
    id bigint NOT NULL,
    parent_id bigint
);

ALTER TABLE ONLY public.no_action_child ADD CONSTRAINT no_action_child_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.no_action_child
    ADD CONSTRAINT no_action_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.no_action_child(id) ON DELETE NO ACTION;
