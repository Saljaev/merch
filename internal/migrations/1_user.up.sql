CREATE TABLE IF NOT EXISTS users (
	id BIGINT PRIMARY KEY,
	username TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	coins INT
);

CREATE UNIQUE INDEX IF NOT EXISTS users_username_idx ON users (username);