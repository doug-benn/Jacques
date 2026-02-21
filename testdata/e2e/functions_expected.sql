CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    price numeric(10,2) NOT NULL,
    category text NOT NULL
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

CREATE OR REPLACE FUNCTION public.get_user_by_id(user_id bigint)
RETURNS TABLE(id bigint, email text, name text) AS $$
BEGIN
    RETURN QUERY
    SELECT u.id, u.email, u.name
    FROM public.users u
    WHERE u.id = user_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION public.get_user_email(user_id bigint)
RETURNS text AS $$
DECLARE
    email_text text;
BEGIN
    SELECT u.email INTO email_text
    FROM public.users u
    WHERE u.id = user_id;
    RETURN email_text;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE PROCEDURE public.create_user(
    p_email text,
    p_name text
) AS $$
BEGIN
    INSERT INTO public.users (email, name)
    VALUES (p_email, p_name);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION public.calculate_discount(price numeric, discount_percent numeric)
RETURNS numeric AS $$
BEGIN
    RETURN price * (1 - discount_percent / 100);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.full_name(first_name text, last_name text)
RETURNS text AS $$
BEGIN
    RETURN first_name || ' ' || last_name;
END;
$$ LANGUAGE SQL;

CREATE OR REPLACE FUNCTION public.get_products_by_category(p_category text)
RETURNS SETOF public.products AS $$
BEGIN
    RETURN QUERY
    SELECT p.*
    FROM public.products p
    WHERE p.category = p_category;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION public.sum_order_totals()
RETURNS numeric AS $$
DECLARE
    total numeric := 0;
BEGIN
    SELECT COALESCE(SUM(total), 0) INTO total
    FROM public.orders;
    RETURN total;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION public.update_order_total()
RETURNS TRIGGER AS $$
BEGIN
    NEW.total := NEW.total * 1.1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION public.is_admin(user_role text)
RETURNS boolean AS $$
BEGIN
    RETURN user_role = 'admin';
END;
$$ LANGUAGE SQL IMMUTABLE;
