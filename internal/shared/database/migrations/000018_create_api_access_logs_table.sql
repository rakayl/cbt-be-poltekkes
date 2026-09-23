-- Migration 000018: Create at_api_access_logs table for M2M API audit & access logging
CREATE TABLE IF NOT EXISTS cat.at_api_access_logs (
    id BIGSERIAL PRIMARY KEY,
    api_key_id INTEGER REFERENCES cat.at_api_keys(id) ON DELETE SET NULL,
    client_name VARCHAR(100) NOT NULL DEFAULT 'Unknown / Unauthenticated',
    ip_address VARCHAR(50) NOT NULL,
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER NOT NULL DEFAULT 0,
    user_agent VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_at_api_access_logs_key_date 
ON cat.at_api_access_logs (api_key_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_at_api_access_logs_created_at 
ON cat.at_api_access_logs (created_at DESC);
