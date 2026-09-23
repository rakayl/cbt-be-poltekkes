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

-- Insert existing seeded banks into cat.at_soal
INSERT INTO cat.at_soal (kodesoal, namasoal, keterangan)
VALUES
    ('TOEFL_L_01', 'TOEFL ITP Listening Comprehension Bank 01', 'Bank Soal Listening TOEFL ITP Resmi'),
    ('TOEFL_S_01', 'TOEFL ITP Structure & Written Expression Bank 01', 'Bank Soal Structure TOEFL ITP'),
    ('TOEFL_R_01', 'TOEFL ITP Reading Comprehension Bank 01', 'Bank Soal Reading TOEFL ITP'),
    ('TOEIC_L_01', 'TOEIC Listening Comprehension Bank 01', 'Bank Soal Listening TOEIC L&R'),
    ('TOEIC_R_01', 'TOEIC Reading Comprehension Bank 01', 'Bank Soal Reading TOEIC L&R')
ON CONFLICT (kodesoal) DO NOTHING;

-- Insert into ep_question_banks using dynamic FK lookups to avoid serial mismatch
INSERT INTO cat.ep_question_banks (kodesoal, namasoal, id_exam_type, id_section, description)
SELECT 'TOEFL_L_01', 'TOEFL ITP Listening Comprehension Bank 01', s.id_exam_type, s.id_section, 'Bank Soal Listening TOEFL ITP Resmi dengan Audio Percakapan & Part A/B/C'
FROM cat.ep_sections s
JOIN cat.ep_exam_types t ON t.id_exam_type = s.id_exam_type
WHERE t.type_code = 'TOEFL_ITP' AND s.section_code = 'LISTENING'
ON CONFLICT (kodesoal) DO UPDATE SET
    namasoal = EXCLUDED.namasoal,
    id_exam_type = EXCLUDED.id_exam_type,
    id_section = EXCLUDED.id_section,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO cat.ep_question_banks (kodesoal, namasoal, id_exam_type, id_section, description)
SELECT 'TOEFL_S_01', 'TOEFL ITP Structure & Written Expression Bank 01', s.id_exam_type, s.id_section, 'Bank Soal Structure & Written Expression TOEFL ITP'
FROM cat.ep_sections s
JOIN cat.ep_exam_types t ON t.id_exam_type = s.id_exam_type
WHERE t.type_code = 'TOEFL_ITP' AND s.section_code = 'STRUCTURE'
ON CONFLICT (kodesoal) DO UPDATE SET
    namasoal = EXCLUDED.namasoal,
    id_exam_type = EXCLUDED.id_exam_type,
    id_section = EXCLUDED.id_section,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO cat.ep_question_banks (kodesoal, namasoal, id_exam_type, id_section, description)
SELECT 'TOEFL_R_01', 'TOEFL ITP Reading Comprehension Bank 01', s.id_exam_type, s.id_section, 'Bank Soal Reading Comprehension TOEFL ITP Naskah Ilmiah'
FROM cat.ep_sections s
JOIN cat.ep_exam_types t ON t.id_exam_type = s.id_exam_type
WHERE t.type_code = 'TOEFL_ITP' AND s.section_code = 'READING'
ON CONFLICT (kodesoal) DO UPDATE SET
    namasoal = EXCLUDED.namasoal,
    id_exam_type = EXCLUDED.id_exam_type,
    id_section = EXCLUDED.id_section,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO cat.ep_question_banks (kodesoal, namasoal, id_exam_type, id_section, description)
SELECT 'TOEIC_L_01', 'TOEIC Listening Comprehension Bank 01', s.id_exam_type, s.id_section, 'Bank Soal Listening TOEIC L&R Topik Dunia Kerja & Bisnis'
FROM cat.ep_sections s
JOIN cat.ep_exam_types t ON t.id_exam_type = s.id_exam_type
WHERE t.type_code = 'TOEIC_LR' AND s.section_code = 'LISTENING'
ON CONFLICT (kodesoal) DO UPDATE SET
    namasoal = EXCLUDED.namasoal,
    id_exam_type = EXCLUDED.id_exam_type,
    id_section = EXCLUDED.id_section,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO cat.ep_question_banks (kodesoal, namasoal, id_exam_type, id_section, description)
SELECT 'TOEIC_R_01', 'TOEIC Reading Comprehension Bank 01', s.id_exam_type, s.id_section, 'Bank Soal Reading TOEIC L&R Topik Memo & Dokumen Kerja'
FROM cat.ep_sections s
JOIN cat.ep_exam_types t ON t.id_exam_type = s.id_exam_type
WHERE t.type_code = 'TOEIC_LR' AND s.section_code = 'READING'
ON CONFLICT (kodesoal) DO UPDATE SET
    namasoal = EXCLUDED.namasoal,
    id_exam_type = EXCLUDED.id_exam_type,
    id_section = EXCLUDED.id_section,
    description = EXCLUDED.description,
    updated_at = NOW();
