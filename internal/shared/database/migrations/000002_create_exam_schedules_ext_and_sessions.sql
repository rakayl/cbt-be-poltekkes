-- Migration: 000002_create_exam_schedules_ext_and_sessions.sql
-- Description: Tabel Ekstensi Jadwal Ujian (Jendela Waktu, Skoring Penalti), Sesi Peserta & Multi-Item Answers

CREATE SCHEMA IF NOT EXISTS cat;

-- 1. Tabel Ekstensi Jadwal Ujian (Jendela Waktu, Aturan Skoring, Auto-Submit)
CREATE TABLE IF NOT EXISTS cat.cat_exam_schedules_ext (
    idjadwalujian INT PRIMARY KEY, -- Ref ke cat.at_jadwalujian(idjadwalujian)
    id_blueprint BIGINT REFERENCES cat.cat_exam_blueprints(id_blueprint) ON DELETE SET NULL,
    window_start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Jam buka jendela ujian (08:00 WIB)
    window_end_time TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '7 hours'), -- Jam tutup jendela ujian (15:00 WIB)
    scoring_rule VARCHAR(50) DEFAULT 'STANDARD', -- STANDARD, PENALTY_MINUS_ONE, CUSTOM
    default_correct_score NUMERIC(5,2) DEFAULT 1.0,
    default_wrong_score NUMERIC(5,2) DEFAULT 0.0,
    default_blank_score NUMERIC(5,2) DEFAULT 0.0,
    auto_submit_on_window_end BOOLEAN DEFAULT TRUE,
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Tabel Ekstensi Sesi Peserta (State Pengerjaan, Timeout, Snapshot, Nilai)
CREATE TABLE IF NOT EXISTS cat.cat_participant_sessions_ext (
    id_session BIGSERIAL PRIMARY KEY,
    idjadwalujian INT NOT NULL,
    kodepeserta VARCHAR(100) NOT NULL,
    session_status VARCHAR(50) DEFAULT 'NOT_STARTED', -- NOT_STARTED, IN_PROGRESS, SUBMITTED, EXPIRED_AUTO_SUBMIT, LOCKED
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    calculated_deadline TIMESTAMP WITH TIME ZONE,
    session_snapshot JSONB,
    total_score NUMERIC(7,2) DEFAULT 0.0,
    total_correct INT DEFAULT 0,
    total_wrong INT DEFAULT 0,
    total_unanswered INT DEFAULT 0,
    pass_status VARCHAR(10) DEFAULT 'TL', -- L, TL
    submit_reason VARCHAR(100),
    client_ip VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_session_participant UNIQUE (idjadwalujian, kodepeserta)
);

CREATE INDEX IF NOT EXISTS idx_sessions_jadwal_peserta ON cat.cat_participant_sessions_ext(idjadwalujian, kodepeserta);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON cat.cat_participant_sessions_ext(session_status);

-- 3. Tabel Rekaman Jawaban Sub-Item Multi Pertanyaan
CREATE TABLE IF NOT EXISTS cat.cat_participant_answers_multi (
    id_answer BIGSERIAL PRIMARY KEY,
    id_session BIGINT NOT NULL REFERENCES cat.cat_participant_sessions_ext(id_session) ON DELETE CASCADE,
    id_stimulus BIGINT NOT NULL REFERENCES cat.cat_bank_stimulus(id_stimulus),
    id_item BIGINT NOT NULL REFERENCES cat.cat_stimulus_items(id_item),
    selected_options VARCHAR(50), -- Opsi yang dipilih peserta ('A', 'B', 'C', dll)
    is_doubtful BOOLEAN DEFAULT FALSE,
    is_correct BOOLEAN DEFAULT FALSE,
    awarded_points NUMERIC(5,2) DEFAULT 0.0,
    answered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_session_item UNIQUE (id_session, id_item)
);

CREATE INDEX IF NOT EXISTS idx_answers_session ON cat.cat_participant_answers_multi(id_session);
CREATE INDEX IF NOT EXISTS idx_answers_item ON cat.cat_participant_answers_multi(id_item);
