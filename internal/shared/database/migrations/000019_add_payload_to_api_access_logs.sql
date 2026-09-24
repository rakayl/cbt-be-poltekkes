-- Migration 000019: Add request_body and response_body to cat.at_api_access_logs for M2M audit payload inspection
ALTER TABLE cat.at_api_access_logs 
ADD COLUMN IF NOT EXISTS request_body TEXT DEFAULT NULL,
ADD COLUMN IF NOT EXISTS response_body TEXT DEFAULT NULL;
