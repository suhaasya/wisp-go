CREATE DATABASE IF NOT EXISTS wisp_meta;
USE wisp_meta;

CREATE TABLE parent (
    id BIGINT PRIMARY KEY,
    status VARCHAR(32) NOT NULL DEFAULT 'new'
);

CREATE TABLE child (
    id BIGINT PRIMARY KEY,
    parent_id BIGINT NOT NULL,
    note TEXT,
    CONSTRAINT fk_child_parent FOREIGN KEY (parent_id) REFERENCES parent (id)
);

CREATE INDEX idx_parent_status ON parent (status);

CREATE VIEW parent_v AS SELECT id, status FROM parent;

CREATE FUNCTION echo_id(x BIGINT) RETURNS BIGINT
DETERMINISTIC
RETURN x;
