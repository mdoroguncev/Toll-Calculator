-- +goose Up
CREATE TABLE distance (
    id         BIGSERIAL   PRIMARY KEY,
    obu_id     INT         NOT NULL,
    value      FLOAT       NOT NULL,
    unix       BIGINT      NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX distance_obu_id_idx
    ON distance(obu_id);

-- +goose Down
DROP TABLE distance;