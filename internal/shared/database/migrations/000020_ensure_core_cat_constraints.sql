-- ─────────────────────────────────────────────────────────────────────────────
-- Migration 000020: Ensure Core CAT Constraints and Primary Keys Exist
-- Resolves "pq: there is no unique or exclusion constraint matching the ON CONFLICT specification"
-- ─────────────────────────────────────────────────────────────────────────────

-- 1. Ensure cat.at_peserta unique constraint / primary key on (kodepeserta)
DELETE FROM cat.at_peserta a
USING cat.at_peserta b
WHERE a.ctid < b.ctid AND a.kodepeserta = b.kodepeserta;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_constraint c
        WHERE c.conrelid = 'cat.at_peserta'::regclass
          AND c.contype IN ('p', 'u')
          AND (
              SELECT array_agg(a.attname::text ORDER BY u.ord)
              FROM unnest(c.conkey) WITH ORDINALITY AS u(attnum, ord)
              JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = u.attnum
          ) = ARRAY['kodepeserta']::text[]
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM pg_constraint c
            WHERE c.conrelid = 'cat.at_peserta'::regclass AND c.contype = 'p'
        ) THEN
            ALTER TABLE cat.at_peserta ADD CONSTRAINT at_peserta_pkey PRIMARY KEY (kodepeserta);
        ELSE
            ALTER TABLE cat.at_peserta ADD CONSTRAINT at_peserta_kodepeserta_key UNIQUE (kodepeserta);
        END IF;
    END IF;
END $$;

-- 2. Ensure cat.at_pesertaujian unique constraint / primary key on (kodepeserta, idujian)
DELETE FROM cat.at_pesertaujian a
USING cat.at_pesertaujian b
WHERE a.ctid < b.ctid AND a.kodepeserta = b.kodepeserta AND a.idujian = b.idujian;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_constraint c
        WHERE c.conrelid = 'cat.at_pesertaujian'::regclass
          AND c.contype IN ('p', 'u')
          AND (
              SELECT array_agg(a.attname::text ORDER BY u.ord)
              FROM unnest(c.conkey) WITH ORDINALITY AS u(attnum, ord)
              JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = u.attnum
          ) = ARRAY['kodepeserta', 'idujian']::text[]
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM pg_constraint c
            WHERE c.conrelid = 'cat.at_pesertaujian'::regclass AND c.contype = 'p'
        ) THEN
            ALTER TABLE cat.at_pesertaujian ADD CONSTRAINT at_pesertaujian_pkey PRIMARY KEY (kodepeserta, idujian);
        ELSE
            ALTER TABLE cat.at_pesertaujian ADD CONSTRAINT at_pesertaujian_kodepeserta_idujian_key UNIQUE (kodepeserta, idujian);
        END IF;
    END IF;
END $$;

-- 3. Ensure cat.at_jadwalpeserta unique constraint / primary key on (kodepeserta, idjadwalujian)
DELETE FROM cat.at_jadwalpeserta a
USING cat.at_jadwalpeserta b
WHERE a.ctid < b.ctid AND a.kodepeserta = b.kodepeserta AND a.idjadwalujian = b.idjadwalujian;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_constraint c
        WHERE c.conrelid = 'cat.at_jadwalpeserta'::regclass
          AND c.contype IN ('p', 'u')
          AND (
              SELECT array_agg(a.attname::text ORDER BY u.ord)
              FROM unnest(c.conkey) WITH ORDINALITY AS u(attnum, ord)
              JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = u.attnum
          ) = ARRAY['kodepeserta', 'idjadwalujian']::text[]
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM pg_constraint c
            WHERE c.conrelid = 'cat.at_jadwalpeserta'::regclass AND c.contype = 'p'
        ) THEN
            ALTER TABLE cat.at_jadwalpeserta ADD CONSTRAINT at_jadwalpeserta_pkey PRIMARY KEY (kodepeserta, idjadwalujian);
        ELSE
            ALTER TABLE cat.at_jadwalpeserta ADD CONSTRAINT at_jadwalpeserta_kodepeserta_idjadwalujian_key UNIQUE (kodepeserta, idjadwalujian);
        END IF;
    END IF;
END $$;
