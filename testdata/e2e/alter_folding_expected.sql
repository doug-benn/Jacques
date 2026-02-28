CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text,
    name text NOT NULL
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    sku text NOT NULL UNIQUE,
    name text NOT NULL
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    order_number text NOT NULL,
    UNIQUE (user_id, order_number)
);

CREATE TABLE public.accounts (
    id bigint PRIMARY KEY,
    balance numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'active',
    CONSTRAINT accounts_balance_check CHECK (balance >= 0),
    CONSTRAINT accounts_status_check CHECK (status IN ('active', 'inactive', 'suspended'))
);

CREATE TABLE public.inventories (
    id bigint PRIMARY KEY,
    product_id bigint NOT NULL,
    quantity integer NOT NULL,
    reorder_point integer NOT NULL,
    CONSTRAINT inventories_quantity_check CHECK (quantity >= 0),
    CONSTRAINT inventories_reorder_check CHECK (reorder_point >= 0),
    CONSTRAINT inventories_reorder_quantity_check CHECK (reorder_point <= quantity)
);

CREATE TABLE public.inline_notnull (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL UNIQUE
);

CREATE TABLE public.complex_check (
    id bigint PRIMARY KEY,
    price numeric(10,2) NOT NULL,
    discount numeric(10,2) NOT NULL,
    final_price numeric(10,2) NOT NULL,
    CONSTRAINT complex_check_price_check CHECK (price >= 0 AND price < 10000),
    CONSTRAINT complex_check_final_check CHECK (final_price = price - discount)
);

CREATE TABLE public.multi_constraint (
    id bigint PRIMARY KEY,
    code text NOT NULL UNIQUE
);

CREATE TABLE public.deferrable_unique (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE
);

CREATE TABLE public.using_clause (
    id bigint PRIMARY KEY
);
