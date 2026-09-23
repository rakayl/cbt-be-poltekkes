-- Migration 000017: Create at_api_keys table for M2M API Key integration management
CREATE TABLE IF NOT EXISTS cat.at_api_keys (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    api_key VARCHAR(100) UNIQUE NOT NULL,
    ip_whitelist TEXT DEFAULT NULL, -- Comma-separated list of IP addresses or CIDR blocks (optional)
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT DEFAULT NULL,
    last_used_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    softdelete CHAR(1) NOT NULL DEFAULT '0'
);

CREATE INDEX IF NOT EXISTS idx_at_api_keys_lookup 
ON cat.at_api_keys (api_key) 
WHERE softdelete = '0' AND is_active = TRUE;
