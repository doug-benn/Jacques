CREATE TABLE public.employees (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL UNIQUE,
    manager_id bigint REFERENCES public.employees(id),
    department text,
    hire_date date NOT NULL
);

CREATE TABLE public.categories (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    parent_id bigint REFERENCES public.categories(id) ON DELETE CASCADE,
    description text
);

CREATE TABLE public.documents (
    id bigint PRIMARY KEY,
    title text NOT NULL,
    version integer NOT NULL DEFAULT 1,
    parent_id bigint REFERENCES public.documents(id) ON DELETE SET NULL,
    content text,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE public.employees
    ADD CONSTRAINT employees_manager_fkey FOREIGN KEY (manager_id) REFERENCES public.employees(id);

ALTER TABLE public.categories
    ADD CONSTRAINT categories_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.categories(id) ON DELETE CASCADE;

ALTER TABLE public.documents
    ADD CONSTRAINT documents_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.documents(id) ON DELETE SET NULL;
