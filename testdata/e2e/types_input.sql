-- Test fixture for types (E2E testable)
-- Covers: enum types, array types, JSON/JSONB, range types
-- Note: COMPOSITE types and DOMAIN with CHECK are gated under ExperimentalFolding

-- Enum types
CREATE TYPE order_status AS ENUM ('pending', 'processing', 'shipped', 'delivered', 'cancelled');
CREATE TYPE priority AS ENUM ('low', 'medium', 'high', 'urgent');

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

-- Range types
CREATE TABLE public.reservations (
    id bigint NOT NULL,
    room_number text NOT NULL,
    stay_period tstzrange NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_pkey PRIMARY KEY (id);

CREATE TABLE public.inventory (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity_range int4range NOT NULL,
    last_updated timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.inventory
    ADD CONSTRAINT inventory_pkey PRIMARY KEY (id);

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    birth_date date,
    age_range int4range,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE TABLE public.order_versions (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    status_range daterange NOT NULL,
    status order_status NOT NULL
);

ALTER TABLE ONLY public.order_versions
    ADD CONSTRAINT order_versions_pkey PRIMARY KEY (id);
