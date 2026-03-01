CREATE TABLE public.match_full_a (
    id bigint PRIMARY KEY
);

CREATE TABLE public.match_full_b (
    id bigint PRIMARY KEY,
    ref_id bigint REFERENCES public.match_full_a(id) MATCH FULL NOT NULL
);

CREATE TABLE public.match_partial_a (
    id bigint PRIMARY KEY
);

CREATE TABLE public.match_partial_b (
    id bigint PRIMARY KEY,
    ref_id bigint REFERENCES public.match_partial_a(id) MATCH PARTIAL
);
