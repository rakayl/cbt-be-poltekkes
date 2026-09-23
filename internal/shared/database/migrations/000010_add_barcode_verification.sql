-- Migration: 000010_add_barcode_verification.sql
-- Description: Add barcode attendance verification columns to cat.at_jadwalpeserta

ALTER TABLE cat.at_jadwalpeserta 
ADD COLUMN IF NOT EXISTS is_verified SMALLINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS barcode_scanned_at TIMESTAMP NULL,
ADD COLUMN IF NOT EXISTS scanned_by VARCHAR(50) NULL,
ADD COLUMN IF NOT EXISTS barcode_data VARCHAR(255) NULL;

CREATE INDEX IF NOT EXISTS idx_at_jadwalpeserta_verified 
ON cat.at_jadwalpeserta (idjadwalujian, kodepeserta, is_verified);
