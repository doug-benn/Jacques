CREATE TABLE public.users (
    id bigint NOT NULL,
    name text
);

ALTER TABLE public.users VALIDATE CONSTRAINT some_constraint;

ALTER TABLE public.users RENAME TO public.new_users;

ALTER TABLE public.new_users RENAME COLUMN name TO user_name;
