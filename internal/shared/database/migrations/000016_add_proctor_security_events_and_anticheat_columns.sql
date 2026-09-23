-- Migration: 000016_add_proctor_security_events_and_anticheat_columns.sql
-- Description: Create security events tables (at_security_events & ep_security_events) and add complete anti-cheat, proctoring, lock, and token columns to existing tables.

-- 1. Create table for CBT Reguler Security Events Log
CREATE TABLE IF NOT EXISTS cat.at_security_events (
    id BIGSERIAL PRIMARY KEY,
    idjadwal INT NOT NULL,
    kodepeserta VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    risk_score INT NOT NULL DEFAULT 0,
    metadata JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_id VARCHAR(100),
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sec_events_created ON cat.at_security_events (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sec_events_jadwal_peserta ON cat.at_security_events (idjadwal, kodepeserta);
CREATE INDEX IF NOT EXISTS idx_security_events_lookup ON cat.at_security_events (idjadwal, kodepeserta, event_type);

-- 2. Create table for English Proficiency (TOEFL/TOEIC) Security Events Log
CREATE TABLE IF NOT EXISTS cat.ep_security_events (
    id BIGSERIAL PRIMARY KEY,
    id_ep_schedule INT NOT NULL,
    kodepeserta VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    risk_score INT NOT NULL DEFAULT 0,
    metadata JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_id VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ep_sec_events_created ON cat.ep_security_events (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ep_sec_events_schedule_peserta ON cat.ep_security_events (id_ep_schedule, kodepeserta);

-- 3. Add Anti-Cheat, Proctoring, Heartbeat & Lock Columns to cat.at_jadwalpeserta
ALTER TABLE cat.at_jadwalpeserta
    ADD COLUMN IF NOT EXISTS is_locked INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS lock_reason TEXT,
    ADD COLUMN IF NOT EXISTS proctor_warning TEXT,
    ADD COLUMN IF NOT EXISTS tab_switch_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fullscreen_exit_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS risk_score INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS risk_level VARCHAR(30) DEFAULT 'NORMAL',
    ADD COLUMN IF NOT EXISTS remaining_seconds INTEGER,
    ADD COLUMN IF NOT EXISTS last_heartbeat TIMESTAMP WITHOUT TIME ZONE,
    ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(255) DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS client_device_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS is_paused BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS session_snapshot JSONB,
    ADD COLUMN IF NOT EXISTS disconnect_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reconnect_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS submit_reason VARCHAR(100) DEFAULT 'MANUAL_SUBMIT';

CREATE INDEX IF NOT EXISTS idx_at_jadwalpeserta_heartbeat ON cat.at_jadwalpeserta (idjadwalujian, kodepeserta, last_heartbeat);
CREATE INDEX IF NOT EXISTS idx_at_jadwalpeserta_locked ON cat.at_jadwalpeserta (idjadwalujian, is_locked);

-- 4. Add token_ujian, grace_period_seconds & exam_deadline_at to cat.at_jadwalujian
ALTER TABLE cat.at_jadwalujian
    ADD COLUMN IF NOT EXISTS token_ujian VARCHAR(20),
    ADD COLUMN IF NOT EXISTS grace_period_seconds INTEGER DEFAULT 60,
    ADD COLUMN IF NOT EXISTS exam_deadline_at TIMESTAMP WITHOUT TIME ZONE;

-- 5. Add snapshot_metadata to cat.at_jawabanpeserta
ALTER TABLE cat.at_jawabanpeserta
    ADD COLUMN IF NOT EXISTS snapshot_metadata JSONB;

-- 6. Add dynamic scoring weights to cat.at_pertanyaan
ALTER TABLE cat.at_pertanyaan
    ADD COLUMN IF NOT EXISTS bobot NUMERIC,
    ADD COLUMN IF NOT EXISTS bobot_benar NUMERIC,
    ADD COLUMN IF NOT EXISTS bobot_salah NUMERIC;

-- 7. Add shuffle_options to cat.ep_exam_sections
ALTER TABLE cat.ep_exam_sections
    ADD COLUMN IF NOT EXISTS shuffle_options BOOLEAN DEFAULT FALSE;

-- 8. Add proctoring & security tracking columns to cat.ep_schedule_participants
ALTER TABLE cat.ep_schedule_participants
    ADD COLUMN IF NOT EXISTS proctor_warning TEXT,
    ADD COLUMN IF NOT EXISTS fullscreen_exit_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS risk_level VARCHAR(30) DEFAULT 'NORMAL',
    ADD COLUMN IF NOT EXISTS total_violations INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS unlock_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ;
