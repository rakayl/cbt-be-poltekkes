package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/english_proficiency/dto"
	"poltekkes-cat-backend/internal/modules/english_proficiency/entity"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EPRepository interface {
	// Exam Types & Sections
	GetExamTypes(ctx context.Context) ([]*entity.ExamType, error)
	GetSectionsByExamType(ctx context.Context, examTypeID int) ([]*entity.Section, error)
	GetSectionParts(ctx context.Context, sectionID int) ([]*entity.SectionPart, error)

	// Conversion Profiles
	GetConversionProfiles(ctx context.Context, examTypeID int) ([]*entity.ConversionProfile, error)
	GetConversionTable(ctx context.Context, profileID int) ([]*entity.ScoreConversion, error)
	SaveConversionProfile(ctx context.Context, req *dto.SaveConversionProfileRequestDTO) (*entity.ConversionProfile, error)

	// Exams
	CreateExam(ctx context.Context, exam *entity.EPExam, sections []dto.ExamSectionConfigDTO) (*entity.EPExam, error)
	GetExams(ctx context.Context, page, perPage int) ([]*entity.EPExam, int64, error)
	GetExamByID(ctx context.Context, examID int) (*entity.EPExam, error)
	UpdateExam(ctx context.Context, examID int, req *dto.CreateEPExamRequestDTO) error
	DeleteExam(ctx context.Context, examID int) error

	// Schedules
	CreateSchedule(ctx context.Context, schedule *entity.EPSchedule) (*entity.EPSchedule, error)
	GetSchedules(ctx context.Context, examID int) ([]*entity.EPSchedule, error)
	GetScheduleByID(ctx context.Context, scheduleID int) (*entity.EPSchedule, error)
	UpdateSchedule(ctx context.Context, scheduleID int, req *dto.UpdateEPScheduleRequestDTO) (*entity.EPSchedule, error)
	DeleteSchedule(ctx context.Context, scheduleID int) error
	RegisterParticipants(ctx context.Context, scheduleID int, codes []string, details []dto.RegisterParticipantItemDTO) (int, error)
	GetScheduleParticipants(ctx context.Context, scheduleID int) ([]*entity.ScheduleParticipant, error)

	// Exam Execution (Dynamic Section Session)
	StartParticipantSession(ctx context.Context, scheduleID int, participantCode, sessionToken string) (*entity.ScheduleParticipant, error)
	GetCurrentSectionSession(ctx context.Context, scheduleID int, participantCode string) (*dto.EPExamSessionResponseDTO, error)
	SaveAnswer(ctx context.Context, participantCode string, req *dto.SaveAnswerDTO) error
	AdvanceToNextSection(ctx context.Context, scheduleID int, participantCode string) (*dto.EPExamSessionResponseDTO, error)
	FinishExam(ctx context.Context, scheduleID int, participantCode, reason string) error
	RemoveParticipantFromSchedule(ctx context.Context, scheduleID int, kodepeserta string) error

	// Anti-Cheat, Security Events & Proctor Telemetry
	RecordSecurityEvent(ctx context.Context, scheduleID int, participantCode string, req *dto.EPSecurityEventRequestDTO, ip, userAgent string) (*dto.EPSecurityEventResponseDTO, error)
	SyncParticipantHeartbeat(ctx context.Context, scheduleID int, participantCode string, req *dto.EPHeartbeatRequestDTO) (*dto.EPHeartbeatResponseDTO, error)
	ApplyProctorAction(ctx context.Context, req *dto.EPProctorActionRequestDTO) error
	GetSecurityEventsByParticipant(ctx context.Context, scheduleID int, participantCode string) ([]*entity.EPSecurityEvent, error)
	SaveProctorSnapshot(ctx context.Context, scheduleID int, participantCode string, req *dto.UploadProctorSnapshotRequestDTO) (*dto.ProctorSnapshotItemDTO, error)
	GetProctorSnapshots(ctx context.Context, scheduleID int, participantCode, snapshotType string, limit, offset int) ([]dto.ProctorSnapshotItemDTO, error)
	GetLatestProctorGrid(ctx context.Context, scheduleID int) ([]dto.ProctorLatestGridItemDTO, error)

	// Scoring Engine
	CalculateAndSaveScores(ctx context.Context, scheduleID int) ([]*entity.EPScore, error)
	GetScoresBySchedule(ctx context.Context, scheduleID int) ([]*entity.EPScore, error)
	GetScoresByExam(ctx context.Context, examID int) ([]*entity.EPScore, error)

	// Certificates
	GenerateCertificates(ctx context.Context, scheduleID int) ([]*entity.EPCertificate, error)
	GetCertificateByNumber(ctx context.Context, certNumber string) (*entity.EPCertificate, error)
	VerifyCertificate(ctx context.Context, verificationCode string) (*entity.EPCertificate, error)
	GetParticipantScoresAndCertificates(ctx context.Context, participantCode string) ([]dto.ParticipantScoreResultDTO, error)

	// Question Banks
	GetEPQuestionBanks(ctx context.Context, examTypeID, sectionID int, search string, page, perPage int) ([]*entity.EPQuestionBank, int, error)
	GetEPQuestionBankByCode(ctx context.Context, kodesoal string) (*entity.EPQuestionBank, error)
	CreateEPQuestionBank(ctx context.Context, req *dto.CreateEPBankDTO) (*entity.EPQuestionBank, error)
	UpdateEPQuestionBank(ctx context.Context, kodesoal string, req *dto.UpdateEPBankDTO) error
	DeleteEPQuestionBank(ctx context.Context, kodesoal string) error
	CopyEPQuestionBank(ctx context.Context, req *dto.CopyEPBankDTO, sourceKode string) (*entity.EPQuestionBank, error)

	GetStimuliByBank(ctx context.Context, kodesoal string) ([]dto.DynamicStimulusDTO, error)
	CreateStimulusWithItems(ctx context.Context, req *dto.CreateStimulusRequestDTO) error
	ImportBankStimuli(ctx context.Context, kodesoal string, stimuli []dto.CreateStimulusRequestDTO) (int, int, error)
	UpdateStimulus(ctx context.Context, idStimulus int64, req *dto.UpdateStimulusRequestDTO) error
	DeleteStimulus(ctx context.Context, idStimulus int64) error

	CreateSubItem(ctx context.Context, stimulusID int64, req *dto.CreateSubItemRequestDTO) error
	UpdateSubItem(ctx context.Context, itemID int64, req *dto.UpdateSubItemRequestDTO) error
	DeleteSubItem(ctx context.Context, itemID int64) error

	// External Participants (Peserta Umum TOEFL)
	GetExternalParticipants(ctx context.Context, search string, limit, offset int) ([]*entity.EPPesertaUmum, int, error)
	GetExternalParticipantByCode(ctx context.Context, code string) (*entity.EPPesertaUmum, error)
	GetNextExternalParticipantCode(ctx context.Context) (string, error)
	CreateExternalParticipant(ctx context.Context, p *entity.EPPesertaUmum) error
	UpdateExternalParticipant(ctx context.Context, p *entity.EPPesertaUmum) error
	DeleteExternalParticipant(ctx context.Context, code string) error
	BatchImportExternalParticipants(ctx context.Context, list []*entity.EPPesertaUmum) (int, error)
}

type epRepository struct {
	db *sqlx.DB
}

func NewEPRepository(db *sqlx.DB) EPRepository {
	return &epRepository{db: db}
}

// ─────────────────────────────────────────────────────────────────────────────
// 1. Exam Types & Sections
// ─────────────────────────────────────────────────────────────────────────────
func (r *epRepository) GetExamTypes(ctx context.Context) ([]*entity.ExamType, error) {
	var list []*entity.ExamType
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM cat.ep_exam_types WHERE is_active = TRUE ORDER BY id_exam_type ASC`)
	return list, err
}

func (r *epRepository) GetSectionsByExamType(ctx context.Context, examTypeID int) ([]*entity.Section, error) {
	var list []*entity.Section
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM cat.ep_sections WHERE id_exam_type = $1 ORDER BY section_order ASC`, examTypeID)
	return list, err
}

func (r *epRepository) GetSectionParts(ctx context.Context, sectionID int) ([]*entity.SectionPart, error) {
	var list []*entity.SectionPart
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM cat.ep_section_parts WHERE id_section = $1 ORDER BY part_order ASC`, sectionID)
	return list, err
}

// ─────────────────────────────────────────────────────────────────────────────
// 2. Conversion Profiles
// ─────────────────────────────────────────────────────────────────────────────
func (r *epRepository) GetConversionProfiles(ctx context.Context, examTypeID int) ([]*entity.ConversionProfile, error) {
	var list []*entity.ConversionProfile
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM cat.ep_conversion_profiles WHERE id_exam_type = $1 ORDER BY is_default DESC, id_profile ASC`, examTypeID)
	return list, err
}

func (r *epRepository) GetConversionTable(ctx context.Context, profileID int) ([]*entity.ScoreConversion, error) {
	var list []*entity.ScoreConversion
	err := r.db.SelectContext(ctx, &list, `SELECT * FROM cat.ep_score_conversion WHERE id_profile = $1 ORDER BY section_code ASC, raw_score ASC`, profileID)
	return list, err
}

func (r *epRepository) SaveConversionProfile(ctx context.Context, req *dto.SaveConversionProfileRequestDTO) (*entity.ConversionProfile, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if req.IsDefault {
		_, _ = tx.ExecContext(ctx, `UPDATE cat.ep_conversion_profiles SET is_default = FALSE WHERE id_exam_type = $1`, req.IDExamType)
	}

	var profile entity.ConversionProfile
	insertQ := `
		INSERT INTO cat.ep_conversion_profiles (id_exam_type, profile_code, profile_name, is_default, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (id_exam_type, profile_code)
		DO UPDATE SET profile_name = EXCLUDED.profile_name, is_default = EXCLUDED.is_default, description = EXCLUDED.description, updated_at = NOW()
		RETURNING *
	`
	err = tx.GetContext(ctx, &profile, insertQ, req.IDExamType, req.ProfileCode, req.ProfileName, req.IsDefault, req.Description)
	if err != nil {
		return nil, err
	}

	// Delete old table entries and re-insert
	_, _ = tx.ExecContext(ctx, `DELETE FROM cat.ep_score_conversion WHERE id_profile = $1`, profile.IDProfile)

	insertTableQ := `INSERT INTO cat.ep_score_conversion (id_profile, section_code, raw_score, scaled_score) VALUES ($1, $2, $3, $4)`
	for _, item := range req.Tables {
		_, err = tx.ExecContext(ctx, insertTableQ, profile.IDProfile, item.SectionCode, item.RawScore, item.ScaledScore)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &profile, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 3. Exams
// ─────────────────────────────────────────────────────────────────────────────
func (r *epRepository) CreateExam(ctx context.Context, exam *entity.EPExam, sections []dto.ExamSectionConfigDTO) (*entity.EPExam, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	insertQ := `
		INSERT INTO cat.ep_exams (id_exam_type, id_profile, id_template, exam_name, passing_score, validity_months_override, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING *
	`
	err = tx.GetContext(ctx, exam, insertQ, exam.IDExamType, exam.IDProfile, exam.IDTemplate, exam.ExamName, exam.PassingScore, exam.ValidityMonthsOverride, exam.Description, exam.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat ujian: %w", err)
	}

	insertSecQ := `
		INSERT INTO cat.ep_exam_sections (id_ep_exam, id_section, kodesoal, section_order, duration_minutes, allow_replay, max_replay_count, shuffle_options, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	for idx, s := range sections {
		order := s.SectionOrder
		if order == 0 {
			order = idx + 1
		}
		_, err = tx.ExecContext(ctx, insertSecQ, exam.IDEPExam, s.IDSection, s.KodeSoal, order, s.DurationMinutes, s.AllowReplay, s.MaxReplayCount, s.ShuffleOptions)
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan konfigurasi seksi: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return exam, nil
}

func (r *epRepository) GetExams(ctx context.Context, page, perPage int) ([]*entity.EPExam, int64, error) {
	offset := (page - 1) * perPage
	var total int64
	_ = r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM cat.ep_exams WHERE is_active = TRUE`)

	query := `
		SELECT e.*, t.type_name as exam_type_name, p.profile_name
		FROM cat.ep_exams e
		JOIN cat.ep_exam_types t ON t.id_exam_type = e.id_exam_type
		LEFT JOIN cat.ep_conversion_profiles p ON p.id_profile = e.id_profile
		WHERE e.is_active = TRUE
		ORDER BY e.id_ep_exam DESC
		LIMIT $1 OFFSET $2
	`
	var exams []*entity.EPExam
	err := r.db.SelectContext(ctx, &exams, query, perPage, offset)
	if err != nil {
		return nil, 0, err
	}

	for _, e := range exams {
		secQuery := `
			SELECT es.*, s.section_code, s.section_name, s.has_audio
			FROM cat.ep_exam_sections es
			JOIN cat.ep_sections s ON s.id_section = es.id_section
			WHERE es.id_ep_exam = $1
			ORDER BY es.section_order ASC
		`
		_ = r.db.SelectContext(ctx, &e.Sections, secQuery, e.IDEPExam)
	}

	return exams, total, nil
}

func (r *epRepository) GetExamByID(ctx context.Context, examID int) (*entity.EPExam, error) {
	var e entity.EPExam
	query := `
		SELECT e.*, t.type_name as exam_type_name, p.profile_name
		FROM cat.ep_exams e
		JOIN cat.ep_exam_types t ON t.id_exam_type = e.id_exam_type
		LEFT JOIN cat.ep_conversion_profiles p ON p.id_profile = e.id_profile
		WHERE e.id_ep_exam = $1
	`
	err := r.db.GetContext(ctx, &e, query, examID)
	if err != nil {
		return nil, err
	}

	secQuery := `
		SELECT es.*, s.section_code, s.section_name, s.has_audio
		FROM cat.ep_exam_sections es
		JOIN cat.ep_sections s ON s.id_section = es.id_section
		WHERE es.id_ep_exam = $1
		ORDER BY es.section_order ASC
	`
	_ = r.db.SelectContext(ctx, &e.Sections, secQuery, e.IDEPExam)
	return &e, nil
}

func (r *epRepository) UpdateExam(ctx context.Context, examID int, req *dto.CreateEPExamRequestDTO) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateQ := `
		UPDATE cat.ep_exams
		SET id_exam_type = $1, id_profile = $2, id_template = $3, exam_name = $4,
		    passing_score = $5, validity_months_override = $6, description = $7, updated_at = NOW()
		WHERE id_ep_exam = $8
	`
	_, err = tx.ExecContext(ctx, updateQ, req.IDExamType, req.IDProfile, req.IDTemplate, req.ExamName, req.PassingScore, req.ValidityMonthsOverride, req.Description, examID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui data ujian: %w", err)
	}

	// Delete old sections and reinsert new configured sections
	_, err = tx.ExecContext(ctx, `DELETE FROM cat.ep_exam_sections WHERE id_ep_exam = $1`, examID)
	if err != nil {
		return fmt.Errorf("gagal mereset konfigurasi seksi lama: %w", err)
	}

	insertSecQ := `
		INSERT INTO cat.ep_exam_sections (id_ep_exam, id_section, kodesoal, section_order, duration_minutes, allow_replay, max_replay_count, shuffle_options, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	for idx, s := range req.Sections {
		order := s.SectionOrder
		if order == 0 {
			order = idx + 1
		}
		_, err = tx.ExecContext(ctx, insertSecQ, examID, s.IDSection, s.KodeSoal, order, s.DurationMinutes, s.AllowReplay, s.MaxReplayCount, s.ShuffleOptions)
		if err != nil {
			return fmt.Errorf("gagal menyimpan konfigurasi seksi baru: %w", err)
		}
	}

	return tx.Commit()
}

func (r *epRepository) DeleteExam(ctx context.Context, examID int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cat.ep_exams SET is_active = FALSE, updated_at = NOW() WHERE id_ep_exam = $1`, examID)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// 4. Schedules & Participants
// ─────────────────────────────────────────────────────────────────────────────
func (r *epRepository) CreateSchedule(ctx context.Context, schedule *entity.EPSchedule) (*entity.EPSchedule, error) {
	insertQ := `
		INSERT INTO cat.ep_schedules (id_ep_exam, id_ruang, exam_date, start_time, end_time, session_token, capacity, proctor_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4::time, $5::time, $6, $7, $8, TRUE, NOW(), NOW())
		RETURNING id_ep_schedule, id_ep_exam, id_ruang, exam_date, 
		          TO_CHAR(start_time, 'HH24:MI') as start_time, 
		          TO_CHAR(end_time, 'HH24:MI') as end_time, 
		          session_token, capacity, proctor_name, is_active, created_at, updated_at
	`
	err := r.db.GetContext(ctx, schedule, insertQ, schedule.IDEPExam, schedule.IDRuang, schedule.ExamDate, schedule.StartTime, schedule.EndTime, schedule.SessionToken, schedule.Capacity, schedule.ProctorName)
	return schedule, err
}

func (r *epRepository) GetSchedules(ctx context.Context, examID int) ([]*entity.EPSchedule, error) {
	query := `
		SELECT s.id_ep_schedule, s.id_ep_exam, s.id_ruang, s.exam_date,
		       TO_CHAR(s.start_time, 'HH24:MI') as start_time,
		       TO_CHAR(s.end_time, 'HH24:MI') as end_time,
		       s.session_token, s.capacity, s.proctor_name, s.is_active, s.created_at, s.updated_at,
		       e.exam_name, COALESCE(r.namaruang, 'Ruang ' || s.id_ruang::text) as room_name,
		       COUNT(p.id_participant)::integer as registered
		FROM cat.ep_schedules s
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		LEFT JOIN cat.at_ruang r ON (r.koderuang = s.id_ruang::text OR r.koderuang = LPAD(s.id_ruang::text, 3, '0'))
		LEFT JOIN cat.ep_schedule_participants p ON p.id_ep_schedule = s.id_ep_schedule AND p.softdelete = '0'
		WHERE ($1 = 0 OR s.id_ep_exam = $1) AND s.is_active = TRUE
		GROUP BY s.id_ep_schedule, e.exam_name, r.namaruang
		ORDER BY s.exam_date DESC, s.start_time ASC
	`
	var list []*entity.EPSchedule
	err := r.db.SelectContext(ctx, &list, query, examID)
	return list, err
}

func (r *epRepository) GetScheduleByID(ctx context.Context, scheduleID int) (*entity.EPSchedule, error) {
	query := `
		SELECT s.id_ep_schedule, s.id_ep_exam, s.id_ruang, s.exam_date,
		       TO_CHAR(s.start_time, 'HH24:MI') as start_time,
		       TO_CHAR(s.end_time, 'HH24:MI') as end_time,
		       s.session_token, s.capacity, s.proctor_name, s.is_active, s.created_at, s.updated_at,
		       e.exam_name, COALESCE(r.namaruang, 'Ruang ' || s.id_ruang::text) as room_name,
		       COUNT(p.id_participant)::integer as registered
		FROM cat.ep_schedules s
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		LEFT JOIN cat.at_ruang r ON (r.koderuang = s.id_ruang::text OR r.koderuang = LPAD(s.id_ruang::text, 3, '0'))
		LEFT JOIN cat.ep_schedule_participants p ON p.id_ep_schedule = s.id_ep_schedule AND (p.softdelete = '0' OR p.softdelete IS NULL)
		WHERE s.id_ep_schedule = $1
		GROUP BY s.id_ep_schedule, e.exam_name, r.namaruang
	`
	var s entity.EPSchedule
	err := r.db.GetContext(ctx, &s, query, scheduleID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *epRepository) UpdateSchedule(ctx context.Context, scheduleID int, req *dto.UpdateEPScheduleRequestDTO) (*entity.EPSchedule, error) {
	query := `
		UPDATE cat.ep_schedules
		SET id_ruang = $1,
		    exam_date = $2,
		    start_time = $3::time,
		    end_time = $4::time,
		    session_token = $5,
		    capacity = $6,
		    proctor_name = $7,
		    is_active = COALESCE($8, is_active),
		    updated_at = NOW()
		WHERE id_ep_schedule = $9
	`
	res, err := r.db.ExecContext(ctx, query, req.IDRuang, req.ExamDate, req.StartTime, req.EndTime, req.SessionToken, req.Capacity, req.ProctorName, req.IsActive, scheduleID)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("jadwal tidak ditemukan")
	}

	return r.GetScheduleByID(ctx, scheduleID)
}

func (r *epRepository) DeleteSchedule(ctx context.Context, scheduleID int) error {
	var count int
	_ = r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) 
		FROM cat.ep_schedule_participants 
		WHERE id_ep_schedule = $1 
		  AND (softdelete = '0' OR softdelete IS NULL) 
		  AND session_status NOT IN ('NOT_STARTED')
	`, scheduleID)
	if count > 0 {
		return fmt.Errorf("sesi tidak dapat dihapus karena ada %d peserta yang sudah mengikuti/menyelesaikan ujian", count)
	}

	// Soft delete participants in this schedule
	_, _ = r.db.ExecContext(ctx, `UPDATE cat.ep_schedule_participants SET softdelete = '1', updated_at = NOW() WHERE id_ep_schedule = $1`, scheduleID)

	// Soft delete the schedule
	_, err := r.db.ExecContext(ctx, `UPDATE cat.ep_schedules SET is_active = FALSE, updated_at = NOW() WHERE id_ep_schedule = $1`, scheduleID)
	return err
}

func (r *epRepository) RegisterParticipants(ctx context.Context, scheduleID int, codes []string, details []dto.RegisterParticipantItemDTO) (int, error) {
	inserted := 0

	// ── Cross-schedule check: ensure participant is not already active in another session of the same exam ──
	crossCheckQ := `
		SELECT sp.session_status, s2.id_ep_schedule
		FROM cat.ep_schedule_participants sp
		JOIN cat.ep_schedules s2 ON s2.id_ep_schedule = sp.id_ep_schedule
		WHERE sp.kodepeserta = $1
		  AND s2.id_ep_exam = (SELECT id_ep_exam FROM cat.ep_schedules WHERE id_ep_schedule = $2)
		  AND sp.id_ep_schedule != $2
		  AND (sp.softdelete = '0' OR sp.softdelete IS NULL)
		LIMIT 1
	`

	upsertPesertaQ := `
		INSERT INTO cat.ep_peserta_umum (kodepeserta, nik, nama, jenis_kelamin, email, hp, instansi, password, is_active, updated_at)
		VALUES (
			$1,
			$1,
			$2,
			'L',
			COALESCE(NULLIF($3, ''), $1 || '@ep.poltekkes-sby.ac.id'),
			COALESCE(NULLIF($4, ''), '08000000000'),
			$5,
			md5($6),
			TRUE,
			NOW()
		)
		ON CONFLICT (kodepeserta) DO UPDATE SET
			nama = COALESCE(NULLIF(EXCLUDED.nama, ''), cat.ep_peserta_umum.nama),
			email = COALESCE(NULLIF(EXCLUDED.email, ''), cat.ep_peserta_umum.email),
			hp = COALESCE(NULLIF(EXCLUDED.hp, ''), cat.ep_peserta_umum.hp),
			instansi = COALESCE(EXCLUDED.instansi, cat.ep_peserta_umum.instansi),
			is_active = TRUE,
			updated_at = NOW()
	`

	insertQ := `
		INSERT INTO cat.ep_schedule_participants (id_ep_schedule, kodepeserta, session_status, softdelete, created_at, updated_at)
		VALUES ($1, $2, 'NOT_STARTED', '0', NOW(), NOW())
		ON CONFLICT (id_ep_schedule, kodepeserta) 
		DO UPDATE SET softdelete = '0', updated_at = NOW()
		WHERE cat.ep_schedule_participants.softdelete = '1'
	`

	registerOne := func(kodepeserta string) bool {
		var existing struct {
			SessionStatus string `db:"session_status"`
			IDEPSchedule  int    `db:"id_ep_schedule"`
		}
		err := r.db.GetContext(ctx, &existing, crossCheckQ, kodepeserta, scheduleID)
		if err == nil {
			// Found in another session of the same exam
			fmt.Printf("[EP INFO] Peserta %s ditolak: sudah terdaftar di sesi %d (status: %s)\n", kodepeserta, existing.IDEPSchedule, existing.SessionStatus)
			return false
		}
		return true
	}

	// 1. Process details if provided
	for _, item := range details {
		if item.KodePeserta == "" {
			continue
		}
		if !registerOne(item.KodePeserta) {
			continue
		}

		var alamat *string
		if item.Institusi != "" {
			alamat = &item.Institusi
		}
		namaVal := item.Nama
		if namaVal == "" {
			namaVal = "Peserta " + item.KodePeserta
		}
		if _, err := r.db.ExecContext(ctx, upsertPesertaQ, item.KodePeserta, namaVal, item.Email, item.HP, alamat, item.KodePeserta); err != nil {
			fmt.Printf("[EP ERROR upsert at_peserta]: %v\n", err)
		}

		res, err := r.db.ExecContext(ctx, insertQ, scheduleID, item.KodePeserta)
		if err == nil {
			if n, _ := res.RowsAffected(); n > 0 {
				inserted++
			}
		}
	}

	// 2. Process codes if provided
	for _, code := range codes {
		if code == "" {
			continue
		}
		if !registerOne(code) {
			continue
		}

		if _, err := r.db.ExecContext(ctx, upsertPesertaQ, code, "Peserta "+code, nil, nil, nil, code); err != nil {
			fmt.Printf("[EP ERROR upsert codes at_peserta]: %v\n", err)
		}
		res, err := r.db.ExecContext(ctx, insertQ, scheduleID, code)
		if err == nil {
			if n, _ := res.RowsAffected(); n > 0 {
				inserted++
			}
		}
	}

	return inserted, nil
}

func (r *epRepository) GetScheduleParticipants(ctx context.Context, scheduleID int) ([]*entity.ScheduleParticipant, error) {
	query := `
		SELECT sp.*, COALESCE(pu.nama, p.nama, 'Peserta ' || sp.kodepeserta) as nama,
		       COALESCE(pu.email, p.email) as email,
		       COALESCE(pu.hp, p.hp) as hp,
		       COALESCE(pu.instansi, p.alamat) as alamat
		FROM cat.ep_schedule_participants sp
		LEFT JOIN cat.ep_peserta_umum pu ON pu.kodepeserta = sp.kodepeserta
		LEFT JOIN cat.at_peserta p ON p.kodepeserta = sp.kodepeserta
		WHERE sp.id_ep_schedule = $1 AND (sp.softdelete = '0' OR sp.softdelete IS NULL)
		ORDER BY sp.kodepeserta ASC
	`
	var list []*entity.ScheduleParticipant
	err := r.db.SelectContext(ctx, &list, query, scheduleID)
	return list, err
}

// ─────────────────────────────────────────────────────────────────────────────
// 5. Test Taker Exam Session (Dynamic Section Countdown & Replay)
// ─────────────────────────────────────────────────────────────────────────────
func (r *epRepository) StartParticipantSession(ctx context.Context, scheduleID int, participantCode, sessionToken string) (*entity.ScheduleParticipant, error) {
	// Verify schedule & token
	var sched entity.EPSchedule
	schedQ := `
		SELECT id_ep_schedule, id_ep_exam, id_ruang, exam_date,
		       TO_CHAR(start_time, 'HH24:MI') as start_time,
		       TO_CHAR(end_time, 'HH24:MI') as end_time,
		       session_token, capacity, proctor_name, is_active, created_at, updated_at
		FROM cat.ep_schedules
		WHERE id_ep_schedule = $1 AND is_active = TRUE
	`
	err := r.db.GetContext(ctx, &sched, schedQ, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("jadwal ujian tidak ditemukan: %w", err)
	}
	if sched.SessionToken != "" && sessionToken != "" && sched.SessionToken != sessionToken {
		return nil, fmt.Errorf("token ujian tidak valid")
	}

	// ── Enforce session time window (WIB) ──
	locWIB, _ := time.LoadLocation("Asia/Jakarta")
	if locWIB == nil {
		locWIB = time.FixedZone("WIB", 7*3600)
	}
	nowWIB := time.Now().In(locWIB)
	examDateStr := sched.ExamDate.Format("2006-01-02")
	startTimeStr := sched.StartTime
	if len(startTimeStr) >= 5 {
		startTimeStr = startTimeStr[:5]
	}
	endTimeStr := sched.EndTime
	if len(endTimeStr) >= 5 {
		endTimeStr = endTimeStr[:5]
	}

	schedStart, errS := time.ParseInLocation("2006-01-02 15:04", examDateStr+" "+startTimeStr, locWIB)
	schedEnd, errE := time.ParseInLocation("2006-01-02 15:04", examDateStr+" "+endTimeStr, locWIB)

	if errS == nil && errE == nil {
		if nowWIB.Before(schedStart) {
			return nil, fmt.Errorf("Sesi ujian belum dibuka. Sesi akan dimulai pada %s pukul %s WIB",
				sched.ExamDate.Format("02 Jan 2006"), startTimeStr)
		}
		if nowWIB.After(schedEnd) {
			return nil, fmt.Errorf("Sesi ujian telah berakhir pada pukul %s WIB. Anda tidak dapat memulai ujian", endTimeStr)
		}
	}

	// Fetch participant record
	var p entity.ScheduleParticipant
	err = r.db.GetContext(ctx, &p, `SELECT * FROM cat.ep_schedule_participants WHERE id_ep_schedule = $1 AND kodepeserta = $2 AND (softdelete = '0' OR softdelete IS NULL)`, scheduleID, participantCode)
	if err != nil {
		return nil, fmt.Errorf("peserta belum terdaftar pada jadwal ujian ini")
	}
	if p.IsLocked {
		return nil, fmt.Errorf("sesi ujian Anda telah dikunci oleh pengawas: %s", *p.LockReason)
	}
	if p.SessionStatus == "SUBMITTED" {
		return nil, fmt.Errorf("Anda telah menyelesaikan ujian ini")
	}

	// If not started -> Initialize section 1 timer
	if p.SessionStatus == "NOT_STARTED" || p.StartedAt == nil {
		// Get Section 1 duration
		var sec1 struct {
			DurationMinutes int `db:"duration_minutes"`
		}
		err = r.db.GetContext(ctx, &sec1, `
			SELECT duration_minutes FROM cat.ep_exam_sections
			WHERE id_ep_exam = $1 AND section_order = 1
		`, sched.IDEPExam)
		duration := 35
		if err == nil && sec1.DurationMinutes > 0 {
			duration = sec1.DurationMinutes
		}

		now := time.Now()
		deadline := now.Add(time.Duration(duration) * time.Minute)

		// Cap section deadline to session end time so it does not exceed the session window
		if errE == nil && deadline.In(locWIB).After(schedEnd) {
			deadline = schedEnd
		}

		updateQ := `
			UPDATE cat.ep_schedule_participants
			SET session_status = 'IN_PROGRESS', started_at = $1, current_section = 1,
			    section_started_at = $1, section_deadline_at = $2, updated_at = NOW()
			WHERE id_participant = $3
			RETURNING *
		`
		err = r.db.GetContext(ctx, &p, updateQ, now, deadline, p.IDParticipant)
	}

	return &p, err
}

func (r *epRepository) GetCurrentSectionSession(ctx context.Context, scheduleID int, participantCode string) (*dto.EPExamSessionResponseDTO, error) {
	// 1. Get Schedule, Exam & Participant
	var info struct {
		IDParticipant     int        `db:"id_participant"`
		IDEPExam          int        `db:"id_ep_exam"`
		ExamName          string     `db:"exam_name"`
		ParticipantName   string     `db:"participant_name"`
		CurrentSection    int        `db:"current_section"`
		SectionStartedAt  *time.Time `db:"section_started_at"`
		SectionDeadlineAt *time.Time `db:"section_deadline_at"`
		SessionStatus     string     `db:"session_status"`
		IsLocked          bool       `db:"is_locked"`
		LockReason        *string    `db:"lock_reason"`
	}

	query := `
		SELECT sp.id_participant, s.id_ep_exam, e.exam_name,
		       COALESCE(pu.nama, p.nama, 'Peserta ' || sp.kodepeserta) as participant_name,
		       COALESCE(sp.current_section, 1) as current_section,
		       sp.section_started_at, sp.section_deadline_at, sp.session_status,
		       sp.is_locked, sp.lock_reason
		FROM cat.ep_schedule_participants sp
		JOIN cat.ep_schedules s ON s.id_ep_schedule = sp.id_ep_schedule
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		LEFT JOIN cat.ep_peserta_umum pu ON pu.kodepeserta = sp.kodepeserta
		LEFT JOIN cat.at_peserta p ON p.kodepeserta = sp.kodepeserta
		WHERE sp.id_ep_schedule = $1 AND sp.kodepeserta = $2
	`
	err := r.db.GetContext(ctx, &info, query, scheduleID, participantCode)
	if err != nil {
		return nil, fmt.Errorf("sesi peserta tidak ditemukan: %w", err)
	}

	if info.IsLocked {
		return nil, fmt.Errorf("sesi Anda dikunci oleh pengawas")
	}
	if info.SessionStatus == "SUBMITTED" {
		return nil, fmt.Errorf("lembar ujian telah selesai dikerjakan")
	}

	// 2. Fetch section config for current section
	var secConfig struct {
		SectionCode     string `db:"section_code"`
		SectionName     string `db:"section_name"`
		KodeSoal        string `db:"kodesoal"`
		DurationMinutes int    `db:"duration_minutes"`
		AllowReplay     bool   `db:"allow_replay"`
		MaxReplayCount  int    `db:"max_replay_count"`
		ShuffleOptions  bool   `db:"shuffle_options"`
		HasAudio        bool   `db:"has_audio"`
		CanGoBack       bool   `db:"can_go_back"`
	}
	secQ := `
		SELECT s.section_code, s.section_name, es.kodesoal, es.duration_minutes,
		       es.allow_replay, es.max_replay_count, es.shuffle_options, s.has_audio, s.can_go_back
		FROM cat.ep_exam_sections es
		JOIN cat.ep_sections s ON s.id_section = es.id_section
		WHERE es.id_ep_exam = $1 AND es.section_order = $2
	`
	err = r.db.GetContext(ctx, &secConfig, secQ, info.IDEPExam, info.CurrentSection)
	if err != nil {
		return nil, fmt.Errorf("konfigurasi seksi ujian tidak ditemukan: %w", err)
	}

	now := time.Now()
	var remainingSec int
	if info.SectionDeadlineAt != nil {
		remainingSec = int(info.SectionDeadlineAt.Sub(now).Seconds())
	} else {
		remainingSec = secConfig.DurationMinutes * 60
	}
	if remainingSec < 0 {
		remainingSec = 0
	}

	// 3. Fetch Dynamic Stimuli and Sub-Items from cat.cat_bank_stimulus for this section's kodesoal
	type rawStimulus struct {
		IDStimulus    int64           `db:"id_stimulus"`
		StimulusCode  string          `db:"stimulus_code"`
		Title         string          `db:"title"`
		NarrativeText string          `db:"narrative_text"`
		MediaType     string          `db:"media_type"`
		MediaURL      *string         `db:"media_url"`
		MediaMetadata json.RawMessage `db:"media_metadata"`
		CategoryTag   string          `db:"category_tag"`
	}
	var rawStimuli []rawStimulus
	stimQuery := `
		SELECT id_stimulus, stimulus_code, title, narrative_text, media_type, media_url, media_metadata, category_tag
		FROM cat.cat_bank_stimulus
		WHERE kodesoal = $1 AND (softdelete = '0' OR softdelete IS NULL)
		ORDER BY sort_order ASC, id_stimulus ASC
	`
	_ = r.db.SelectContext(ctx, &rawStimuli, stimQuery, secConfig.KodeSoal)

	// Fetch existing answers from cat.cat_participant_answers_multi
	type existingAns struct {
		IDItem          int64  `db:"id_item"`
		SelectedOptions string `db:"selected_options"`
		IsDoubtful      bool   `db:"is_doubtful"`
	}
	var answers []existingAns
	ansQuery := `
		SELECT id_item, selected_options, is_doubtful
		FROM cat.cat_participant_answers_multi
		WHERE id_session = $1
	`
	_ = r.db.SelectContext(ctx, &answers, ansQuery, info.IDParticipant)
	ansMap := make(map[int64]existingAns)
	for _, a := range answers {
		ansMap[a.IDItem] = a
	}

	totalQuestions := 0
	totalAnswered := 0
	var stimuliDTOs []dto.DynamicStimulusDTO

	stimulusNumber := 1
	for _, s := range rawStimuli {
		// Fetch child items
		type rawItem struct {
			IDItem            int64   `db:"id_item"`
			ItemOrder         int     `db:"item_order"`
			QuestionText      string  `db:"question_text"`
			QuestionMediaType string  `db:"question_media_type"`
			QuestionMediaURL  *string `db:"question_media_url"`
		}
		var items []rawItem
		itemQ := `
			SELECT id_item, item_order, question_text, question_media_type, question_media_url
			FROM cat.cat_stimulus_items
			WHERE id_stimulus = $1 AND (softdelete = '0' OR softdelete IS NULL)
			ORDER BY item_order ASC, id_item ASC
		`
		_ = r.db.SelectContext(ctx, &items, itemQ, s.IDStimulus)

		var itemDTOs []dto.DynamicItemDTO
		for _, it := range items {
			totalQuestions++

			// Fetch options
			type rawOpt struct {
				IDOption    int64   `db:"id_option"`
				OptionLabel string  `db:"option_label"`
				OptionText  string  `db:"option_text"`
				MediaType   string  `db:"media_type"`
				MediaURL    *string `db:"media_url"`
			}
			var opts []rawOpt
			optQ := `
				SELECT id_option, option_label, option_text, media_type, media_url
				FROM cat.cat_item_options
				WHERE id_item = $1
				ORDER BY sort_order ASC, option_label ASC
			`
			_ = r.db.SelectContext(ctx, &opts, optQ, it.IDItem)

			// Option Shuffling: Shuffle choices deterministically per participant & question item
			shuffledOpts := make([]rawOpt, len(opts))
			copy(shuffledOpts, opts)
			if secConfig.ShuffleOptions && len(shuffledOpts) > 1 {
				seed := int64(info.IDParticipant)*100003 + it.IDItem*97
				rng := rand.New(rand.NewSource(seed))
				rng.Shuffle(len(shuffledOpts), func(i, j int) {
					shuffledOpts[i], shuffledOpts[j] = shuffledOpts[j], shuffledOpts[i]
				})
			}

			var optDTOs []dto.DynamicOptionDTO
			for idx, o := range shuffledOpts {
				displayLabel := string(rune('A' + idx))
				optDTOs = append(optDTOs, dto.DynamicOptionDTO{
					IDOption:    o.IDOption,
					OptionLabel: displayLabel,
					OptionText:  o.OptionText,
					MediaType:   o.MediaType,
					MediaURL:    o.MediaURL,
					OriginalKey: o.OptionLabel,
				})
			}

			var userAns *string
			var doubtful bool
			if ans, exists := ansMap[it.IDItem]; exists {
				userAns = &ans.SelectedOptions
				doubtful = ans.IsDoubtful
				if ans.SelectedOptions != "" {
					totalAnswered++
				}
			}

			itemDTOs = append(itemDTOs, dto.DynamicItemDTO{
				IDItem:             it.IDItem,
				ItemOrder:          it.ItemOrder,
				QuestionText:       it.QuestionText,
				QuestionMediaType:  it.QuestionMediaType,
				QuestionMediaURL:   it.QuestionMediaURL,
				Options:            optDTOs,
				UserSelectedOption: userAns,
				IsDoubtful:         doubtful,
			})
		}

		var metaMap map[string]any
		_ = json.Unmarshal(s.MediaMetadata, &metaMap)

		stimuliDTOs = append(stimuliDTOs, dto.DynamicStimulusDTO{
			IDStimulus:     s.IDStimulus,
			StimulusNumber: stimulusNumber,
			StimulusCode:   s.StimulusCode,
			Title:          s.Title,
			NarrativeText:  s.NarrativeText,
			MediaType:      s.MediaType,
			MediaURL:       s.MediaURL,
			MediaMetadata:  metaMap,
			CategoryTag:    s.CategoryTag,
			SubItems:       itemDTOs,
		})
		stimulusNumber++
	}

	locWIB, _ := time.LoadLocation("Asia/Jakarta")
	if locWIB == nil {
		locWIB = time.FixedZone("WIB", 7*3600)
	}

	var totalSectionsCount int
	_ = r.db.GetContext(ctx, &totalSectionsCount, `SELECT COUNT(*) FROM cat.ep_exam_sections WHERE id_ep_exam = $1`, info.IDEPExam)
	if totalSectionsCount == 0 {
		totalSectionsCount = 1
	}

	return &dto.EPExamSessionResponseDTO{
		ScheduleID:          scheduleID,
		ExamName:            info.ExamName,
		ParticipantCode:     participantCode,
		ParticipantName:     info.ParticipantName,
		TotalSections:       totalSectionsCount,
		CurrentSectionOrder: info.CurrentSection,
		CurrentSectionCode:  secConfig.SectionCode,
		CurrentSectionName:  secConfig.SectionName,
		DurationMinutes:     secConfig.DurationMinutes,
		RemainingSeconds:    remainingSec,
		AllowReplay:         secConfig.AllowReplay,
		MaxReplayCount:      secConfig.MaxReplayCount,
		ShuffleOptions:      secConfig.ShuffleOptions,
		HasAudio:            secConfig.HasAudio,
		CanGoBack:           secConfig.CanGoBack,
		SessionStatus:       info.SessionStatus,
		ServerTimeWIB:       now.In(locWIB).Format("2006-01-02 15:04:05"),
		TotalQuestions:      totalQuestions,
		TotalAnswered:       totalAnswered,
		Stimuli:             stimuliDTOs,
	}, nil
}

func (r *epRepository) SaveAnswer(ctx context.Context, participantCode string, req *dto.SaveAnswerDTO) error {
	var part entity.ScheduleParticipant
	err := r.db.GetContext(ctx, &part, `
		SELECT * FROM cat.ep_schedule_participants
		WHERE id_ep_schedule = $1 AND kodepeserta = $2
	`, req.ScheduleID, participantCode)
	if err != nil {
		return fmt.Errorf("peserta tidak ditemukan")
	}

	// UPSERT to cat.cat_participant_answers_multi
	upsertQ := `
		INSERT INTO cat.cat_participant_answers_multi (id_session, id_stimulus, id_item, selected_options, is_doubtful, answered_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id_session, id_item)
		DO UPDATE SET selected_options = EXCLUDED.selected_options, is_doubtful = EXCLUDED.is_doubtful, answered_at = NOW()
	`
	_, err = r.db.ExecContext(ctx, upsertQ, part.IDParticipant, req.IDStimulus, req.IDItem, req.SelectedOption, req.IsDoubtful)
	return err
}

func (r *epRepository) AdvanceToNextSection(ctx context.Context, scheduleID int, participantCode string) (*dto.EPExamSessionResponseDTO, error) {
	var part entity.ScheduleParticipant
	err := r.db.GetContext(ctx, &part, `
		SELECT sp.*
		FROM cat.ep_schedule_participants sp
		WHERE sp.id_ep_schedule = $1 AND sp.kodepeserta = $2
	`, scheduleID, participantCode)
	if err != nil {
		return nil, fmt.Errorf("peserta tidak ditemukan")
	}

	nextSec := part.CurrentSection + 1

	// Check if next section exists in ep_exam_sections
	var nextSecConfig struct {
		DurationMinutes int `db:"duration_minutes"`
	}
	var examID int
	_ = r.db.GetContext(ctx, &examID, `SELECT id_ep_exam FROM cat.ep_schedules WHERE id_ep_schedule = $1`, scheduleID)

	err = r.db.GetContext(ctx, &nextSecConfig, `
		SELECT duration_minutes FROM cat.ep_exam_sections
		WHERE id_ep_exam = $1 AND section_order = $2
	`, examID, nextSec)

	if err != nil {
		// No more sections -> finish exam
		_ = r.FinishExam(ctx, scheduleID, participantCode, "COMPLETED_ALL_SECTIONS")
		return nil, fmt.Errorf("seluruh seksi ujian telah selesai dikerjakan")
	}

	// Advance and set new dynamic timer for next section
	now := time.Now()
	deadline := now.Add(time.Duration(nextSecConfig.DurationMinutes) * time.Minute)

	updateQ := `
		UPDATE cat.ep_schedule_participants
		SET current_section = $1, section_started_at = $2, section_deadline_at = $3, updated_at = NOW()
		WHERE id_participant = $4
	`
	_, err = r.db.ExecContext(ctx, updateQ, nextSec, now, deadline, part.IDParticipant)
	if err != nil {
		return nil, err
	}

	return r.GetCurrentSectionSession(ctx, scheduleID, participantCode)
}

func (r *epRepository) FinishExam(ctx context.Context, scheduleID int, participantCode, reason string) error {
	updateQ := `
		UPDATE cat.ep_schedule_participants
		SET session_status = 'SUBMITTED', finished_at = NOW(), updated_at = NOW()
		WHERE id_ep_schedule = $1 AND kodepeserta = $2
	`
	_, err := r.db.ExecContext(ctx, updateQ, scheduleID, participantCode)
	if err != nil {
		return err
	}
	fmt.Printf("[EP INFO] Sesi ujian diselesaikan: schedule=%d, peserta=%s, reason=%s\n", scheduleID, participantCode, reason)

	// Automatically calculate scores for schedule so ep_scores is populated immediately
	_, _ = r.CalculateAndSaveScores(ctx, scheduleID)

	return nil
}

func (r *epRepository) RemoveParticipantFromSchedule(ctx context.Context, scheduleID int, kodepeserta string) error {
	// Only allow removal if session is NOT_STARTED
	var status string
	err := r.db.GetContext(ctx, &status,
		`SELECT session_status FROM cat.ep_schedule_participants
		 WHERE id_ep_schedule = $1 AND kodepeserta = $2 AND (softdelete = '0' OR softdelete IS NULL)`,
		scheduleID, kodepeserta)
	if err != nil {
		return fmt.Errorf("peserta tidak ditemukan pada sesi ujian ini")
	}
	if status == "IN_PROGRESS" {
		return fmt.Errorf("peserta sedang aktif mengerjakan ujian, tidak dapat dihapus dari sesi ini")
	}
	if status == "SUBMITTED" {
		return fmt.Errorf("peserta sudah menyelesaikan ujian, riwayat tidak dapat dihapus")
	}

	// Soft-delete participant from this schedule
	_, err = r.db.ExecContext(ctx,
		`UPDATE cat.ep_schedule_participants SET softdelete = '1', updated_at = NOW()
		 WHERE id_ep_schedule = $1 AND kodepeserta = $2`, scheduleID, kodepeserta)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// 5B. Anti-Cheat, Security Events & Proctor Telemetry
// ─────────────────────────────────────────────────────────────────────────────

func (r *epRepository) RecordSecurityEvent(ctx context.Context, scheduleID int, participantCode string, req *dto.EPSecurityEventRequestDTO, ip, userAgent string) (*dto.EPSecurityEventResponseDTO, error) {
	severity := "LOW"
	riskScore := 5
	isViolationEvent := false
	maxViolations := 3

	switch req.EventType {
	case "TAB_SWITCH":
		severity = "MEDIUM"
		riskScore = 10
		isViolationEvent = true
	case "WINDOW_BLUR":
		severity = "LOW"
		riskScore = 5
		isViolationEvent = true
	case "FULLSCREEN_EXIT", "FULLSCREEN_REQUIRED":
		severity = "HIGH"
		riskScore = 10
		isViolationEvent = true
	case "COPY_ATTEMPT", "PASTE_ATTEMPT", "CUT_ATTEMPT":
		severity = "LOW"
		riskScore = 5
		isViolationEvent = true
	case "CONTEXT_MENU_ATTEMPT":
		severity = "LOW"
		riskScore = 2
	case "DEVTOOLS_ATTEMPT", "DEVTOOLS_OPEN":
		severity = "HIGH"
		riskScore = 15
		isViolationEvent = true
	case "MULTI_MONITOR", "VM_DETECTED":
		severity = "CRITICAL"
		riskScore = 25
		isViolationEvent = true
	case "PROCTOR_WARNING":
		severity = "WARNING"
		riskScore = 0
	case "PROCTOR_LOCK":
		severity = "CRITICAL"
		riskScore = 0
	case "PROCTOR_UNLOCK":
		severity = "INFO"
		riskScore = 0
	default:
		severity = "INFO"
		riskScore = 2
	}

	isCoolingDown := false
	if isViolationEvent {
		var recentCount int
		cooldownQ := `
			SELECT COUNT(*)
			FROM cat.ep_security_events
			WHERE id_ep_schedule = $1 AND kodepeserta = $2
			  AND created_at >= (NOW() - INTERVAL '5 seconds')
			  AND event_type IN ('TAB_SWITCH', 'WINDOW_BLUR', 'FULLSCREEN_EXIT', 'FULLSCREEN_REQUIRED', 'DEVTOOLS_ATTEMPT')
		`
		_ = r.db.QueryRowContext(ctx, cooldownQ, scheduleID, participantCode).Scan(&recentCount)
		if recentCount > 0 {
			isCoolingDown = true
		}
	}

	metaBytes, _ := json.Marshal(req.Metadata)
	metaStr := string(metaBytes)
	if isCoolingDown {
		if metaStr == "" || metaStr == "{}" {
			metaStr = `{"cooldown_suppressed":true}`
		} else {
			metaStr = strings.TrimSuffix(metaStr, "}") + `,"cooldown_suppressed":true}`
		}
	}

	insertQ := `
		INSERT INTO cat.ep_security_events (
			id_ep_schedule, kodepeserta, event_type, severity, risk_score, metadata, ip_address, user_agent, device_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, NOW())
	`
	_, _ = r.db.ExecContext(ctx, insertQ, scheduleID, participantCode, req.EventType, severity, riskScore, metaStr, ip, userAgent, req.DeviceID)

	isTabSwitch := 0
	isFsExit := 0
	effectiveRisk := riskScore

	if isCoolingDown {
		effectiveRisk = 0
	} else {
		if isViolationEvent {
			isTabSwitch = 1
		}
		if req.EventType == "FULLSCREEN_EXIT" || req.EventType == "FULLSCREEN_REQUIRED" {
			isFsExit = 1
		}
	}

	updateQ := `
		UPDATE cat.ep_schedule_participants
		SET risk_score = COALESCE(risk_score, 0) + $1,
		    tab_switch_count = COALESCE(tab_switch_count, 0) + $2,
		    fullscreen_exit_count = COALESCE(fullscreen_exit_count, 0) + $3,
		    total_violations = COALESCE(total_violations, 0) + $2,
		    risk_level = CASE
		        WHEN (COALESCE(risk_score, 0) + $1) >= 81 THEN 'CRITICAL'
		        WHEN (COALESCE(risk_score, 0) + $1) >= 61 THEN 'HIGH_RISK'
		        WHEN (COALESCE(risk_score, 0) + $1) >= 41 THEN 'SUSPICIOUS'
		        WHEN (COALESCE(risk_score, 0) + $1) >= 21 THEN 'ATTENTION'
		        ELSE 'NORMAL'
		    END,
		    is_locked = CASE
		        WHEN is_locked = TRUE THEN TRUE
		        WHEN (COALESCE(tab_switch_count, 0) + $2) >= $4 THEN TRUE
		        WHEN (COALESCE(risk_score, 0) + $1) >= 81 THEN TRUE
		        ELSE is_locked
		    END,
		    lock_reason = CASE
		        WHEN is_locked = TRUE THEN lock_reason
		        WHEN (COALESCE(tab_switch_count, 0) + $2) >= $4 THEN 'Sesi dikunci otomatis: jumlah pelanggaran mencapai batas maksimal (' || $4::text || 'x). Hubungi pengawas untuk membuka kunci.'
		        WHEN (COALESCE(risk_score, 0) + $1) >= 81 THEN 'Sesi dikunci otomatis: skor risiko mencapai level CRITICAL. Hubungi pengawas untuk membuka kunci.'
		        ELSE lock_reason
		    END,
		    last_seen_at = NOW(),
		    updated_at = NOW()
		WHERE id_ep_schedule = $5 AND kodepeserta = $6
		RETURNING risk_score, risk_level, tab_switch_count, fullscreen_exit_count, total_violations, is_locked
	`

	var prevLocked bool
	_ = r.db.QueryRowContext(ctx, `SELECT COALESCE(is_locked, FALSE) FROM cat.ep_schedule_participants WHERE id_ep_schedule = $1 AND kodepeserta = $2`, scheduleID, participantCode).Scan(&prevLocked)

	var res dto.EPSecurityEventResponseDTO
	var isLockedBool bool
	err := r.db.QueryRowContext(ctx, updateQ, effectiveRisk, isTabSwitch, isFsExit, maxViolations, scheduleID, participantCode).
		Scan(&res.CurrentRiskScore, &res.RiskLevel, &res.TabSwitchCount, &res.FullscreenExitCount, &res.TotalViolations, &isLockedBool)
	if err != nil {
		return nil, err
	}

	if isLockedBool {
		res.IsLocked = 1
	} else {
		res.IsLocked = 0
	}
	res.AutoLocked = !prevLocked && isLockedBool
	res.MaxViolations = maxViolations
	return &res, nil
}

func (r *epRepository) SyncParticipantHeartbeat(ctx context.Context, scheduleID int, participantCode string, req *dto.EPHeartbeatRequestDTO) (*dto.EPHeartbeatResponseDTO, error) {
	locWIB, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now()

	_, _ = r.db.ExecContext(ctx, `UPDATE cat.ep_schedule_participants SET last_seen_at = NOW() WHERE id_ep_schedule = $1 AND kodepeserta = $2`, scheduleID, participantCode)

	var row struct {
		IsLocked          bool       `db:"is_locked"`
		LockReason        *string    `db:"lock_reason"`
		ProctorWarning    *string    `db:"proctor_warning"`
		CurrentSection    int        `db:"current_section"`
		SessionStatus     string     `db:"session_status"`
		SectionStartedAt  *time.Time `db:"section_started_at"`
		SectionDeadlineAt *time.Time `db:"section_deadline_at"`
		DurationMinutes   int        `db:"duration_minutes"`
	}

	query := `
		SELECT sp.is_locked, sp.lock_reason, sp.proctor_warning, sp.current_section, sp.session_status,
		       sp.section_started_at, sp.section_deadline_at,
		       COALESCE(es.duration_minutes, 35) as duration_minutes
		FROM cat.ep_schedule_participants sp
		JOIN cat.ep_schedules s ON s.id_ep_schedule = sp.id_ep_schedule
		LEFT JOIN cat.ep_exam_sections es ON es.id_ep_exam = s.id_ep_exam AND es.section_order = sp.current_section
		WHERE sp.id_ep_schedule = $1 AND sp.kodepeserta = $2
	`
	err := r.db.GetContext(ctx, &row, query, scheduleID, participantCode)
	if err != nil {
		return nil, fmt.Errorf("peserta tidak ditemukan")
	}

	remainingSec := 0
	if row.SectionDeadlineAt != nil {
		rem := int(row.SectionDeadlineAt.Sub(now).Seconds())
		if rem > 0 {
			remainingSec = rem
		}
	} else if row.SectionStartedAt != nil {
		deadline := row.SectionStartedAt.Add(time.Duration(row.DurationMinutes) * time.Minute)
		rem := int(deadline.Sub(now).Seconds())
		if rem > 0 {
			remainingSec = rem
		}
	}

	isLockedInt := 0
	if row.IsLocked {
		isLockedInt = 1
	}

	return &dto.EPHeartbeatResponseDTO{
		ServerTime:       now.In(locWIB).Format("2006-01-02 15:04:05"),
		IsLocked:         isLockedInt,
		LockReason:       row.LockReason,
		ProctorWarning:   row.ProctorWarning,
		RemainingSeconds: remainingSec,
		SessionStatus:    row.SessionStatus,
		CurrentSection:   row.CurrentSection,
	}, nil
}

func (r *epRepository) ApplyProctorAction(ctx context.Context, req *dto.EPProctorActionRequestDTO) error {
	switch req.Action {
	case "SEND_WARNING":
		query := `UPDATE cat.ep_schedule_participants SET proctor_warning = $1, updated_at = NOW() WHERE id_ep_schedule = $2 AND kodepeserta = $3`
		_, err := r.db.ExecContext(ctx, query, req.WarningMessage, req.ScheduleID, req.ParticipantCode)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{"action":"SEND_WARNING","warning":%q,"note":"Pengawas mengirimkan pesan peringatan kepada peserta"}`, req.WarningMessage)
		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO cat.ep_security_events (id_ep_schedule, kodepeserta, event_type, severity, risk_score, metadata, created_at)
			VALUES ($1, $2, 'PROCTOR_WARNING', 'WARNING', 5, $3::jsonb, NOW())
		`, req.ScheduleID, req.ParticipantCode, meta)
		return nil

	case "LOCK_SESSION":
		reason := req.Reason
		if reason == "" {
			reason = "Sesi ujian dikunci manual oleh pengawas ruang"
		}
		query := `UPDATE cat.ep_schedule_participants SET is_locked = TRUE, lock_reason = $1, updated_at = NOW() WHERE id_ep_schedule = $2 AND kodepeserta = $3`
		_, err := r.db.ExecContext(ctx, query, reason, req.ScheduleID, req.ParticipantCode)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{"action":"LOCK_SESSION","reason":%q,"note":"Sesi ujian dikunci manual oleh pengawas ruang"}`, reason)
		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO cat.ep_security_events (id_ep_schedule, kodepeserta, event_type, severity, risk_score, metadata, created_at)
			VALUES ($1, $2, 'PROCTOR_LOCK', 'CRITICAL', 30, $3::jsonb, NOW())
		`, req.ScheduleID, req.ParticipantCode, meta)
		return nil

	case "UNLOCK_SESSION":
		reason := req.Reason
		if reason == "" {
			reason = "Kunci ujian dibuka oleh pengawas ruang (peserta diizinkan melanjutkan ujian)"
		}
		// PRESERVE VIOLATIONS:
		// Reset is_locked to FALSE, reset lock_reason and proctor_warning, reset active tab_switch_count to 0 so they don't immediately re-lock on next tick,
		// BUT PRESERVE total_violations and risk_score, and increment unlock_count!
		query := `
			UPDATE cat.ep_schedule_participants
			SET is_locked = FALSE,
			    lock_reason = NULL,
			    proctor_warning = NULL,
			    tab_switch_count = 0,
			    unlock_count = COALESCE(unlock_count, 0) + 1,
			    updated_at = NOW()
			WHERE id_ep_schedule = $1 AND kodepeserta = $2
		`
		_, err := r.db.ExecContext(ctx, query, req.ScheduleID, req.ParticipantCode)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{"action":"UNLOCK_SESSION","reason":%q,"note":"Kunci ujian dibuka oleh pengawas ruang (riwayat pelanggaran total & rekaman audit tetap tersimpan)"}`, reason)
		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO cat.ep_security_events (id_ep_schedule, kodepeserta, event_type, severity, risk_score, metadata, created_at)
			VALUES ($1, $2, 'PROCTOR_UNLOCK', 'INFO', 0, $3::jsonb, NOW())
		`, req.ScheduleID, req.ParticipantCode, meta)
		return nil

	case "FORCE_SUBMIT":
		query := `
			UPDATE cat.ep_schedule_participants
			SET session_status = 'FINISHED', finished_at = NOW(), is_locked = TRUE, lock_reason = 'Ujian dihentikan paksa oleh pengawas ruang', updated_at = NOW()
			WHERE id_ep_schedule = $1 AND kodepeserta = $2
		`
		_, err := r.db.ExecContext(ctx, query, req.ScheduleID, req.ParticipantCode)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{"action":"FORCE_SUBMIT","reason":%q,"note":"Ujian diselesaikan paksa oleh pengawas ruang"}`, req.Reason)
		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO cat.ep_security_events (id_ep_schedule, kodepeserta, event_type, severity, risk_score, metadata, created_at)
			VALUES ($1, $2, 'PROCTOR_FORCE_SUBMIT', 'CRITICAL', 50, $3::jsonb, NOW())
		`, req.ScheduleID, req.ParticipantCode, meta)
		return nil

	default:
		return fmt.Errorf("aksi pengawas tidak dikenal: %s", req.Action)
	}
}

func (r *epRepository) GetSecurityEventsByParticipant(ctx context.Context, scheduleID int, participantCode string) ([]*entity.EPSecurityEvent, error) {
	query := `
		SELECT se.id, se.id_ep_schedule, se.kodepeserta, se.event_type, se.severity, se.risk_score,
		       se.metadata::text, se.ip_address, se.user_agent, se.device_id, se.created_at,
		       COALESCE(p.nama, 'Peserta ' || se.kodepeserta) as nama_peserta
		FROM cat.ep_security_events se
		LEFT JOIN cat.at_peserta p ON p.kodepeserta = se.kodepeserta
		WHERE se.id_ep_schedule = $1
		  AND ($2 = '' OR se.kodepeserta = $2)
		ORDER BY se.id DESC
		LIMIT 100
	`
	var events []*entity.EPSecurityEvent
	err := r.db.SelectContext(ctx, &events, query, scheduleID, participantCode)
	return events, err
}

func (r *epRepository) SaveProctorSnapshot(ctx context.Context, scheduleID int, participantCode string, req *dto.UploadProctorSnapshotRequestDTO) (*dto.ProctorSnapshotItemDTO, error) {
	snapType := req.SnapshotType
	if snapType == "" {
		snapType = "PERIODIC"
	}
	trigEvent := req.TriggerEvent
	if trigEvent == "" {
		trigEvent = "INTERVAL_TIMER"
	}

	var insertedID int64
	var createdAt time.Time
	query := `
		INSERT INTO cat.ep_proctor_snapshots (
			id_ep_schedule, kodepeserta, snapshot_type, trigger_event, image_data, risk_score, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query, scheduleID, participantCode, snapType, trigEvent, req.ImageData, req.RiskScore).Scan(&insertedID, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan snapshot proctor: %w", err)
	}

	return &dto.ProctorSnapshotItemDTO{
		ID:           insertedID,
		IDEPSchedule: scheduleID,
		KodePeserta:  participantCode,
		SnapshotType: snapType,
		TriggerEvent: trigEvent,
		ImageData:    req.ImageData,
		RiskScore:    req.RiskScore,
		CreatedAt:    createdAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (r *epRepository) GetProctorSnapshots(ctx context.Context, scheduleID int, participantCode, snapshotType string, limit, offset int) ([]dto.ProctorSnapshotItemDTO, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT s.id, s.id_ep_schedule, s.kodepeserta,
		       COALESCE(pu.nama, ap.nama, s.kodepeserta) as nama_peserta,
		       s.snapshot_type, s.trigger_event, s.image_data, s.risk_score, s.created_at
		FROM cat.ep_proctor_snapshots s
		LEFT JOIN cat.ep_peserta_umum pu ON pu.kodepeserta = s.kodepeserta
		LEFT JOIN cat.at_peserta ap ON ap.kodepeserta = s.kodepeserta
		WHERE s.id_ep_schedule = $1
		  AND ($2 = '' OR s.kodepeserta = $2)
		  AND ($3 = '' OR $3 = 'ALL' OR s.snapshot_type = $3)
		ORDER BY s.id DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.QueryContext(ctx, query, scheduleID, participantCode, snapshotType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil snapshot proctor: %w", err)
	}
	defer rows.Close()

	var result []dto.ProctorSnapshotItemDTO
	for rows.Next() {
		var item dto.ProctorSnapshotItemDTO
		var t time.Time
		if err := rows.Scan(
			&item.ID, &item.IDEPSchedule, &item.KodePeserta,
			&item.NamaPeserta, &item.SnapshotType, &item.TriggerEvent,
			&item.ImageData, &item.RiskScore, &t,
		); err != nil {
			continue
		}
		item.CreatedAt = t.Format("2006-01-02 15:04:05")
		result = append(result, item)
	}

	return result, nil
}

func (r *epRepository) GetLatestProctorGrid(ctx context.Context, scheduleID int) ([]dto.ProctorLatestGridItemDTO, error) {
	query := `
		SELECT 
			sp.kodepeserta,
			COALESCE(pu.nama, ap.nama, sp.kodepeserta) as nama_peserta,
			snap.image_data as latest_snapshot,
			snap.snapshot_type,
			snap.trigger_event,
			snap.created_at as last_captured_at,
			COALESCE(sp.session_status, 'NOT_STARTED') as session_status,
			CASE WHEN sp.is_locked THEN 1 ELSE 0 END as is_locked,
			COALESCE(v.total_violations, 0) as total_violations
		FROM cat.ep_schedule_participants sp
		LEFT JOIN cat.ep_peserta_umum pu ON pu.kodepeserta = sp.kodepeserta
		LEFT JOIN cat.at_peserta ap ON ap.kodepeserta = sp.kodepeserta
		LEFT JOIN LATERAL (
			SELECT s.image_data, s.snapshot_type, s.trigger_event, s.created_at
			FROM cat.ep_proctor_snapshots s
			WHERE s.id_ep_schedule = sp.id_ep_schedule AND s.kodepeserta = sp.kodepeserta
			ORDER BY s.id DESC
			LIMIT 1
		) snap ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as total_violations
			FROM cat.ep_security_events se
			WHERE se.id_ep_schedule = sp.id_ep_schedule AND se.kodepeserta = sp.kodepeserta
			  AND se.event_type IN ('TAB_SWITCH', 'WINDOW_BLUR', 'FULLSCREEN_EXIT', 'DEVTOOLS_ATTEMPT', 'CAMERA_BLACKOUT')
		) v ON true
		WHERE sp.id_ep_schedule = $1
		ORDER BY sp.kodepeserta ASC
	`

	rows, err := r.db.QueryContext(ctx, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data grid pengawas: %w", err)
	}
	defer rows.Close()

	var result []dto.ProctorLatestGridItemDTO
	for rows.Next() {
		var item dto.ProctorLatestGridItemDTO
		var t sql.NullTime
		var snapData, sType, tEvent sql.NullString
		if err := rows.Scan(
			&item.KodePeserta,
			&item.NamaPeserta,
			&snapData,
			&sType,
			&tEvent,
			&t,
			&item.SessionStatus,
			&item.IsLocked,
			&item.TotalViolations,
		); err != nil {
			continue
		}

		if snapData.Valid {
			item.LatestSnapshot = &snapData.String
		}
		if sType.Valid {
			item.SnapshotType = &sType.String
		}
		if tEvent.Valid {
			item.TriggerEvent = &tEvent.String
		}
		if t.Valid {
			formatted := t.Time.Format("2006-01-02 15:04:05")
			item.LastCapturedAt = &formatted
		}

		if item.TotalViolations >= 3 || item.IsLocked == 1 {
			item.RiskLevel = "HIGH"
		} else if item.TotalViolations >= 1 {
			item.RiskLevel = "MEDIUM"
		} else {
			item.RiskLevel = "NORMAL"
		}

		result = append(result, item)
	}

	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 6. Scoring Engine (Raw Score -> Scaled Score Conversion & CEFR)
// ─────────────────────────────────────────────────────────────────────────────
func (r *epRepository) CalculateAndSaveScores(ctx context.Context, scheduleID int) ([]*entity.EPScore, error) {
	// 1. Get Exam Details & Profile
	var exam struct {
		IDEPExam     int  `db:"id_ep_exam"`
		IDExamType   int  `db:"id_exam_type"`
		IDProfile    *int `db:"id_profile"`
		PassingScore int  `db:"passing_score"`
	}
	err := r.db.GetContext(ctx, &exam, `
		SELECT e.id_ep_exam, e.id_exam_type, e.id_profile, COALESCE(e.passing_score, 450) as passing_score
		FROM cat.ep_schedules s
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		WHERE s.id_ep_schedule = $1
	`, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("ujian tidak ditemukan")
	}

	// Fallback to default conversion profile if none selected
	profileID := 1
	if exam.IDProfile != nil && *exam.IDProfile > 0 {
		profileID = *exam.IDProfile
	} else {
		_ = r.db.GetContext(ctx, &profileID, `SELECT id_profile FROM cat.ep_conversion_profiles WHERE id_exam_type = $1 AND is_default = TRUE LIMIT 1`, exam.IDExamType)
	}

	// 2. Load Score Conversion Matrix into Memory
	type convRow struct {
		SectionCode string `db:"section_code"`
		RawScore    int    `db:"raw_score"`
		ScaledScore int    `db:"scaled_score"`
	}
	var convRows []convRow
	_ = r.db.SelectContext(ctx, &convRows, `SELECT section_code, raw_score, scaled_score FROM cat.ep_score_conversion WHERE id_profile = $1`, profileID)
	convMap := make(map[string]map[int]int)
	for _, c := range convRows {
		if convMap[c.SectionCode] == nil {
			convMap[c.SectionCode] = make(map[int]int)
		}
		convMap[c.SectionCode][c.RawScore] = c.ScaledScore
	}

	// 3. Load CEFR Mappings
	var cefrRows []*entity.CEFRMapping
	_ = r.db.SelectContext(ctx, &cefrRows, `SELECT * FROM cat.ep_cefr_mapping WHERE id_exam_type = $1 ORDER BY score_min ASC`, exam.IDExamType)

	// 4. Fetch all participants in schedule
	var participants []*entity.ScheduleParticipant
	_ = r.db.SelectContext(ctx, &participants, `SELECT * FROM cat.ep_schedule_participants WHERE id_ep_schedule = $1 AND softdelete = '0'`, scheduleID)

	var calculatedScores []*entity.EPScore

	for _, p := range participants {
		// Calculate raw correct count per section
		// Section 1: Listening
		var rawL int
		_ = r.db.GetContext(ctx, &rawL, `
			SELECT COUNT(*)::integer
			FROM cat.cat_participant_answers_multi a
			JOIN cat.cat_stimulus_items i ON i.id_item = a.id_item
			JOIN cat.cat_bank_stimulus s ON s.id_stimulus = a.id_stimulus
			JOIN cat.ep_exam_sections es ON es.kodesoal = s.kodesoal AND es.id_ep_exam = $1 AND es.section_order = 1
			WHERE a.id_session = $2 AND UPPER(TRIM(a.selected_options)) = UPPER(TRIM(i.correct_answer))
		`, exam.IDEPExam, p.IDParticipant)

		// Section 2: Structure
		var rawS int
		_ = r.db.GetContext(ctx, &rawS, `
			SELECT COUNT(*)::integer
			FROM cat.cat_participant_answers_multi a
			JOIN cat.cat_stimulus_items i ON i.id_item = a.id_item
			JOIN cat.cat_bank_stimulus s ON s.id_stimulus = a.id_stimulus
			JOIN cat.ep_exam_sections es ON es.kodesoal = s.kodesoal AND es.id_ep_exam = $1 AND es.section_order = 2
			WHERE a.id_session = $2 AND UPPER(TRIM(a.selected_options)) = UPPER(TRIM(i.correct_answer))
		`, exam.IDEPExam, p.IDParticipant)

		// Section 3: Reading
		var rawR int
		_ = r.db.GetContext(ctx, &rawR, `
			SELECT COUNT(*)::integer
			FROM cat.cat_participant_answers_multi a
			JOIN cat.cat_stimulus_items i ON i.id_item = a.id_item
			JOIN cat.cat_bank_stimulus s ON s.id_stimulus = a.id_stimulus
			JOIN cat.ep_exam_sections es ON es.kodesoal = s.kodesoal AND es.id_ep_exam = $1 AND es.section_order = 3
			WHERE a.id_session = $2 AND UPPER(TRIM(a.selected_options)) = UPPER(TRIM(i.correct_answer))
		`, exam.IDEPExam, p.IDParticipant)

		// Lookup Scaled Scores
		scaledL := 24
		if m, ok := convMap["LISTENING"]; ok {
			if v, found := m[rawL]; found {
				scaledL = v
			}
		}

		scaledS := 20
		if m, ok := convMap["STRUCTURE"]; ok {
			if v, found := m[rawS]; found {
				scaledS = v
			}
		}

		scaledR := 20
		if m, ok := convMap["READING"]; ok {
			if v, found := m[rawR]; found {
				scaledR = v
			}
		}

		// Calculate Total Score formula: (L + S + R) / 3 * 10
		totalScore := int(float64(scaledL+scaledS+scaledR) / 3.0 * 10.0)

		// Determine CEFR level
		var cefrLevel *string
		for _, cf := range cefrRows {
			if totalScore >= cf.ScoreMin && totalScore <= cf.ScoreMax {
				lvl := cf.CEFRLevel
				cefrLevel = &lvl
				break
			}
		}

		passStatus := "FAIL"
		if totalScore >= exam.PassingScore {
			passStatus = "PASS"
		}

		scoreEntity := &entity.EPScore{
			IDParticipant:    p.IDParticipant,
			ListeningRaw:     rawL,
			ListeningScaled:  scaledL,
			StructureRaw:     rawS,
			StructureScaled:  scaledS,
			ReadingRaw:       rawR,
			ReadingScaled:    scaledR,
			TotalScaledScore: totalScore,
			CEFRLevel:        cefrLevel,
			PassingStatus:    passStatus,
			CalculatedAt:     time.Now(),
		}

		// Save score record
		upsertScoreQ := `
			INSERT INTO cat.ep_scores (id_participant, listening_raw, listening_scaled, structure_raw, structure_scaled, reading_raw, reading_scaled, total_scaled_score, cefr_level, passing_status, calculated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
			ON CONFLICT (id_participant)
			DO UPDATE SET
				listening_raw = EXCLUDED.listening_raw,
				listening_scaled = EXCLUDED.listening_scaled,
				structure_raw = EXCLUDED.structure_raw,
				structure_scaled = EXCLUDED.structure_scaled,
				reading_raw = EXCLUDED.reading_raw,
				reading_scaled = EXCLUDED.reading_scaled,
				total_scaled_score = EXCLUDED.total_scaled_score,
				cefr_level = EXCLUDED.cefr_level,
				passing_status = EXCLUDED.passing_status,
				calculated_at = NOW()
			RETURNING id_score
		`
		_ = r.db.GetContext(ctx, &scoreEntity.IDScore, upsertScoreQ,
			scoreEntity.IDParticipant, scoreEntity.ListeningRaw, scoreEntity.ListeningScaled,
			scoreEntity.StructureRaw, scoreEntity.StructureScaled, scoreEntity.ReadingRaw,
			scoreEntity.ReadingScaled, scoreEntity.TotalScaledScore, scoreEntity.CEFRLevel, scoreEntity.PassingStatus)

		calculatedScores = append(calculatedScores, scoreEntity)
	}

	return calculatedScores, nil
}

func (r *epRepository) GetScoresBySchedule(ctx context.Context, scheduleID int) ([]*entity.EPScore, error) {
	query := `
		SELECT sc.*, COALESCE(pu.nama, p.nama, 'Peserta ' || sp.kodepeserta) as nama, sp.kodepeserta, e.exam_name,
		       COALESCE(r.namaruang, 'Ruang ' || s.id_ruang::text) as room_name,
		       TO_CHAR(s.exam_date, 'YYYY-MM-DD') as exam_date,
		       TO_CHAR(s.start_time, 'HH24:MI') as start_time
		FROM cat.ep_scores sc
		JOIN cat.ep_schedule_participants sp ON sp.id_participant = sc.id_participant
		JOIN cat.ep_schedules s ON s.id_ep_schedule = sp.id_ep_schedule
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		LEFT JOIN cat.at_ruang r ON (r.koderuang = s.id_ruang::text OR r.koderuang = LPAD(s.id_ruang::text, 3, '0'))
		LEFT JOIN cat.ep_peserta_umum pu ON pu.kodepeserta = sp.kodepeserta
		LEFT JOIN cat.at_peserta p ON p.kodepeserta = sp.kodepeserta
		WHERE sp.id_ep_schedule = $1
		ORDER BY sc.total_scaled_score DESC, sc.listening_scaled DESC
	`
	var scores []*entity.EPScore
	err := r.db.SelectContext(ctx, &scores, query, scheduleID)
	return scores, err
}

func (r *epRepository) GetScoresByExam(ctx context.Context, examID int) ([]*entity.EPScore, error) {
	query := `
		SELECT sc.*, COALESCE(pu.nama, p.nama, 'Peserta ' || sp.kodepeserta) as nama, sp.kodepeserta, e.exam_name,
		       COALESCE(r.namaruang, 'Ruang ' || s.id_ruang::text) as room_name,
		       TO_CHAR(s.exam_date, 'YYYY-MM-DD') as exam_date,
		       TO_CHAR(s.start_time, 'HH24:MI') as start_time
		FROM cat.ep_scores sc
		JOIN cat.ep_schedule_participants sp ON sp.id_participant = sc.id_participant
		JOIN cat.ep_schedules s ON s.id_ep_schedule = sp.id_ep_schedule
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		LEFT JOIN cat.at_ruang r ON (r.koderuang = s.id_ruang::text OR r.koderuang = LPAD(s.id_ruang::text, 3, '0'))
		LEFT JOIN cat.ep_peserta_umum pu ON pu.kodepeserta = sp.kodepeserta
		LEFT JOIN cat.at_peserta p ON p.kodepeserta = sp.kodepeserta
		WHERE e.id_ep_exam = $1
		ORDER BY s.exam_date ASC, s.start_time ASC, sc.total_scaled_score DESC, sc.listening_scaled DESC
	`
	var scores []*entity.EPScore
	err := r.db.SelectContext(ctx, &scores, query, examID)
	return scores, err
}

// ─────────────────────────────────────────────────────────────────────────────
// 7. Certificates & Verification Engine
// ─────────────────────────────────────────────────────────────────────────────
func (r *epRepository) GenerateCertificates(ctx context.Context, scheduleID int) ([]*entity.EPCertificate, error) {
	// First calculate all scores
	_, err := r.CalculateAndSaveScores(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	// Fetch exam info & validity
	var schedInfo struct {
		ExamDate               time.Time `db:"exam_date"`
		ExamTypeName           string    `db:"exam_type_name"`
		ValidityMonths         int       `db:"validity_months"`
		ValidityMonthsOverride *int      `db:"validity_months_override"`
	}
	err = r.db.GetContext(ctx, &schedInfo, `
		SELECT s.exam_date, t.type_name as exam_type_name, t.validity_months, e.validity_months_override
		FROM cat.ep_schedules s
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		JOIN cat.ep_exam_types t ON t.id_exam_type = e.id_exam_type
		WHERE s.id_ep_schedule = $1
	`, scheduleID)
	if err != nil {
		return nil, err
	}

	validityMonths := schedInfo.ValidityMonths
	if schedInfo.ValidityMonthsOverride != nil && *schedInfo.ValidityMonthsOverride > 0 {
		validityMonths = *schedInfo.ValidityMonthsOverride
	}

	scores, err := r.GetScoresBySchedule(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	var generated []*entity.EPCertificate
	year := schedInfo.ExamDate.Year()

	for _, sc := range scores {
		verifyCode := uuid.New().String()
		certNum := fmt.Sprintf("TOEFL/PKK/%d/%05d", year, sc.IDScore)
		expiryDate := schedInfo.ExamDate.AddDate(0, validityMonths, 0)

		cert := &entity.EPCertificate{
			IDScore:           sc.IDScore,
			CertificateNumber: certNum,
			ParticipantName:   sc.Nama,
			ParticipantCode:   sc.KodePeserta,
			ExamTypeName:      schedInfo.ExamTypeName,
			ExamDate:          schedInfo.ExamDate,
			ExamLocation:      "Lab CBT Poltekkes Kemenkes Semarang",
			ListeningScore:    sc.ListeningScaled,
			StructureScore:    sc.StructureScaled,
			ReadingScore:      sc.ReadingScaled,
			TotalScore:        sc.TotalScaledScore,
			CEFRLevel:         sc.CEFRLevel,
			IssuedDate:        time.Now(),
			ExpiryDate:        expiryDate,
			VerificationCode:  verifyCode,
		}

		insertQ := `
			INSERT INTO cat.ep_certificates (
				id_score, certificate_number, participant_name, participant_code, exam_type_name,
				exam_date, exam_location, listening_score, structure_score, reading_score,
				total_score, cefr_level, issued_date, expiry_date, verification_code, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_DATE, $13, $14, NOW(), NOW()
			)
			ON CONFLICT (certificate_number)
			DO UPDATE SET
				listening_score = EXCLUDED.listening_score,
				structure_score = EXCLUDED.structure_score,
				reading_score = EXCLUDED.reading_score,
				total_score = EXCLUDED.total_score,
				cefr_level = EXCLUDED.cefr_level,
				updated_at = NOW()
			RETURNING *
		`
		_ = r.db.GetContext(ctx, cert, insertQ,
			cert.IDScore, cert.CertificateNumber, cert.ParticipantName, cert.ParticipantCode,
			cert.ExamTypeName, cert.ExamDate, cert.ExamLocation, cert.ListeningScore,
			cert.StructureScore, cert.ReadingScore, cert.TotalScore, cert.CEFRLevel,
			cert.ExpiryDate, cert.VerificationCode)

		generated = append(generated, cert)
	}

	return generated, nil
}

func (r *epRepository) GetCertificateByNumber(ctx context.Context, certNumber string) (*entity.EPCertificate, error) {
	var cert entity.EPCertificate
	err := r.db.GetContext(ctx, &cert, `SELECT * FROM cat.ep_certificates WHERE certificate_number = $1`, certNumber)
	return &cert, err
}

func (r *epRepository) VerifyCertificate(ctx context.Context, verificationCode string) (*entity.EPCertificate, error) {
	var cert entity.EPCertificate
	err := r.db.GetContext(ctx, &cert, `SELECT * FROM cat.ep_certificates WHERE verification_code = $1`, verificationCode)
	return &cert, err
}

func (r *epRepository) GetParticipantScoresAndCertificates(ctx context.Context, participantCode string) ([]dto.ParticipantScoreResultDTO, error) {
	// 1. Check any finished schedules for this participant that need score calculation or certificate generation
	var finishedSchedules []struct {
		IDEPSchedule  int `db:"id_ep_schedule"`
		IDParticipant int `db:"id_participant"`
	}
	err := r.db.SelectContext(ctx, &finishedSchedules, `
		SELECT sp.id_ep_schedule, sp.id_participant
		FROM cat.ep_schedule_participants sp
		WHERE sp.kodepeserta = $1 AND sp.is_finished = true
	`, participantCode)
	if err == nil {
		for _, fs := range finishedSchedules {
			var scoreCount int
			_ = r.db.GetContext(ctx, &scoreCount, `SELECT COUNT(*) FROM cat.ep_scores WHERE id_participant = $1`, fs.IDParticipant)
			if scoreCount == 0 {
				_, _ = r.CalculateAndSaveScores(ctx, fs.IDEPSchedule)
			}
			var certCount int
			_ = r.db.GetContext(ctx, &certCount, `SELECT COUNT(*) FROM cat.ep_certificates WHERE participant_code = $1`, participantCode)
			if certCount == 0 {
				_, _ = r.GenerateCertificates(ctx, fs.IDEPSchedule)
			}
		}
	}

	// 2. Query final scores & certificates for this participant
	query := `
		SELECT 
			sc.id_score,
			s.id_ep_schedule as schedule_id,
			e.exam_name,
			COALESCE(r.namaruang, 'Ruang ' || s.id_ruang::text) as room_name,
			TO_CHAR(s.exam_date, 'YYYY-MM-DD') as exam_date,
			TO_CHAR(s.start_time, 'HH24:MI') as start_time,
			sp.kodepeserta as participant_code,
			COALESCE(pu.nama, p.nama, 'Peserta ' || sp.kodepeserta) as participant_name,
			sc.listening_raw,
			sc.listening_scaled,
			sc.structure_raw,
			sc.structure_scaled,
			sc.reading_raw,
			sc.reading_scaled,
			sc.total_scaled_score,
			sc.cefr_level,
			sc.passing_status,
			TO_CHAR(sc.calculated_at, 'YYYY-MM-DD HH24:MI:SS') as calculated_at,
			(c.id_certificate IS NOT NULL) as has_certificate,
			c.certificate_number,
			c.verification_code,
			TO_CHAR(c.issued_date, 'YYYY-MM-DD') as issued_date,
			TO_CHAR(c.expiry_date, 'YYYY-MM-DD') as expiry_date,
			COALESCE(c.exam_location, 'Lab CBT UPT Bahasa Poltekkes Kemenkes Surabaya') as exam_location
		FROM cat.ep_scores sc
		JOIN cat.ep_schedule_participants sp ON sp.id_participant = sc.id_participant
		JOIN cat.ep_schedules s ON s.id_ep_schedule = sp.id_ep_schedule
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		LEFT JOIN cat.at_ruang r ON (r.koderuang = s.id_ruang::text OR r.koderuang = LPAD(s.id_ruang::text, 3, '0'))
		LEFT JOIN cat.ep_peserta_umum pu ON pu.kodepeserta = sp.kodepeserta
		LEFT JOIN cat.at_peserta p ON p.kodepeserta = sp.kodepeserta
		LEFT JOIN cat.ep_certificates c ON c.id_score = sc.id_score
		WHERE sp.kodepeserta = $1
		ORDER BY s.exam_date DESC, s.start_time DESC
	`
	var results []dto.ParticipantScoreResultDTO
	err = r.db.SelectContext(ctx, &results, query, participantCode)
	if err != nil {
		return nil, err
	}
	if results == nil {
		results = []dto.ParticipantScoreResultDTO{}
	}
	return results, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 8. Dynamic Question Banks
// ─────────────────────────────────────────────────────────────────────────────
func (r *epRepository) GetStimuliByBank(ctx context.Context, kodesoal string) ([]dto.DynamicStimulusDTO, error) {
	type rawStimulus struct {
		IDStimulus      int64           `db:"id_stimulus"`
		StimulusCode    string          `db:"stimulus_code"`
		Title           string          `db:"title"`
		NarrativeText   string          `db:"narrative_text"`
		MediaType       string          `db:"media_type"`
		MediaURL        *string         `db:"media_url"`
		MediaMetadata   json.RawMessage `db:"media_metadata"`
		CategoryTag     string          `db:"category_tag"`
		DifficultyLevel int             `db:"difficulty_level"`
	}
	var rawStimuli []rawStimulus
	stimQuery := `
		SELECT id_stimulus, stimulus_code, title, narrative_text, media_type, media_url, media_metadata, category_tag, difficulty_level
		FROM cat.cat_bank_stimulus
		WHERE kodesoal = $1 AND (softdelete = '0' OR softdelete IS NULL)
		ORDER BY sort_order ASC, id_stimulus ASC
	`
	err := r.db.SelectContext(ctx, &rawStimuli, stimQuery, kodesoal)
	if err != nil {
		return nil, err
	}

	var results []dto.DynamicStimulusDTO
	for idx, s := range rawStimuli {
		type rawItem struct {
			IDItem            int64   `db:"id_item"`
			ItemOrder         int     `db:"item_order"`
			QuestionText      string  `db:"question_text"`
			QuestionMediaType string  `db:"question_media_type"`
			QuestionMediaURL  *string `db:"question_media_url"`
			CorrectAnswer     string  `db:"correct_answer"`
			Explanation       *string `db:"explanation"`
		}
		var items []rawItem
		itemQ := `
			SELECT id_item, item_order, question_text, question_media_type, question_media_url, correct_answer, explanation
			FROM cat.cat_stimulus_items
			WHERE id_stimulus = $1 AND (softdelete = '0' OR softdelete IS NULL)
			ORDER BY item_order ASC, id_item ASC
		`
		_ = r.db.SelectContext(ctx, &items, itemQ, s.IDStimulus)

		var itemDTOs []dto.DynamicItemDTO
		for _, it := range items {
			type rawOpt struct {
				IDOption    int64   `db:"id_option"`
				OptionLabel string  `db:"option_label"`
				OptionText  string  `db:"option_text"`
				MediaType   string  `db:"media_type"`
				MediaURL    *string `db:"media_url"`
			}
			var opts []rawOpt
			optQ := `
				SELECT id_option, option_label, option_text, media_type, media_url
				FROM cat.cat_item_options
				WHERE id_item = $1
				ORDER BY sort_order ASC, option_label ASC
			`
			_ = r.db.SelectContext(ctx, &opts, optQ, it.IDItem)

			var optDTOs []dto.DynamicOptionDTO
			for _, o := range opts {
				optDTOs = append(optDTOs, dto.DynamicOptionDTO{
					IDOption:    o.IDOption,
					OptionLabel: o.OptionLabel,
					OptionText:  o.OptionText,
					MediaType:   o.MediaType,
					MediaURL:    o.MediaURL,
				})
			}

			itemDTOs = append(itemDTOs, dto.DynamicItemDTO{
				IDItem:            it.IDItem,
				ItemOrder:         it.ItemOrder,
				QuestionText:      it.QuestionText,
				QuestionMediaType: it.QuestionMediaType,
				QuestionMediaURL:  it.QuestionMediaURL,
				CorrectAnswer:     it.CorrectAnswer,
				Explanation:       it.Explanation,
				Options:           optDTOs,
			})
		}

		var metaMap map[string]any
		_ = json.Unmarshal(s.MediaMetadata, &metaMap)

		results = append(results, dto.DynamicStimulusDTO{
			IDStimulus:     s.IDStimulus,
			StimulusNumber: idx + 1,
			StimulusCode:   s.StimulusCode,
			Title:          s.Title,
			NarrativeText:  s.NarrativeText,
			MediaType:      s.MediaType,
			MediaURL:       s.MediaURL,
			MediaMetadata:  metaMap,
			CategoryTag:    s.CategoryTag,
			SubItems:       itemDTOs,
		})
	}

	return results, nil
}

func (r *epRepository) CreateStimulusWithItems(ctx context.Context, req *dto.CreateStimulusRequestDTO) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if strings.TrimSpace(req.StimulusCode) == "" {
		var count int
		_ = tx.GetContext(ctx, &count, "SELECT COUNT(*) FROM cat.cat_bank_stimulus WHERE kodesoal = $1", req.KodeSoal)
		req.StimulusCode = fmt.Sprintf("%s_STIM_%02d", req.KodeSoal, count+1)
	}
	if strings.TrimSpace(req.Title) == "" {
		req.Title = fmt.Sprintf("Naskah Stimulus #%s", strings.TrimPrefix(req.StimulusCode, req.KodeSoal+"_STIM_"))
	}

	var stimulusID int64
	insertStimQ := `
		INSERT INTO cat.cat_bank_stimulus (
			kodesoal, stimulus_code, title, narrative_text, media_type, media_url, category_tag, difficulty_level, sort_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, 1, NOW(), NOW()
		) RETURNING id_stimulus
	`
	err = tx.GetContext(ctx, &stimulusID, insertStimQ,
		req.KodeSoal, req.StimulusCode, req.Title, req.NarrativeText,
		req.MediaType, req.MediaURL, req.CategoryTag, req.DifficultyLevel)
	if err != nil {
		return fmt.Errorf("gagal menyimpan stimulus: %w", err)
	}

	insertItemQ := `
		INSERT INTO cat.cat_stimulus_items (
			id_stimulus, item_order, question_text, question_media_type, question_media_url, item_type, weight_correct, correct_answer, explanation, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 'SINGLE_CHOICE', 1.0, $6, $7, NOW(), NOW()
		) RETURNING id_item
	`

	insertOptQ := `
		INSERT INTO cat.cat_item_options (
			id_item, option_label, option_text, media_type, media_url, sort_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
	`

	for idx, it := range req.Items {
		var itemID int64
		order := it.ItemOrder
		if order == 0 {
			order = idx + 1
		}
		qMediaType := it.QuestionMediaType
		if qMediaType == "" {
			qMediaType = "NONE"
		}
		err = tx.GetContext(ctx, &itemID, insertItemQ, stimulusID, order, it.QuestionText, qMediaType, it.QuestionMediaURL, it.CorrectAnswer, it.Explanation)
		if err != nil {
			return fmt.Errorf("gagal menyimpan butir soal: %w", err)
		}

		for oIdx, op := range it.Options {
			mediaType := op.MediaType
			if mediaType == "" {
				mediaType = "NONE"
			}
			_, err = tx.ExecContext(ctx, insertOptQ, itemID, op.OptionLabel, op.OptionText, mediaType, op.MediaURL, oIdx+1)
			if err != nil {
				return fmt.Errorf("gagal menyimpan pilihan jawaban: %w", err)
			}
		}
	}

	return tx.Commit()
}

func (r *epRepository) GetEPQuestionBanks(ctx context.Context, examTypeID, sectionID int, search string, page, perPage int) ([]*entity.EPQuestionBank, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	baseQuery := `
		FROM cat.ep_question_banks b
		LEFT JOIN cat.ep_exam_types t ON t.id_exam_type = b.id_exam_type
		LEFT JOIN cat.ep_sections s ON s.id_section = b.id_section
		WHERE (b.softdelete = '0' OR b.softdelete IS NULL)
	`
	var args []any
	argCount := 1

	if search != "" {
		baseQuery += fmt.Sprintf(" AND (b.kodesoal ILIKE $%d OR b.namasoal ILIKE $%d OR b.description ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			b.id_ep_bank, b.kodesoal, b.namasoal, 
			COALESCE(b.id_exam_type, 0) as id_exam_type, COALESCE(t.type_name, 'TOEFL / TOEIC') as exam_type_name,
			COALESCE(b.id_section, 0) as id_section, COALESCE(s.section_name, 'Umum') as section_name, COALESCE(s.section_code, 'GENERAL') as section_code, 
			b.description, b.created_at, b.updated_at,
			(SELECT COUNT(*) FROM cat.cat_bank_stimulus stim WHERE stim.kodesoal = b.kodesoal AND (stim.softdelete = '0' OR stim.softdelete IS NULL)) as total_stimuli,
			(SELECT COUNT(*) FROM cat.cat_stimulus_items item JOIN cat.cat_bank_stimulus stim ON stim.id_stimulus = item.id_stimulus WHERE stim.kodesoal = b.kodesoal AND (stim.softdelete = '0' OR stim.softdelete IS NULL)) as total_questions
		%s
		ORDER BY b.id_ep_bank DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argCount, argCount+1)

	args = append(args, perPage, offset)

	var list []*entity.EPQuestionBank
	err = r.db.SelectContext(ctx, &list, selectQuery, args...)
	return list, total, err
}

func (r *epRepository) GetEPQuestionBankByCode(ctx context.Context, kodesoal string) (*entity.EPQuestionBank, error) {
	var bank entity.EPQuestionBank
	query := `
		SELECT 
			b.id_ep_bank, b.kodesoal, b.namasoal, 
			COALESCE(b.id_exam_type, 0) as id_exam_type, COALESCE(t.type_name, 'TOEFL / TOEIC') as exam_type_name,
			COALESCE(b.id_section, 0) as id_section, COALESCE(s.section_name, 'Umum') as section_name, COALESCE(s.section_code, 'GENERAL') as section_code, 
			b.description, b.created_at, b.updated_at,
			(SELECT COUNT(*) FROM cat.cat_bank_stimulus stim WHERE stim.kodesoal = b.kodesoal AND (stim.softdelete = '0' OR stim.softdelete IS NULL)) as total_stimuli,
			(SELECT COUNT(*) FROM cat.cat_stimulus_items item JOIN cat.cat_bank_stimulus stim ON stim.id_stimulus = item.id_stimulus WHERE stim.kodesoal = b.kodesoal AND (stim.softdelete = '0' OR stim.softdelete IS NULL)) as total_questions
		FROM cat.ep_question_banks b
		LEFT JOIN cat.ep_exam_types t ON t.id_exam_type = b.id_exam_type
		LEFT JOIN cat.ep_sections s ON s.id_section = b.id_section
		WHERE b.kodesoal = $1 AND (b.softdelete = '0' OR b.softdelete IS NULL)
	`
	err := r.db.GetContext(ctx, &bank, query, kodesoal)
	return &bank, err
}

func (r *epRepository) CreateEPQuestionBank(ctx context.Context, req *dto.CreateEPBankDTO) (*entity.EPQuestionBank, error) {
	if strings.TrimSpace(req.KodeSoal) == "" {
		var count int
		_ = r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM cat.ep_question_banks")
		for i := count + 1; ; i++ {
			candidate := fmt.Sprintf("TOEFL_BANK_%02d", i)
			var exists bool
			_ = r.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM cat.ep_question_banks WHERE kodesoal = $1)", candidate)
			if !exists {
				req.KodeSoal = candidate
				break
			}
		}
	}

	// 1. Sync to at_soal
	_, _ = r.db.ExecContext(ctx, `
		INSERT INTO cat.at_soal (kodesoal, namasoal, keterangan)
		VALUES ($1, $2, $3)
		ON CONFLICT (kodesoal) DO NOTHING
	`, req.KodeSoal, req.NamaSoal, req.Description)

	// 2. Insert ep_question_banks
	insertQ := `
		INSERT INTO cat.ep_question_banks (
			kodesoal, namasoal, description, created_at, updated_at
		) VALUES (
			$1, $2, $3, NOW(), NOW()
		) RETURNING id_ep_bank
	`
	var id int64
	err := r.db.GetContext(ctx, &id, insertQ, req.KodeSoal, req.NamaSoal, req.Description)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat bank soal: %w", err)
	}

	return r.GetEPQuestionBankByCode(ctx, req.KodeSoal)
}

func (r *epRepository) UpdateEPQuestionBank(ctx context.Context, kodesoal string, req *dto.UpdateEPBankDTO) error {
	_, _ = r.db.ExecContext(ctx, `
		UPDATE cat.at_soal SET namasoal = $1, keterangan = $2, t_updatetime = NOW()
		WHERE kodesoal = $3
	`, req.NamaSoal, req.Description, kodesoal)

	updateQ := `
		UPDATE cat.ep_question_banks
		SET namasoal = $1, description = $2, updated_at = NOW()
		WHERE kodesoal = $3
	`
	_, err := r.db.ExecContext(ctx, updateQ, req.NamaSoal, req.Description, kodesoal)
	return err
}

func (r *epRepository) DeleteEPQuestionBank(ctx context.Context, kodesoal string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cat.ep_question_banks SET softdelete = '1', updated_at = NOW() WHERE kodesoal = $1`, kodesoal)
	return err
}

func (r *epRepository) CopyEPQuestionBank(ctx context.Context, req *dto.CopyEPBankDTO, sourceKode string) (*entity.EPQuestionBank, error) {
	src, err := r.GetEPQuestionBankByCode(ctx, sourceKode)
	if err != nil {
		return nil, fmt.Errorf("bank soal sumber tidak ditemukan: %w", err)
	}

	createReq := &dto.CreateEPBankDTO{
		KodeSoal:    req.NewKodeSoal,
		NamaSoal:    req.NewNamaSoal,
		Description: src.Description,
	}

	newBank, err := r.CreateEPQuestionBank(ctx, createReq)
	if err != nil {
		return nil, err
	}

	// Copy stimuli & questions
	stimuli, err := r.GetStimuliByBank(ctx, sourceKode)
	if err == nil {
		for sIdx, st := range stimuli {
			var subItems []dto.CreateStimulusSubItemDTO
			for _, it := range st.SubItems {
				var opts []dto.CreateStimulusItemOptionDTO
				for _, o := range it.Options {
					opts = append(opts, dto.CreateStimulusItemOptionDTO{
						OptionLabel: o.OptionLabel,
						OptionText:  o.OptionText,
						MediaType:   o.MediaType,
						MediaURL:    o.MediaURL,
					})
				}

				var correctAns string
				_ = r.db.GetContext(ctx, &correctAns, `SELECT correct_answer FROM cat.cat_stimulus_items WHERE id_item = $1`, it.IDItem)
				if correctAns == "" {
					correctAns = "A"
				}

				subItems = append(subItems, dto.CreateStimulusSubItemDTO{
					ItemOrder:     it.ItemOrder,
					QuestionText:  it.QuestionText,
					CorrectAnswer: correctAns,
					Options:       opts,
				})
			}

			_ = r.CreateStimulusWithItems(ctx, &dto.CreateStimulusRequestDTO{
				KodeSoal:        req.NewKodeSoal,
				StimulusCode:    fmt.Sprintf("%s_S%d", req.NewKodeSoal, sIdx+1),
				Title:           st.Title,
				NarrativeText:   st.NarrativeText,
				MediaType:       st.MediaType,
				MediaURL:        st.MediaURL,
				CategoryTag:     st.CategoryTag,
				DifficultyLevel: 2,
				Items:           subItems,
			})
		}
	}

	return newBank, nil
}

func (r *epRepository) UpdateStimulus(ctx context.Context, idStimulus int64, req *dto.UpdateStimulusRequestDTO) error {
	query := `
		UPDATE cat.cat_bank_stimulus
		SET title = $1, narrative_text = $2, media_type = $3, media_url = $4, category_tag = $5, difficulty_level = $6, updated_at = NOW()
		WHERE id_stimulus = $7
	`
	_, err := r.db.ExecContext(ctx, query, req.Title, req.NarrativeText, req.MediaType, req.MediaURL, req.CategoryTag, req.DifficultyLevel, idStimulus)
	return err
}

func (r *epRepository) DeleteStimulus(ctx context.Context, idStimulus int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cat.cat_bank_stimulus SET softdelete = '1', updated_at = NOW() WHERE id_stimulus = $1`, idStimulus)
	return err
}

func (r *epRepository) CreateSubItem(ctx context.Context, stimulusID int64, req *dto.CreateSubItemRequestDTO) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var maxOrder int
	_ = tx.GetContext(ctx, &maxOrder, `SELECT COALESCE(MAX(item_order), 0) FROM cat.cat_stimulus_items WHERE id_stimulus = $1 AND (softdelete = '0' OR softdelete IS NULL)`, stimulusID)

	order := req.ItemOrder
	if order <= 0 {
		order = maxOrder + 1
	}

	insertItemQ := `
		INSERT INTO cat.cat_stimulus_items (
			id_stimulus, item_order, question_text, question_media_type, question_media_url, item_type, weight_correct, correct_answer, explanation, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 'SINGLE_CHOICE', 1.0, $6, $7, NOW(), NOW()
		) RETURNING id_item
	`
	var itemID int64
	qMediaType := req.QuestionMediaType
	if qMediaType == "" {
		qMediaType = "NONE"
	}
	err = tx.GetContext(ctx, &itemID, insertItemQ, stimulusID, order, req.QuestionText, qMediaType, req.QuestionMediaURL, req.CorrectAnswer, req.Explanation)
	if err != nil {
		return fmt.Errorf("gagal menambahkan butir soal: %w", err)
	}

	insertOptQ := `
		INSERT INTO cat.cat_item_options (
			id_item, option_label, option_text, media_type, media_url, sort_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
	`
	for oIdx, op := range req.Options {
		mediaType := op.MediaType
		if mediaType == "" {
			mediaType = "NONE"
		}
		_, err = tx.ExecContext(ctx, insertOptQ, itemID, op.OptionLabel, op.OptionText, mediaType, op.MediaURL, oIdx+1)
		if err != nil {
			return fmt.Errorf("gagal menyimpan opsi jawaban: %w", err)
		}
	}

	return tx.Commit()
}

func (r *epRepository) UpdateSubItem(ctx context.Context, itemID int64, req *dto.UpdateSubItemRequestDTO) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qMediaType := req.QuestionMediaType
	if qMediaType == "" {
		qMediaType = "NONE"
	}

	updateItemQ := `
		UPDATE cat.cat_stimulus_items
		SET question_text = $1, question_media_type = $2, question_media_url = $3, correct_answer = $4, explanation = $5, updated_at = NOW()
		WHERE id_item = $6
	`
	_, err = tx.ExecContext(ctx, updateItemQ, req.QuestionText, qMediaType, req.QuestionMediaURL, req.CorrectAnswer, req.Explanation, itemID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui butir soal: %w", err)
	}

	// Delete and re-insert options
	_, _ = tx.ExecContext(ctx, `DELETE FROM cat.cat_item_options WHERE id_item = $1`, itemID)

	insertOptQ := `
		INSERT INTO cat.cat_item_options (
			id_item, option_label, option_text, media_type, media_url, sort_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
	`
	for oIdx, op := range req.Options {
		mediaType := op.MediaType
		if mediaType == "" {
			mediaType = "NONE"
		}
		_, err = tx.ExecContext(ctx, insertOptQ, itemID, op.OptionLabel, op.OptionText, mediaType, op.MediaURL, oIdx+1)
		if err != nil {
			return fmt.Errorf("gagal menyimpan opsi jawaban: %w", err)
		}
	}

	return tx.Commit()
}

func (r *epRepository) DeleteSubItem(ctx context.Context, itemID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cat.cat_stimulus_items SET softdelete = '1', updated_at = NOW() WHERE id_item = $1`, itemID)
	return err
}

func (r *epRepository) ImportBankStimuli(ctx context.Context, kodesoal string, stimuli []dto.CreateStimulusRequestDTO) (int, int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	stimulusCount := 0
	questionCount := 0

	var currentCount int
	_ = tx.GetContext(ctx, &currentCount, "SELECT COUNT(*) FROM cat.cat_bank_stimulus WHERE kodesoal = $1", kodesoal)

	insertStimQ := `
		INSERT INTO cat.cat_bank_stimulus (
			kodesoal, stimulus_code, title, narrative_text, media_type, media_url, category_tag, difficulty_level, sort_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		) RETURNING id_stimulus
	`

	insertItemQ := `
		INSERT INTO cat.cat_stimulus_items (
			id_stimulus, item_order, question_text, question_media_type, question_media_url, item_type, weight_correct, correct_answer, explanation, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 'SINGLE_CHOICE', 1.0, $6, $7, NOW(), NOW()
		) RETURNING id_item
	`

	insertOptQ := `
		INSERT INTO cat.cat_item_options (
			id_item, option_label, option_text, media_type, media_url, sort_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
	`

	for sIdx, stim := range stimuli {
		currentCount++
		code := stim.StimulusCode
		if strings.TrimSpace(code) == "" {
			code = fmt.Sprintf("%s_STIM_%02d", kodesoal, currentCount)
		}
		title := stim.Title
		if strings.TrimSpace(title) == "" {
			title = fmt.Sprintf("Stimulus #%d", currentCount)
		}
		mediaType := stim.MediaType
		if mediaType == "" {
			mediaType = "NONE"
		}
		catTag := stim.CategoryTag
		if catTag == "" {
			catTag = "Umum"
		}
		diff := stim.DifficultyLevel
		if diff == 0 {
			diff = 2
		}

		var stimulusID int64
		err = tx.GetContext(ctx, &stimulusID, insertStimQ,
			kodesoal, code, title, stim.NarrativeText,
			mediaType, stim.MediaURL, catTag, diff, sIdx+1)
		if err != nil {
			return 0, 0, fmt.Errorf("gagal insert stimulus %s: %w", title, err)
		}
		stimulusCount++

		for itIdx, item := range stim.Items {
			order := item.ItemOrder
			if order == 0 {
				order = itIdx + 1
			}
			qMediaType := item.QuestionMediaType
			if qMediaType == "" {
				qMediaType = "NONE"
			}
			correct := strings.ToUpper(strings.TrimSpace(item.CorrectAnswer))
			if correct == "" {
				correct = "A"
			}

			var itemID int64
			err = tx.GetContext(ctx, &itemID, insertItemQ,
				stimulusID, order, item.QuestionText, qMediaType, item.QuestionMediaURL, correct, item.Explanation)
			if err != nil {
				return 0, 0, fmt.Errorf("gagal insert butir soal #%d: %w", order, err)
			}
			questionCount++

			for optIdx, opt := range item.Options {
				optMediaType := opt.MediaType
				if optMediaType == "" {
					optMediaType = "NONE"
				}
				label := opt.OptionLabel
				if label == "" {
					label = string(rune('A' + optIdx))
				}
				_, err = tx.ExecContext(ctx, insertOptQ,
					itemID, label, opt.OptionText, optMediaType, opt.MediaURL, optIdx+1)
				if err != nil {
					return 0, 0, fmt.Errorf("gagal insert opsi %s: %w", label, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return stimulusCount, questionCount, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 12. External Participants (Peserta Umum TOEFL)
// ─────────────────────────────────────────────────────────────────────────────

func (r *epRepository) GetExternalParticipants(ctx context.Context, search string, limit, offset int) ([]*entity.EPPesertaUmum, int, error) {
	countQ := `SELECT COUNT(*) FROM cat.ep_peserta_umum WHERE is_active = TRUE`
	dataQ := `SELECT * FROM cat.ep_peserta_umum WHERE is_active = TRUE`
	var args []interface{}

	if search != "" {
		filter := "%" + search + "%"
		countQ += ` AND (kodepeserta ILIKE $1 OR nama ILIKE $1 OR email ILIKE $1 OR COALESCE(instansi, '') ILIKE $1 OR COALESCE(nik, '') ILIKE $1)`
		dataQ += ` AND (kodepeserta ILIKE $1 OR nama ILIKE $1 OR email ILIKE $1 OR COALESCE(instansi, '') ILIKE $1 OR COALESCE(nik, '') ILIKE $1)`
		args = append(args, filter)
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, err
	}

	dataQ += ` ORDER BY id_peserta_umum DESC`
	if limit > 0 {
		if len(args) > 0 {
			dataQ += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		} else {
			dataQ += " LIMIT $1 OFFSET $2"
		}
		args = append(args, limit, offset)
	}

	var list []*entity.EPPesertaUmum
	if err := r.db.SelectContext(ctx, &list, dataQ, args...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *epRepository) GetExternalParticipantByCode(ctx context.Context, code string) (*entity.EPPesertaUmum, error) {
	var item entity.EPPesertaUmum
	query := `SELECT * FROM cat.ep_peserta_umum WHERE kodepeserta = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &item, query, code)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *epRepository) GetNextExternalParticipantCode(ctx context.Context) (string, error) {
	prefix := "T" + time.Now().Format("020106")
	query := `
		SELECT COALESCE(MAX(
			CASE 
				WHEN length(kodepeserta) >= 13 AND SUBSTRING(kodepeserta FROM 8) ~ '^[0-9]+$' 
				THEN SUBSTRING(kodepeserta FROM 8)::bigint 
				ELSE 0 
			END
		), 0) + 1
		FROM cat.ep_peserta_umum
		WHERE kodepeserta LIKE $1
	`
	var nextSeq int64
	err := r.db.GetContext(ctx, &nextSeq, query, prefix+"%")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%06d", prefix, nextSeq), nil
}

func (r *epRepository) CreateExternalParticipant(ctx context.Context, p *entity.EPPesertaUmum) error {
	if p.KodePeserta == "" {
		code, err := r.GetNextExternalParticipantCode(ctx)
		if err != nil {
			return fmt.Errorf("gagal men-generate kode peserta: %w", err)
		}
		p.KodePeserta = code
	}

	pwd := p.Password
	if pwd == "" {
		if p.TanggalLahir != nil && !p.TanggalLahir.IsZero() {
			pwd = p.TanggalLahir.Format("02012006")
		} else {
			pwd = p.KodePeserta
		}
		p.Password = pwd
	}

	query := `
		INSERT INTO cat.ep_peserta_umum (
			kodepeserta, nik, nama, jenis_kelamin, email, hp, instansi,
			tempat_lahir, tanggal_lahir, alamat, password, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, md5($11), TRUE, NOW(), NOW()
		)
		ON CONFLICT (kodepeserta) DO UPDATE SET
			nik = COALESCE(EXCLUDED.nik, cat.ep_peserta_umum.nik),
			nama = EXCLUDED.nama,
			jenis_kelamin = EXCLUDED.jenis_kelamin,
			email = EXCLUDED.email,
			hp = EXCLUDED.hp,
			instansi = EXCLUDED.instansi,
			tempat_lahir = EXCLUDED.tempat_lahir,
			tanggal_lahir = EXCLUDED.tanggal_lahir,
			alamat = EXCLUDED.alamat,
			is_active = TRUE,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		p.KodePeserta, p.NIK, p.Nama, p.JenisKelamin, p.Email, p.HP, p.Instansi,
		p.TempatLahir, p.TanggalLahir, p.Alamat, pwd,
	)
	return err
}

func (r *epRepository) UpdateExternalParticipant(ctx context.Context, p *entity.EPPesertaUmum) error {
	query := `
		UPDATE cat.ep_peserta_umum SET
			nik = $2,
			nama = $3,
			jenis_kelamin = $4,
			email = $5,
			hp = $6,
			instansi = $7,
			tempat_lahir = $8,
			tanggal_lahir = $9,
			alamat = $10,
			is_active = $11,
			updated_at = NOW()
	`
	args := []interface{}{
		p.KodePeserta, p.NIK, p.Nama, p.JenisKelamin, p.Email, p.HP, p.Instansi,
		p.TempatLahir, p.TanggalLahir, p.Alamat, p.IsActive,
	}
	if p.Password != "" {
		query += `, password = md5($12) WHERE kodepeserta = $1`
		args = append(args, p.Password)
	} else {
		query += ` WHERE kodepeserta = $1`
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *epRepository) DeleteExternalParticipant(ctx context.Context, code string) error {
	query := `UPDATE cat.ep_peserta_umum SET is_active = FALSE, updated_at = NOW() WHERE kodepeserta = $1`
	_, err := r.db.ExecContext(ctx, query, code)
	return err
}

func (r *epRepository) BatchImportExternalParticipants(ctx context.Context, list []*entity.EPPesertaUmum) (int, error) {
	inserted := 0
	for _, p := range list {
		if p.KodePeserta == "" || p.Nama == "" {
			continue
		}
		if p.Password == "" {
			p.Password = p.KodePeserta
		}
		if err := r.CreateExternalParticipant(ctx, p); err == nil {
			inserted++
		}
	}
	return inserted, nil
}
