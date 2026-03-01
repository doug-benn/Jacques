-- Test fixture for FK MATCH FULL/PARTIAL (gated - requires ExperimentalFolding)
-- These transformations are NOT supported by pg-schema-diff, so they're gated
-- Note: Simple FK inlining is now default behavior (E2E tested)

-- Negative test: FK with MATCH FULL (cannot inline - passes through)
CREATE TABLE public.match_full_a (
    id bigint NOT NULL
);

ALTER TABLE ONLY public.match_full_a ADD CONSTRAINT match_full_a_pkey PRIMARY KEY (id);

CREATE TABLE public.match_full_b (
    id bigint NOT NULL,
    ref_id bigint NOT NULL
);

ALTER TABLE ONLY public.match_full_b ADD CONSTRAINT match_full_b_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.match_full_b
    ADD CONSTRAINT match_full_fkey FOREIGN KEY (ref_id) REFERENCES public.match_full_a(id) MATCH FULL;

-- Negative test: FK with MATCH PARTIAL (cannot inline - passes through)
CREATE TABLE public.match_partial_a (
    id bigint NOT NULL
);

ALTER TABLE ONLY public.match_partial_a ADD CONSTRAINT match_partial_a_pkey PRIMARY KEY (id);

CREATE TABLE public.match_partial_b (
    id bigint NOT NULL,
    ref_id bigint
);

ALTER TABLE ONLY public.match_partial_b ADD CONSTRAINT match_partial_b_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.match_partial_b
    ADD CONSTRAINT match_partial_fkey FOREIGN KEY (ref_id) REFERENCES public.match_partial_a(id) MATCH PARTIAL;
