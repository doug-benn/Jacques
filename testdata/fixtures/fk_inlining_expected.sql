CREATE TABLE public.match_full_a (
    id bigint PRIMARY KEY
);

CREATE TABLE public.match_full_b (
    id bigint PRIMARY KEY,
    ref_id bigint NOT NULL
);

CREATE TABLE public.match_partial_a (
    id bigint PRIMARY KEY
);

CREATE TABLE public.match_partial_b (
    id bigint PRIMARY KEY,
    ref_id bigint
);

ALTER TABLE ONLY public.match_full_b
    ADD CONSTRAINT match_full_fkey FOREIGN KEY (ref_id) REFERENCES public.match_full_a(id) MATCH FULL;

ALTER TABLE ONLY public.match_partial_b
    ADD CONSTRAINT match_partial_fkey FOREIGN KEY (ref_id) REFERENCES public.match_partial_a(id) MATCH PARTIAL;
