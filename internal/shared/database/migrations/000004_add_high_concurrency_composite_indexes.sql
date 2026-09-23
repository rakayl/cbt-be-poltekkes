-- Migration: 000004_add_high_concurrency_composite_indexes.sql
-- Description: Creates composite indexes on hot transaction tables for high concurrent load

-- 1. Index on participant answer transactions (UPSERT and SELECT lookups)
CREATE INDEX IF NOT EXISTS idx_jawabanpeserta_composite
ON cat.at_jawabanpeserta (idjadwalujian, kodepeserta, nourut);

-- 2. Index on participant exam status and scoring
CREATE INDEX IF NOT EXISTS idx_jadwalpeserta_lookup
ON cat.at_jadwalpeserta (idjadwalujian, kodepeserta, nilai);

-- 3. Index on question master queries
CREATE INDEX IF NOT EXISTS idx_pertanyaan_lookup
ON cat.at_pertanyaan (kodesoal, nourut);

-- 4. Index on participant session extension table
CREATE INDEX IF NOT EXISTS idx_sessions_ext_lookup
ON cat.cat_participant_sessions_ext (idjadwalujian, kodepeserta, session_status);
