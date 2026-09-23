-- Migration: 000003_support_static_pertanyaan_multimedia.sql
-- Description: Mendukung URL/Path Multimedia Penuh (Teks, Audio, Gambar, Video) pada cat.at_pertanyaan

CREATE SCHEMA IF NOT EXISTS cat;

-- Ubah tipe kolom media menjadi TEXT agar dapat menampung URL penuh ataupun ekstensi lama
ALTER TABLE cat.at_pertanyaan ALTER COLUMN pertanyaanimage TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN pertanyaanaudio TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN pertanyaanvideo TYPE TEXT;

ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban1image TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban1audio TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban1video TYPE TEXT;

ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban2image TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban2audio TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban2video TYPE TEXT;

ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban3image TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban3audio TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban3video TYPE TEXT;

ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban4image TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban4audio TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban4video TYPE TEXT;

ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban5image TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban5audio TYPE TEXT;
ALTER TABLE cat.at_pertanyaan ALTER COLUMN jawaban5video TYPE TEXT;
