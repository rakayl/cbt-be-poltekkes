-- ─────────────────────────────────────────────────────────────────────────────
-- Migration 000008: Allow Dynamic Duplicate Section Types Per Exam
-- Replaces UNIQUE(id_ep_exam, id_section) with UNIQUE(id_ep_exam, section_order)
-- ─────────────────────────────────────────────────────────────────────────────

-- 1. Drop old constraint that prevents having multiple sections of the same type in one exam
ALTER TABLE cat.ep_exam_sections DROP CONSTRAINT IF EXISTS ep_exam_sections_id_ep_exam_id_section_key;

-- 2. Add new unique constraint on (id_ep_exam, section_order)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'ep_exam_sections_id_ep_exam_section_order_key'
    ) THEN
        ALTER TABLE cat.ep_exam_sections ADD CONSTRAINT ep_exam_sections_id_ep_exam_section_order_key UNIQUE (id_ep_exam, section_order);
    END IF;
END $$;

-- 3. Recreate / add index for lookup
DROP INDEX IF EXISTS cat.idx_ep_exam_sections_lookup;
CREATE INDEX IF NOT EXISTS idx_ep_exam_sections_order ON cat.ep_exam_sections(id_ep_exam, section_order);
CREATE INDEX IF NOT EXISTS idx_ep_exam_sections_lookup ON cat.ep_exam_sections(id_ep_exam, id_section);
