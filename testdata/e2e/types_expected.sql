CREATE TYPE order_status AS ENUM ('pending', 'processing', 'shipped', 'delivered', 'cancelled');

CREATE TYPE priority AS ENUM ('low', 'medium', 'high', 'urgent');

CREATE TYPE contact_info AS (
    email text,
    phone text,
    mobile text,
    fax text
);

CREATE TABLE public.tickets (
    id bigint PRIMARY KEY,
    title text NOT NULL,
    status order_status NOT NULL DEFAULT 'pending',
    priority priority NOT NULL DEFAULT 'medium',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.projects (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    tags text[],
    milestones text[],
    budget numeric(10,2)
);

CREATE TABLE public.events (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    metadata jsonb,
    payload json,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.task_assignments (
    id bigint PRIMARY KEY,
    task_id bigint NOT NULL,
    assignees bigint[],
    status order_status NOT NULL DEFAULT 'pending',
    labels text[],
    extra_data jsonb,
    notes text
);

CREATE TABLE public.customers (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    contact contact_info,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);
