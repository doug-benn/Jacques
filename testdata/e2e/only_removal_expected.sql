CREATE TABLE public.users (
    id bigint NOT NULL,
    name text,
    new_col int
);

ALTER TABLE public.users RENAME TO new_users;

ALTER TABLE public.new_users RENAME COLUMN name TO user_name;
