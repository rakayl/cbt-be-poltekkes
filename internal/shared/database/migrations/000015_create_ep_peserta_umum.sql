-- =============================================================================
-- Migration: 000015_create_ep_peserta_umum.sql
-- Description: Tabel master peserta umum / eksternal untuk modul TOEFL / EPT
-- =============================================================================

CREATE TABLE IF NOT EXISTS cat.ep_peserta_umum (
    id_peserta_umum     SERIAL PRIMARY KEY,
    kodepeserta         VARCHAR(30) UNIQUE NOT NULL, -- NIK atau Nomor Registrasi TOEFL (misal: 'UM260001')
    nik                 VARCHAR(20),                 -- NIK KTP resmi (untuk verifikasi sertifikat)
    nama                VARCHAR(150) NOT NULL,
    jenis_kelamin       CHAR(1) DEFAULT 'L',         -- 'L' / 'P'
    email               VARCHAR(100) UNIQUE NOT NULL,
    hp                  VARCHAR(30) NOT NULL,        -- Nomor WhatsApp / HP
    instansi            VARCHAR(150),                -- Asal institusi / tempat kerja / kampus luar
    tempat_lahir        VARCHAR(100),
    tanggal_lahir       DATE,                        -- Diperlukan untuk penerbitan Sertifikat EPT resmi
    alamat              TEXT,
    password            VARCHAR(255) NOT NULL,       -- Hash password untuk login portal /toefl/login
    is_active           BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- Indeks performa untuk autentikasi dan pencarian cepat
CREATE INDEX IF NOT EXISTS idx_ep_peserta_umum_lookup ON cat.ep_peserta_umum(kodepeserta);
CREATE INDEX IF NOT EXISTS idx_ep_peserta_umum_email ON cat.ep_peserta_umum(email);
CREATE INDEX IF NOT EXISTS idx_ep_peserta_umum_nik ON cat.ep_peserta_umum(nik);

-- Komentar dokumentasi
COMMENT ON TABLE cat.ep_peserta_umum IS 'Master data peserta umum/eksternal untuk ujian bahasa TOEFL/EPT terpisah dari pendaftar SPMB';
