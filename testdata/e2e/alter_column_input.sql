-- Test fixture for ALTER COLUMN patterns
-- Covers: SET DEFAULT, DROP DEFAULT, SET NOT NULL, DROP NOT NULL, SET TYPE

CREATE TABLE public.users (
    id bigint NOT NULL,
    username text NOT NULL,
    email text,
    status text NOT NULL DEFAULT 'active',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

-- SET DEFAULT
ALTER TABLE public.users ALTER COLUMN created_at SET DEFAULT NOW();

-- DROP DEFAULT
ALTER TABLE public.users ALTER COLUMN status DROP DEFAULT;

-- SET NOT NULL
ALTER TABLE public.users ALTER COLUMN email SET NOT NULL;

-- DROP NOT NULL (this should pass through)
ALTER TABLE public.users ALTER COLUMN email DROP NOT NULL;

-- SET TYPE
ALTER TABLE public.users ALTER COLUMN status TYPE varchar(50);

-- Another table for more tests
CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL,
    description text,
    is_active boolean NOT NULL DEFAULT true
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

-- Multiple column alterations on same table
ALTER TABLE public.products ALTER COLUMN price TYPE integer;
ALTER TABLE public.products ALTER COLUMN description SET NOT NULL;
ALTER TABLE public.products ALTER COLUMN is_active DROP DEFAULT;

-- Table with DROP NOT NULL (passthrough)
CREATE TABLE public.logs (
    id bigint NOT NULL,
    message text NOT NULL,
    severity text
);

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (id);

ALTER TABLE public.logs ALTER COLUMN severity DROP NOT NULL;
