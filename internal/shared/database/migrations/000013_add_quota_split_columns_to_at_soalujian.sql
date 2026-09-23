-- Add split quota columns for static questions and dynamic clinical stimuli
ALTER TABLE cat.at_soalujian ADD COLUMN IF NOT EXISTS jumlah_soal_statis integer DEFAULT 0;
ALTER TABLE cat.at_soalujian ADD COLUMN IF NOT EXISTS jumlah_kasus_dinamis integer DEFAULT 0;
