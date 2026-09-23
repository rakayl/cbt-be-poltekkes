-- Migration: Create Table cat.ep_proctor_snapshots
CREATE TABLE IF NOT EXISTS cat.ep_proctor_snapshots (
    id BIGSERIAL PRIMARY KEY,
    id_ep_schedule INTEGER NOT NULL,
    kodepeserta VARCHAR(50) NOT NULL,
    snapshot_type VARCHAR(50) NOT NULL DEFAULT 'PERIODIC', -- 'EXAM_START', 'PERIODIC', 'VIOLATION'
    trigger_event VARCHAR(50) DEFAULT 'INTERVAL_TIMER',   -- 'EXAM_START', 'INTERVAL_TIMER', 'TAB_SWITCH', 'WINDOW_BLUR', 'FULLSCREEN_EXIT', 'CAMERA_BLACKOUT'
    image_data TEXT NOT NULL,                             -- data:image/jpeg;base64,...
    risk_score INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ep_snapshots_sched_peserta ON cat.ep_proctor_snapshots(id_ep_schedule, kodepeserta);
CREATE INDEX IF NOT EXISTS idx_ep_snapshots_sched_created ON cat.ep_proctor_snapshots(id_ep_schedule, created_at DESC);
