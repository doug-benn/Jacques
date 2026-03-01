CREATE TABLE public.customers (
    id bigint NOT NULL,
    name text NOT NULL,
    contact contact_info,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.users (
    id bigint NOT NULL,
    email public.email NOT NULL,
    phone public.phone_number,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    quantity public.positive_int NOT NULL,
    status public.order_status NOT NULL DEFAULT 'pending'
);
