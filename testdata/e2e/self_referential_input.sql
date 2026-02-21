-- Test fixture for self-referential tables
-- Covers: adjacency list (manager_id -> employee.id), tree structures

-- Employee table with self-referential FK
CREATE TABLE public.employees (
    id bigint NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    manager_id bigint,
    department text,
    hire_date date NOT NULL
);

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_email_key UNIQUE (email);

-- Self-referential FK (employee references itself)
ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_manager_fkey FOREIGN KEY (manager_id) REFERENCES public.employees(id);

-- Category table with self-reference for hierarchy
CREATE TABLE public.categories (
    id bigint NOT NULL,
    name text NOT NULL,
    parent_id bigint,
    description text
);

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);

-- Self-referential FK with ON DELETE CASCADE (for tree structures)
ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.categories(id) ON DELETE CASCADE;

-- Document table with version history (self-reference)
CREATE TABLE public.documents (
    id bigint NOT NULL,
    title text NOT NULL,
    version integer NOT NULL DEFAULT 1,
    parent_id bigint,
    content text,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);

-- Self-referential FK with ON DELETE SET NULL
ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.documents(id) ON DELETE SET NULL;
