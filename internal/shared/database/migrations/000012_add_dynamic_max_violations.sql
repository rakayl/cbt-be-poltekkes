-- 000012_add_dynamic_max_violations.sql
-- Add dynamic max_violations configuration to cat.at_ujian and cat.at_jadwalujian

ALTER TABLE cat.at_ujian ADD COLUMN IF NOT EXISTS max_violations integer DEFAULT 5;
ALTER TABLE cat.at_jadwalujian ADD COLUMN IF NOT EXISTS max_violations integer DEFAULT 5;

UPDATE cat.at_ujian SET max_violations = 5 WHERE max_violations IS NULL OR max_violations <= 0;
UPDATE cat.at_jadwalujian SET max_violations = 5 WHERE max_violations IS NULL OR max_violations <= 0;
