-- PostgreSQL: creates 500 tables in wisp_meta for list timing tests.
CREATE SCHEMA IF NOT EXISTS wisp_meta;

DO $$
BEGIN
    FOR i IN 1..500 LOOP
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS wisp_meta.t_%s (id bigint PRIMARY KEY)',
            lpad(i::text, 3, '0')
        );
    END LOOP;
END;
$$;
