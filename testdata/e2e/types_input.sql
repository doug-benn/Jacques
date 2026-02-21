-- Test fixture for complex types
-- Covers: enum types, array types, JSON/JSONB

-- Enum types
CREATE TYPE order_status AS ENUM ('pending', 'processing', 'shipped', 'delivered', 'cancelled');
CREATE TYPE priority AS ENUM ('low', 'medium', 'high', 'urgent');

-- Composite type (defined but not used by tables to avoid pg-schema-diff issues)
CREATE TYPE contact_info AS (
    email text,
    phone text,
    mobile text,
    fax text
);

-- Table with enum type
CREATE TABLE public.tickets (
    id bigint NOT NULL,
    title text NOT NULL,
    status order_status NOT NULL DEFAULT 'pending',
    priority priority NOT NULL DEFAULT 'medium',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_pkey PRIMARY KEY (id);

-- Table with array types
CREATE TABLE public.projects (
    id bigint NOT NULL,
    name text NOT NULL,
    tags text[],
    milestones text[],
    budget numeric(10,2)
);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);

-- Table with JSON/JSONB
CREATE TABLE public.events (
    id bigint NOT NULL,
    name text NOT NULL,
    metadata jsonb,
    payload json,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);

-- Table with multiple complex types
CREATE TABLE public.task_assignments (
    id bigint NOT NULL,
    task_id bigint NOT NULL,
    assignees bigint[],
    status order_status NOT NULL DEFAULT 'pending',
    labels text[],
    extra_data jsonb,
    notes text
);

ALTER TABLE ONLY public.task_assignments
    ADD CONSTRAINT task_assignments_pkey PRIMARY KEY (id);
