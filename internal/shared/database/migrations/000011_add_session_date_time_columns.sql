-- Migration: Add tglmulai, tglselesai, waktumulai, waktuselesai to cat.at_jadwalujian
ALTER TABLE cat.at_jadwalujian ADD COLUMN IF NOT EXISTS tglmulai date;
ALTER TABLE cat.at_jadwalujian ADD COLUMN IF NOT EXISTS tglselesai date;
ALTER TABLE cat.at_jadwalujian ADD COLUMN IF NOT EXISTS waktumulai character(4);
ALTER TABLE cat.at_jadwalujian ADD COLUMN IF NOT EXISTS waktuselesai character(4);

ALTER TABLE cat.at_ruangujian ADD COLUMN IF NOT EXISTS waktuselesai character(4);

-- Backfill from at_ruangujian if available
UPDATE cat.at_jadwalujian j
SET tglmulai = r.tglmulai,
    tglselesai = r.tglselesai,
    waktumulai = r.waktumulai,
    waktuselesai = r.waktuselesai
FROM (
    SELECT idjadwalujian,
           MIN(tglmulai) AS tglmulai,
           MAX(COALESCE(tglselesai, tglmulai)) AS tglselesai,
           MIN(waktumulai) AS waktumulai,
           MAX(waktuselesai) AS waktuselesai
    FROM cat.at_ruangujian
    WHERE (softdelete = '0' OR softdelete IS NULL)
    GROUP BY idjadwalujian
) r
WHERE j.idjadwalujian = r.idjadwalujian
  AND j.tglmulai IS NULL;
