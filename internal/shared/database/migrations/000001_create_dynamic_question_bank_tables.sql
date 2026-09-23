-- Migration: 000001_create_dynamic_question_bank_tables.sql
-- Description: Tabel Bank Soal Dinamis, Stimulus Kasus (Teks/Audio/Gambar/Video), Sub-Items, dan Blueprint

CREATE SCHEMA IF NOT EXISTS cat;

-- 1. Tabel Stimulus / Naskah Kasus Induk
CREATE TABLE IF NOT EXISTS cat.cat_bank_stimulus (
    id_stimulus BIGSERIAL PRIMARY KEY,
    kodesoal VARCHAR(100) NOT NULL,
    stimulus_code VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    narrative_text TEXT NOT NULL,
    media_type VARCHAR(50) DEFAULT 'NONE', -- NONE, AUDIO, IMAGE, VIDEO, MULTI
    media_url TEXT,
    media_metadata JSONB DEFAULT '{}'::jsonb,
    category_tag VARCHAR(100) DEFAULT 'Umum',
    difficulty_level INT DEFAULT 2, -- 1=Mudah, 2=Sedang, 3=Sulit
    sort_order INT DEFAULT 1,
    softdelete CHAR(1) DEFAULT '0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stimulus_kodesoal ON cat.cat_bank_stimulus(kodesoal);
CREATE INDEX IF NOT EXISTS idx_stimulus_category_tag ON cat.cat_bank_stimulus(category_tag);
CREATE INDEX IF NOT EXISTS idx_stimulus_code ON cat.cat_bank_stimulus(stimulus_code);

-- 2. Tabel Sub-Pertanyaan di bawah Stimulus (Multi-Item per Stimulus)
CREATE TABLE IF NOT EXISTS cat.cat_stimulus_items (
    id_item BIGSERIAL PRIMARY KEY,
    id_stimulus BIGINT NOT NULL REFERENCES cat.cat_bank_stimulus(id_stimulus) ON DELETE CASCADE,
    item_order INT NOT NULL DEFAULT 1,
    question_text TEXT NOT NULL,
    question_media_type VARCHAR(50) DEFAULT 'NONE', -- NONE, AUDIO, IMAGE, VIDEO
    question_media_url TEXT,
    item_type VARCHAR(50) DEFAULT 'SINGLE_CHOICE', -- SINGLE_CHOICE, MULTI_CHOICE
    weight_correct NUMERIC(5,2) DEFAULT 1.0,
    weight_wrong NUMERIC(5,2) DEFAULT 0.0,
    weight_blank NUMERIC(5,2) DEFAULT 0.0,
    correct_answer VARCHAR(50) NOT NULL, -- Misal 'A' atau '1,3'
    explanation TEXT,
    softdelete CHAR(1) DEFAULT '0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_items_id_stimulus ON cat.cat_stimulus_items(id_stimulus);

-- 3. Tabel Pilihan Jawaban per Sub-Pertanyaan (A-E)
CREATE TABLE IF NOT EXISTS cat.cat_item_options (
    id_option BIGSERIAL PRIMARY KEY,
    id_item BIGINT NOT NULL REFERENCES cat.cat_stimulus_items(id_item) ON DELETE CASCADE,
    option_label VARCHAR(10) NOT NULL, -- A, B, C, D, E
    option_text TEXT NOT NULL,
    media_type VARCHAR(50) DEFAULT 'NONE', -- NONE, AUDIO, IMAGE
    media_url TEXT,
    sort_order INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_options_id_item ON cat.cat_item_options(id_item);

-- 4. Tabel Cetak Biru (Blueprints) untuk Peracikan Dinamis
CREATE TABLE IF NOT EXISTS cat.cat_exam_blueprints (
    id_blueprint BIGSERIAL PRIMARY KEY,
    blueprint_code VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    total_target_questions INT NOT NULL DEFAULT 50,
    passing_score NUMERIC(5,2) DEFAULT 60.0,
    duration_minutes INT NOT NULL DEFAULT 90,
    softdelete CHAR(1) DEFAULT '0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Tabel Aturan Komposisi Blueprint (Blueprint Rules)
CREATE TABLE IF NOT EXISTS cat.cat_blueprint_rules (
    id_rule BIGSERIAL PRIMARY KEY,
    id_blueprint BIGINT NOT NULL REFERENCES cat.cat_exam_blueprints(id_blueprint) ON DELETE CASCADE,
    category_tag VARCHAR(100) NOT NULL,
    difficulty_level INT DEFAULT 0, -- 0 = Semua, 1 = Mudah, 2 = Sedang, 3 = Sulit
    quota_count INT NOT NULL DEFAULT 10,
    weight_multiplier NUMERIC(5,2) DEFAULT 1.0,
    sort_order INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blueprint_rules_id ON cat.cat_blueprint_rules(id_blueprint);
