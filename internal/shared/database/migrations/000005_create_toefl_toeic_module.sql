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


