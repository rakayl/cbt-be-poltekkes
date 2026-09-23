-- =============================================================================
-- Migration 000007: Create EP Question Banks Table
-- Schema: cat
-- =============================================================================

CREATE TABLE IF NOT EXISTS cat.ep_question_banks (
    id_ep_bank BIGSERIAL PRIMARY KEY,
    kodesoal VARCHAR(100) NOT NULL UNIQUE,
    namasoal VARCHAR(255) NOT NULL,
    id_exam_type INT NOT NULL REFERENCES cat.ep_exam_types(id_exam_type) ON DELETE CASCADE,
    id_section INT NOT NULL REFERENCES cat.ep_sections(id_section) ON DELETE CASCADE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    softdelete CHAR(1) DEFAULT '0'
);

CREATE INDEX IF NOT EXISTS idx_ep_bank_kodesoal ON cat.ep_question_banks(kodesoal);
CREATE INDEX IF NOT EXISTS idx_ep_bank_exam_type ON cat.ep_question_banks(id_exam_type);
CREATE INDEX IF NOT EXISTS idx_ep_bank_section ON cat.ep_question_banks(id_section);
