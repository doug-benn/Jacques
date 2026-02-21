SET statement_timeout = 0;
SET lock_timeout = 0;

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text,
    created_at timestamp without time zone
);

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

ALTER TABLE public.users ALTER COLUMN id
    SET DEFAULT nextval('public.users_id_seq'::regclass);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE public.users OWNER TO testuser;

COMMENT ON TABLE public.users IS 'Application users';
