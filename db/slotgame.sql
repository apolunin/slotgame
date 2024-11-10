CREATE TABLE IF NOT EXISTS "user" (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    login VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    balance BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS spin_result (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES "user"(id) NOT NULL,
    combination VARCHAR(5) NOT NULL,
    result SMALLINT NOT NULL,
    bet_amount BIGINT NOT NULL,
    win_amount BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_spin_result_created_at ON spin_result (created_at);