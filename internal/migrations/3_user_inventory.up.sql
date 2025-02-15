CREATE TABLE IF NOT EXISTS inventory (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    item TEXT NOT NULL,
    quantity INT
);

CREATE INDEX IF NOT EXISTS inventory_user_idx ON inventory (user_id);
ALTER TABLE inventory ADD CONSTRAINT inventory_unique_user_item UNIQUE (user_id, item);