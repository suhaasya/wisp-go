CREATE SCHEMA IF NOT EXISTS wisp_meta;

CREATE TABLE wisp_meta.parent (
    id bigint PRIMARY KEY,
    status text NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'done'))
);

CREATE TABLE wisp_meta.child (
    id bigint PRIMARY KEY REFERENCES wisp_meta.parent (id),
    note text
);

CREATE INDEX idx_parent_status ON wisp_meta.parent (status);

CREATE VIEW wisp_meta.parent_v AS SELECT id, status FROM wisp_meta.parent;

CREATE SEQUENCE wisp_meta.order_seq;

CREATE TYPE wisp_meta.order_status AS ENUM ('pending', 'paid');

CREATE OR REPLACE FUNCTION wisp_meta.echo_id(x bigint) RETURNS bigint
LANGUAGE sql IMMUTABLE AS $$ SELECT x $$;

CREATE OR REPLACE FUNCTION wisp_meta.touch_parent() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RETURN NEW;
END;
$$;

CREATE TABLE wisp_meta.parted (
    id bigint,
    created_at date NOT NULL
) PARTITION BY RANGE (created_at);

CREATE TABLE wisp_meta.parted_2024 PARTITION OF wisp_meta.parted
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TRIGGER parent_touch
    BEFORE UPDATE ON wisp_meta.parent
    FOR EACH ROW EXECUTE PROCEDURE wisp_meta.touch_parent();
