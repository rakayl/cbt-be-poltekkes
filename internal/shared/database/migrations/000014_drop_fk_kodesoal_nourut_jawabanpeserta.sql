-- Migration 000014: Drop foreign key constraint on at_jawabanpeserta (kodesoal, nourut)
-- This allows answers for dynamic stimuli and multi-bank questions to be safely saved without table constraint conflicts.

ALTER TABLE cat.at_jawabanpeserta DROP CONSTRAINT IF EXISTS at_jawabanpeserta_kodesoal_fkey;
