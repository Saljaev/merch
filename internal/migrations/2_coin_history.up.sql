CREATE TABLE IF NOT EXISTS coin_history (
	id BIGSERIAL PRIMARY KEY,
	from_user TEXT NOT NULL,
	from_user_id BIGINT,
	to_user TEXT NOT NULL,
	to_user_id BIGINT,
	amount INT NOT NULL,
	created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS coin_history_from_user_id_idx ON coin_history (from_user_id);
CREATE INDEX IF NOT EXISTS coin_history_to_user_id_idx ON coin_history (to_user_id);