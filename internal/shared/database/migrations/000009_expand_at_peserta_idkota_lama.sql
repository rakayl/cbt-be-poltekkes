-- Migration 000009: Expand idkota_lama in cat.at_peserta to support text wilayah names and drop FK constraint
ALTER TABLE cat.at_peserta DROP CONSTRAINT IF EXISTS at_peserta_idkota_lama_fkey;
ALTER TABLE cat.at_peserta ALTER COLUMN idkota_lama TYPE character varying(100);
