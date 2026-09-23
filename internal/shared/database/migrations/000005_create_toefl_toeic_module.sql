-- =============================================================================
-- Migration 000005: TOEFL ITP & TOEIC L&R Dynamic Proficiency Testing Module
-- Schema: cat
-- Fully Dynamic Architecture:
-- 1. Dynamic Mode Question Engine (Stimulus -> Multi-Item -> Options)
-- 2. Dynamic Section Time & Duration per Exam Section
-- 3. Dynamic Audio Replay (Configurable allow_replay & max_replay_count)
-- 4. Dynamic Score Conversion Profiles (ETS Standard vs Custom Poltekkes)
-- 5. Dynamic Official Certificates (Poltekkes Logo, Validity Override, QR Code & CEFR)
-- =============================================================================

CREATE SCHEMA IF NOT EXISTS cat;

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. Jenis Ujian (TOEFL ITP, TOEIC L&R)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_exam_types (
    id_exam_type    SERIAL PRIMARY KEY,
    type_code       VARCHAR(30) NOT NULL UNIQUE,      -- 'TOEFL_ITP', 'TOEIC_LR'
    type_name       VARCHAR(100) NOT NULL,             -- 'TOEFL ITP (Institutional Testing Program)'
    total_sections  INT NOT NULL DEFAULT 3,
    score_min       INT NOT NULL DEFAULT 310,
    score_max       INT NOT NULL DEFAULT 677,
    score_formula   VARCHAR(50) NOT NULL DEFAULT 'SCALED_AVERAGE', -- 'SCALED_AVERAGE', 'SCALED_SUM'
    validity_months INT NOT NULL DEFAULT 24,           -- Default masa berlaku sertifikat (2 tahun)
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. Seksi Standar Ujian (Listening, Structure, Reading)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_sections (
    id_section       SERIAL PRIMARY KEY,
    id_exam_type     INT NOT NULL REFERENCES cat.ep_exam_types(id_exam_type),
    section_code     VARCHAR(30) NOT NULL,             -- 'LISTENING', 'STRUCTURE', 'READING'
    section_name     VARCHAR(100) NOT NULL,             -- 'Section 1: Listening Comprehension'
    section_order    INT NOT NULL DEFAULT 1,            -- 1, 2, 3
    default_questions INT NOT NULL DEFAULT 50,          -- Jumlah soal standar
    default_duration INT NOT NULL DEFAULT 35,           -- Default durasi menit
    has_audio        BOOLEAN NOT NULL DEFAULT FALSE,
    allow_replay     BOOLEAN NOT NULL DEFAULT TRUE,     -- Replay audio diizinkan secara dinamis
    max_replay_count INT NOT NULL DEFAULT 0,            -- 0 = Unlimited replay, >0 = batasan putar
    can_go_back      BOOLEAN NOT NULL DEFAULT TRUE,     -- Navigasi balik antar soal
    score_scale_min  INT NOT NULL DEFAULT 20,
    score_scale_max  INT NOT NULL DEFAULT 68,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(id_exam_type, section_code)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 3. Sub-Bagian Dalam Seksi (Part A, B, C)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_section_parts (
    id_part          SERIAL PRIMARY KEY,
    id_section       INT NOT NULL REFERENCES cat.ep_sections(id_section),
    part_code        VARCHAR(20) NOT NULL,             -- 'PART_A', 'PART_B', 'PART_C'
    part_name        VARCHAR(100) NOT NULL,             -- 'Part A: Short Conversations'
    part_order       INT NOT NULL DEFAULT 1,
    question_start   INT NOT NULL,                      -- 1
    question_end     INT NOT NULL,                      -- 30
    instruction_text TEXT,
    has_passage      BOOLEAN NOT NULL DEFAULT FALSE,
    has_audio        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(id_section, part_code)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 4. Profil Konversi Skor Dinamis (ETS Standard vs Kustom Poltekkes)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_conversion_profiles (
    id_profile   SERIAL PRIMARY KEY,
    id_exam_type INT NOT NULL REFERENCES cat.ep_exam_types(id_exam_type),
    profile_code VARCHAR(50) NOT NULL,                -- 'ETS_STANDARD_2026', 'POLTEKKES_CUSTOM_2026'
    profile_name VARCHAR(200) NOT NULL,                -- 'Standar Resmi ETS', 'Tabel Kustom Poltekkes'
    is_default   BOOLEAN NOT NULL DEFAULT FALSE,
    description  TEXT,
    created_by   VARCHAR(50),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(id_exam_type, profile_code)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 5. Tabel Nilai Konversi Skor Berskala (Raw -> Scaled)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_score_conversion (
    id_conversion SERIAL PRIMARY KEY,
    id_profile    INT NOT NULL REFERENCES cat.ep_conversion_profiles(id_profile) ON DELETE CASCADE,
    section_code  VARCHAR(30) NOT NULL,                -- 'LISTENING', 'STRUCTURE', 'READING'
    raw_score     INT NOT NULL,                        -- 0 - 50 (Jumlah benar)
    scaled_score  INT NOT NULL,                        -- 24 - 68 (Skor berskala)
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(id_profile, section_code, raw_score)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 6. Pemetaan Level CEFR (Skor Total -> A1, A2, B1, B1+, B2, C1, C2)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_cefr_mapping (
    id_mapping   SERIAL PRIMARY KEY,
    id_exam_type INT NOT NULL REFERENCES cat.ep_exam_types(id_exam_type),
    cefr_level   VARCHAR(5) NOT NULL,                -- 'A2', 'B1', 'B1+', 'B2', 'C1'
    score_min    INT NOT NULL,
    score_max    INT NOT NULL,
    description  TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(id_exam_type, cefr_level)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 7. Template Sertifikat Resmi Poltekkes
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_certificate_templates (
    id_template     SERIAL PRIMARY KEY,
    id_exam_type    INT NOT NULL REFERENCES cat.ep_exam_types(id_exam_type),
    template_name   VARCHAR(100) NOT NULL,
    logo_url        TEXT DEFAULT '/assets/logo-poltekkes.png',
    header_text     TEXT DEFAULT 'KEMENTERIAN KESEHATAN REPUBLIK INDONESIA',
    institution_name TEXT DEFAULT 'POLITEKNIK KESEHATAN KEMENKES SEMARANG',
    cert_title      VARCHAR(150) DEFAULT 'OFFICIAL SCORE CERTIFICATE',
    signatory_name  VARCHAR(200) DEFAULT 'Direktur Poltekkes Kemenkes',
    signatory_title VARCHAR(200) DEFAULT 'NIP. 19750815 199903 1 002',
    signature_url   TEXT,
    stamp_url       TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 8. Instansi Ujian TOEFL / TOEIC (Header Ujian Admin)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_exams (
    id_ep_exam               SERIAL PRIMARY KEY,
    id_exam_type             INT NOT NULL REFERENCES cat.ep_exam_types(id_exam_type),
    id_profile               INT REFERENCES cat.ep_conversion_profiles(id_profile),
    id_template              INT REFERENCES cat.ep_certificate_templates(id_template),
    exam_name                VARCHAR(200) NOT NULL,   -- 'Uji Kemampuan TOEFL ITP Mahasiswa 2026'
    passing_score            INT DEFAULT 450,          -- Skor minimum kelulusan
    validity_months_override INT,                      -- Override masa berlaku dinamis (opsional)
    description              TEXT,
    is_active                BOOLEAN NOT NULL DEFAULT TRUE,
    created_by               VARCHAR(50),
    created_at               TIMESTAMPTZ DEFAULT NOW(),
    updated_at               TIMESTAMPTZ DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 9. Konfigurasi Seksi & Waktu Dinamis Per Ujian (Dynamic Section Timer & Bank Soal)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_exam_sections (
    id_ep_exam_section SERIAL PRIMARY KEY,
    id_ep_exam         INT NOT NULL REFERENCES cat.ep_exams(id_ep_exam) ON DELETE CASCADE,
    id_section         INT NOT NULL REFERENCES cat.ep_sections(id_section),
    kodesoal           VARCHAR(100) NOT NULL,         -- Relasi ke cat.cat_bank_stimulus(kodesoal) mode dinamis
    section_order      INT NOT NULL DEFAULT 1,
    duration_minutes   INT NOT NULL DEFAULT 35,       -- Durasi pengerjaan dinamis per seksi
    allow_replay       BOOLEAN NOT NULL DEFAULT TRUE, -- Izin putar ulang audio per seksi
    max_replay_count   INT NOT NULL DEFAULT 0,        -- Batasan putar (0 = unlimited)
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(id_ep_exam, id_section)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 10. Jadwal Sesi Ujian TOEFL / TOEIC
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_schedules (
    id_ep_schedule SERIAL PRIMARY KEY,
    id_ep_exam     INT NOT NULL REFERENCES cat.ep_exams(id_ep_exam),
    id_ruang       INT NOT NULL,
    exam_date      DATE NOT NULL,
    start_time     TIME NOT NULL,
    end_time       TIME NOT NULL,
    session_token  VARCHAR(50) NOT NULL,
    capacity       INT NOT NULL DEFAULT 40,
    proctor_name   VARCHAR(100),
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 11. Peserta Terdaftar Per Sesi & Tracking Sesi Seksi Dinamis
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_schedule_participants (
    id_participant      SERIAL PRIMARY KEY,
    id_ep_schedule      INT NOT NULL REFERENCES cat.ep_schedules(id_ep_schedule) ON DELETE CASCADE,
    kodepeserta         VARCHAR(20) NOT NULL,
    seat_number         INT,
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ,
    current_section     INT DEFAULT 1,                 -- Seksi aktif saat ini (1, 2, atau 3)
    section_started_at  TIMESTAMPTZ,                   -- Waktu mulai seksi aktif (untuk countdown dinamis)
    section_deadline_at TIMESTAMPTZ,                   -- Batas waktu seksi aktif
    session_status      VARCHAR(30) DEFAULT 'NOT_STARTED', -- 'NOT_STARTED', 'IN_PROGRESS', 'SUBMITTED', 'EXPIRED'
    is_locked           BOOLEAN DEFAULT FALSE,
    lock_reason         TEXT,
    risk_score          INT DEFAULT 0,
    tab_switch_count    INT DEFAULT 0,
    softdelete          VARCHAR(1) DEFAULT '0',
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(id_ep_schedule, kodepeserta)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 12. Rekapitulasi Nilai & Skor Konversi Peserta
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_scores (
    id_score           SERIAL PRIMARY KEY,
    id_participant     INT NOT NULL REFERENCES cat.ep_schedule_participants(id_participant) ON DELETE CASCADE,
    listening_raw      INT DEFAULT 0,
    listening_scaled   INT DEFAULT 0,
    structure_raw      INT DEFAULT 0,
    structure_scaled   INT DEFAULT 0,
    reading_raw        INT DEFAULT 0,
    reading_scaled     INT DEFAULT 0,
    total_scaled_score INT DEFAULT 0,
    cefr_level         VARCHAR(5),
    passing_status     VARCHAR(10) DEFAULT 'PENDING', -- 'PASS', 'FAIL', 'PENDING'
    calculated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(id_participant)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- 13. Sertifikat & Transkrip Skor Resmi (QR Code & Verifikasi)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS cat.ep_certificates (
    id_certificate     SERIAL PRIMARY KEY,
    id_score           INT NOT NULL REFERENCES cat.ep_scores(id_score) ON DELETE CASCADE,
    certificate_number VARCHAR(50) NOT NULL UNIQUE,  -- e.g. 'TOEFL/PKK/2026/00001'
    participant_name   VARCHAR(200) NOT NULL,
    participant_code   VARCHAR(20) NOT NULL,
    exam_type_name     VARCHAR(100) NOT NULL,
    exam_date          DATE NOT NULL,
    exam_location      VARCHAR(200) DEFAULT 'Lab CBT Poltekkes Kemenkes Semarang',
    listening_score    INT NOT NULL,
    structure_score    INT NOT NULL DEFAULT 0,
    reading_score      INT NOT NULL,
    total_score        INT NOT NULL,
    cefr_level         VARCHAR(5),
    issued_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    expiry_date        DATE NOT NULL,                  -- Dihitung dinamis: issued_date + validity
    is_revoked         BOOLEAN DEFAULT FALSE,
    revoked_reason     TEXT,
    verification_code  VARCHAR(100) NOT NULL UNIQUE,   -- Unique UUID untuk verifikasi publik via QR
    qr_code_url        TEXT,
    pdf_url            TEXT,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- INDEXES for Ultra-Fast Retrieval & High Concurrency
-- ─────────────────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_ep_exam_sections_lookup ON cat.ep_exam_sections(id_ep_exam, id_section);
CREATE INDEX IF NOT EXISTS idx_ep_conversion_profile_lookup ON cat.ep_score_conversion(id_profile, section_code, raw_score);
CREATE INDEX IF NOT EXISTS idx_ep_schedule_participants_lookup ON cat.ep_schedule_participants(id_ep_schedule, kodepeserta);
CREATE INDEX IF NOT EXISTS idx_ep_scores_participant ON cat.ep_scores(id_participant);
CREATE INDEX IF NOT EXISTS idx_ep_certificates_verify ON cat.ep_certificates(verification_code);

-- ─────────────────────────────────────────────────────────────────────────────
-- SEED DATA: 1. Exam Types
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO cat.ep_exam_types (type_code, type_name, total_sections, score_min, score_max, score_formula, validity_months)
VALUES
    ('TOEFL_ITP', 'TOEFL ITP (Institutional Testing Program)', 3, 310, 677, 'SCALED_AVERAGE', 24),
    ('TOEIC_LR',  'TOEIC Listening & Reading',                 2, 10,  990, 'SCALED_SUM',     24)
ON CONFLICT (type_code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- SEED DATA: 2. Standard Sections (TOEFL ITP)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO cat.ep_sections (id_exam_type, section_code, section_name, section_order, default_questions, default_duration, has_audio, allow_replay, max_replay_count, can_go_back, score_scale_min, score_scale_max)
VALUES
    ((SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code='TOEFL_ITP'), 'LISTENING',  'Section 1: Listening Comprehension',       1, 50, 35, TRUE,  TRUE, 0, TRUE, 24, 68),
    ((SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code='TOEFL_ITP'), 'STRUCTURE',  'Section 2: Structure & Written Expression', 2, 40, 25, FALSE, FALSE, 0, TRUE, 20, 68),
    ((SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code='TOEFL_ITP'), 'READING',    'Section 3: Reading Comprehension',          3, 50, 55, FALSE, FALSE, 0, TRUE, 20, 67)
ON CONFLICT (id_exam_type, section_code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- SEED DATA: 3. Standard Sections (TOEIC L&R)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO cat.ep_sections (id_exam_type, section_code, section_name, section_order, default_questions, default_duration, has_audio, allow_replay, max_replay_count, can_go_back, score_scale_min, score_scale_max)
VALUES
    ((SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code='TOEIC_LR'), 'LISTENING', 'Section 1: Listening',  1, 100, 45, TRUE,  TRUE, 0, TRUE, 5, 495),
    ((SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code='TOEIC_LR'), 'READING',   'Section 2: Reading',    2, 100, 75, FALSE, FALSE, 0, TRUE, 5, 495)
ON CONFLICT (id_exam_type, section_code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- SEED DATA: 4. Section Parts (TOEFL ITP)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO cat.ep_section_parts (id_section, part_code, part_name, part_order, question_start, question_end, instruction_text, has_passage, has_audio)
SELECT s.id_section, p.part_code, p.part_name, p.part_order, p.question_start, p.question_end, p.instruction_text, p.has_passage, p.has_audio
FROM cat.ep_sections s
CROSS JOIN (VALUES
    ('PART_A', 'Part A: Short Conversations',  1,  1, 30, 'In Part A, you will hear short conversations between two people. After each conversation, choose the best answer.', FALSE, TRUE),
    ('PART_B', 'Part B: Longer Conversations', 2, 31, 38, 'In Part B, you will hear longer conversations followed by questions. Choose the best answer.', FALSE, TRUE),
    ('PART_C', 'Part C: Short Talks / Lectures', 3, 39, 50, 'In Part C, you will hear short lectures or talks followed by questions. Choose the best answer.', FALSE, TRUE)
) AS p(part_code, part_name, part_order, question_start, question_end, instruction_text, has_passage, has_audio)
WHERE s.section_code = 'LISTENING'
AND s.id_exam_type = (SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code = 'TOEFL_ITP')
ON CONFLICT (id_section, part_code) DO NOTHING;

INSERT INTO cat.ep_section_parts (id_section, part_code, part_name, part_order, question_start, question_end, instruction_text, has_passage, has_audio)
SELECT s.id_section, p.part_code, p.part_name, p.part_order, p.question_start, p.question_end, p.instruction_text, p.has_passage, p.has_audio
FROM cat.ep_sections s
CROSS JOIN (VALUES
    ('PART_A', 'Part A: Incomplete Sentences', 1,  1, 15, 'Choose the word or phrase that best completes the sentence correctly.', FALSE, FALSE),
    ('PART_B', 'Part B: Error Recognition',     2, 16, 40, 'Identify the underlined word or phrase that must be changed for the sentence to be correct.', FALSE, FALSE)
) AS p(part_code, part_name, part_order, question_start, question_end, instruction_text, has_passage, has_audio)
WHERE s.section_code = 'STRUCTURE'
AND s.id_exam_type = (SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code = 'TOEFL_ITP')
ON CONFLICT (id_section, part_code) DO NOTHING;

INSERT INTO cat.ep_section_parts (id_section, part_code, part_name, part_order, question_start, question_end, instruction_text, has_passage, has_audio)
SELECT s.id_section, 'PART_A', 'Reading Passages', 1, 1, 50,
       'Read the passage and choose the best answer to each question based on what is stated or implied in the text.',
       TRUE, FALSE
FROM cat.ep_sections s
WHERE s.section_code = 'READING'
AND s.id_exam_type = (SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code = 'TOEFL_ITP')
ON CONFLICT (id_section, part_code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- SEED DATA: 5. Conversion Profiles (ETS Standard & Poltekkes Custom)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO cat.ep_conversion_profiles (id_exam_type, profile_code, profile_name, is_default, description)
VALUES
    ((SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code='TOEFL_ITP'), 'ETS_STANDARD', 'Tabel Konversi Resmi ETS (TOEFL ITP)', TRUE, 'Tabel konversi standar ETS 310 - 677.'),
    ((SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code='TOEFL_ITP'), 'POLTEKKES_CUSTOM', 'Tabel Konversi Kustom Poltekkes', FALSE, 'Tabel konversi institusi Poltekkes Kemenkes.')
ON CONFLICT (id_exam_type, profile_code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- SEED DATA: 6. ETS Official Conversion Table (TOEFL ITP)
-- ─────────────────────────────────────────────────────────────────────────────
-- Section 1: Listening (50 raw -> scaled 24-68)
INSERT INTO cat.ep_score_conversion (id_profile, section_code, raw_score, scaled_score)
SELECT p.id_profile, 'LISTENING', v.raw, v.scaled
FROM cat.ep_conversion_profiles p
CROSS JOIN (VALUES
    (50,68),(49,67),(48,66),(47,65),(46,63),(45,63),(44,62),(43,61),(42,60),(41,59),
    (40,58),(39,57),(38,57),(37,56),(36,55),(35,54),(34,53),(33,52),(32,52),(31,51),
    (30,51),(29,50),(28,49),(27,49),(26,48),(25,48),(24,47),(23,47),(22,46),(21,45),
    (20,45),(19,44),(18,43),(17,42),(16,41),(15,41),(14,40),(13,39),(12,38),(11,37),
    (10,36),(9,35),(8,34),(7,33),(6,32),(5,31),(4,30),(3,29),(2,28),(1,26),(0,24)
) AS v(raw, scaled)
WHERE p.profile_code = 'ETS_STANDARD'
ON CONFLICT (id_profile, section_code, raw_score) DO NOTHING;

-- Section 2: Structure (40 raw -> scaled 20-68)
INSERT INTO cat.ep_score_conversion (id_profile, section_code, raw_score, scaled_score)
SELECT p.id_profile, 'STRUCTURE', v.raw, v.scaled
FROM cat.ep_conversion_profiles p
CROSS JOIN (VALUES
    (40,68),(39,67),(38,66),(37,64),(36,63),(35,62),(34,61),(33,60),(32,59),(31,58),
    (30,57),(29,56),(28,55),(27,54),(26,53),(25,53),(24,52),(23,51),(22,50),(21,49),
    (20,48),(19,47),(18,46),(17,45),(16,44),(15,43),(14,42),(13,41),(12,40),(11,39),
    (10,38),(9,37),(8,36),(7,34),(6,32),(5,31),(4,29),(3,27),(2,25),(1,23),(0,20)
) AS v(raw, scaled)
WHERE p.profile_code = 'ETS_STANDARD'
ON CONFLICT (id_profile, section_code, raw_score) DO NOTHING;

-- Section 3: Reading (50 raw -> scaled 20-67)
INSERT INTO cat.ep_score_conversion (id_profile, section_code, raw_score, scaled_score)
SELECT p.id_profile, 'READING', v.raw, v.scaled
FROM cat.ep_conversion_profiles p
CROSS JOIN (VALUES
    (50,67),(49,66),(48,65),(47,64),(46,63),(45,62),(44,61),(43,60),(42,59),(41,58),
    (40,58),(39,57),(38,56),(37,55),(36,55),(35,54),(34,53),(33,52),(32,52),(31,51),
    (30,50),(29,49),(28,49),(27,48),(26,47),(25,46),(24,45),(23,44),(22,43),(21,43),
    (20,42),(19,41),(18,40),(17,39),(16,38),(15,37),(14,36),(13,35),(12,34),(11,33),
    (10,32),(9,31),(8,30),(7,29),(6,28),(5,27),(4,26),(3,25),(2,23),(1,22),(0,20)
) AS v(raw, scaled)
WHERE p.profile_code = 'ETS_STANDARD'
ON CONFLICT (id_profile, section_code, raw_score) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- SEED DATA: 7. CEFR Mapping (TOEFL ITP)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO cat.ep_cefr_mapping (id_exam_type, cefr_level, score_min, score_max, description)
SELECT et.id_exam_type, v.cefr, v.smin, v.smax, v.description
FROM cat.ep_exam_types et
CROSS JOIN (VALUES
    ('A2',  310, 336, 'Elementary: Memahami kalimat dan ekspresi dasar terkait hal-hal langsung.'),
    ('B1',  337, 459, 'Intermediate: Memahami poin utama masukan standar tentang hal-hal umum/kerja.'),
    ('B1+', 460, 542, 'Upper Intermediate: Memahami gagasan utama teks kompleks dan abstrak.'),
    ('B2',  543, 626, 'Advanced: Memahami teks panjang dan menuntut serta makna implisit secara lancar.'),
    ('C1',  627, 677, 'Proficient: Menggunakan bahasa secara fleksibel dan efektif untuk tujuan sosial, akademis, dan profesional.')
) AS v(cefr, smin, smax, description)
WHERE et.type_code = 'TOEFL_ITP'
ON CONFLICT (id_exam_type, cefr_level) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- SEED DATA: 8. Certificate Template (Poltekkes Kemenkes)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO cat.ep_certificate_templates (id_exam_type, template_name, logo_url, header_text, institution_name, cert_title, signatory_name, signatory_title)
VALUES (
    (SELECT id_exam_type FROM cat.ep_exam_types WHERE type_code='TOEFL_ITP'),
    'Template Resmi Poltekkes Kemenkes',
    '/assets/logo-poltekkes.png',
    'KEMENTERIAN KESEHATAN REPUBLIK INDONESIA',
    'POLITEKNIK KESEHATAN KEMENKES SEMARANG',
    'ENGLISH PROFICIENCY TEST CERTIFICATE',
    'Dr. H. Ahmad, S.Kp, M.Kes',
    'Direktur Poltekkes Kemenkes Semarang'
);
