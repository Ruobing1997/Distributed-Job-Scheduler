CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role int NOT NULL DEFAULT 0
);


CREATE INDEX IF NOT EXISTS user_name_idx on users(username);