package repository

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/examination/dto"
	"poltekkes-cat-backend/internal/modules/examination/entity"
	"poltekkes-cat-backend/internal/shared/utils"

	"github.com/jmoiron/sqlx"
)

type ExaminationRepository interface {
	GetExams(ctx context.Context, page, perPage, periodID int, search string) ([]*entity.Ujian, int64, error)
	GetExamByID(ctx context.Context, examID int) (*entity.Ujian, error)
	CreateExam(ctx context.Context, req *dto.CreateExamRequestDTO) (int, error)
	UpdateExam(ctx context.Context, examID int, req *dto.UpdateExamRequestDTO) error
	DeleteExam(ctx context.Context, examID int) error
	GetSchedules(ctx context.Context, examID, periodID int) ([]*entity.JadwalUjian, error)
	GetExamParticipants(ctx context.Context, examID int) ([]*entity.ExamParticipant, error)
	StartExam(ctx context.Context, participantCode string, scheduleID int, sessionToken string, deviceID string, deviceFingerprint string) (*entity.JadwalUjian, int, string, error)
	GetSessionQuestions(ctx context.Context, participantCode string, scheduleID int) (*dto.SessionQuestionsResponseDTO, error)
	SaveAnswer(ctx context.Context, participantCode string, scheduleID int, questionBankCode string, questionNumber int, selectedOption int, isDoubtful bool) error
	GetLiveMonitoring(ctx context.Context, scheduleID int) ([]*entity.LiveMonitorRow, error)
	FinishExam(ctx context.Context, participantCode string, scheduleID int, submitReason string) (*entity.JadwalUjian, int, int, int, int, float64, string, error)

	// Security & Anti-Cheat Methods
	RecordSecurityEvent(ctx context.Context, event *entity.SecurityEvent) (*dto.SecurityEventResponseDTO, error)
	GetRecentSecurityEvents(ctx context.Context, scheduleID int, limit int) ([]*entity.SecurityEvent, error)
	ApplyProctorAction(ctx context.Context, scheduleID int, participantCode, action, reason, warning string) error
	UpdateHeartbeatAndCheckStatus(ctx context.Context, scheduleID int, participantCode string, req *dto.HeartbeatRequestDTO) (*dto.HeartbeatResponseDTO, error)
	GetCollusionAnalysis(ctx context.Context, scheduleID int) (*dto.CollusionReportResponseDTO, error)

	// Dynamic Multi-Media & Multi-Item Methods
	GetDynamicExamSession(ctx context.Context, participantCode string, scheduleID int) (*dto.DynamicExamSessionResponseDTO, error)
	SaveDynamicAnswer(ctx context.Context, req *dto.SaveMultiAnswerDTO) error
	FinishDynamicExam(ctx context.Context, participantCode string, scheduleID int, submitReason string) (*dto.FinishExamResponseDTO, error)
	// === Admin CRUD: Jadwal Sesi ===
	GetSessionsByExam(ctx context.Context, examID int) ([]*entity.JadwalSesi, error)
	CreateSession(ctx context.Context, req *dto.CreateScheduleRequestDTO) (int, error)
	UpdateSession(ctx context.Context, sessionID int, req *dto.UpdateSessionRequestDTO) error
	DeleteSession(ctx context.Context, sessionID int) error
	RefreshToken(ctx context.Context, scheduleID int) (string, error)

	// === Admin CRUD: Ruang per Sesi ===
	GetSessionRooms(ctx context.Context, sessionID int) ([]*entity.RuangUjian, error)
	AddRoomToSession(ctx context.Context, req *dto.CreateRoomSessionRequestDTO) (int, error)
	UpdateSessionRoom(ctx context.Context, roomSessionID int, req *dto.UpdateRoomSessionRequestDTO) error
	DeleteSessionRoom(ctx context.Context, roomSessionID int) error

	// === Admin CRUD: Peserta Ujian ===
	AddExamParticipant(ctx context.Context, examID int, participantCode string) error
	RemoveExamParticipant(ctx context.Context, examID int, participantCode string) error
	GetAvailableParticipants(ctx context.Context, examID int) ([]*entity.AvailableParticipant, error)
	ImportSipenmaruToExam(ctx context.Context, examID int, req *dto.ExamImportSipenmaruRequestDTO) (*dto.ExamImportSipenmaruResponseDTO, error)

	// === Admin: Plotting / Distribusi Peserta & Ruang ===
	AutoDistributeParticipants(ctx context.Context, examID int, overwrite bool) (*dto.AutoDistributeResponseDTO, error)
	GetRoomParticipants(ctx context.Context, roomSessionID int) ([]*entity.RoomParticipantRow, error)
	AddRoomParticipants(ctx context.Context, roomSessionID int, participantCodes []string) error
	RemoveRoomParticipant(ctx context.Context, roomSessionID int, participantCode string) error

	// === Barcode Attendance Verification ===
	VerifyParticipantBarcode(ctx context.Context, scheduleID int, participantCode string, barcode string) (*dto.VerifyBarcodeResponseDTO, error)
	VerifyProctorBarcode(ctx context.Context, scheduleID int, barcode string, proctorUser string) (*dto.VerifyBarcodeResponseDTO, error)
	GetSessionAttendanceSummary(ctx context.Context, scheduleID int) (*dto.AttendanceSummaryResponseDTO, error)
	SetParticipantVerificationStatus(ctx context.Context, scheduleID int, participantCode string, isVerified int, verifier string) error
	GetParticipantSchedules(ctx context.Context, participantCode string) ([]*dto.ParticipantScheduleDTO, error)

	// === Admin CRUD: Bank Soal per Sesi ===
	GetSessionBankSoal(ctx context.Context, sessionID int) ([]*entity.SoalUjian, error)
	AddBankSoalToSession(ctx context.Context, sessionID int, req *dto.AddBankSoalToSessionDTO) error
	RemoveBankSoalFromSession(ctx context.Context, sessionID int, kodeSoal string) error

	// === Completed Participants & Detailed Review ===
	GetCompletedParticipants(ctx context.Context, examID int, scheduleID int) ([]*dto.CompletedParticipantDTO, error)
	GetParticipantCompletedExams(ctx context.Context, participantCode string) ([]*dto.ParticipantExamSummaryDTO, error)
	GetParticipantExamResultDetail(ctx context.Context, scheduleID int, participantCode string) (*dto.ParticipantExamResultDetailDTO, error)
}

type examinationRepository struct {
	db       *sqlx.DB // cbtDB
	siakadDB *sqlx.DB
}

func NewExaminationRepository(db *sqlx.DB, siakadDBs ...*sqlx.DB) ExaminationRepository {
	sDB := db
	if len(siakadDBs) > 0 && siakadDBs[0] != nil {
		sDB = siakadDBs[0]
	}
	return &examinationRepository{db: db, siakadDB: sDB}
}

func (r *examinationRepository) getSatkerNames(ctx context.Context) map[string]string {
	satkerMap := make(map[string]string)
	if r.siakadDB == nil {
		return satkerMap
	}
	type satkerItem struct {
		IDSatker   string `db:"idsatker"`
		NamaSatker string `db:"namasatker"`
	}
	var items []satkerItem
	err := r.siakadDB.SelectContext(ctx, &items, `
		SELECT idsatker, namasatker 
		FROM gate.sc_unit 
		WHERE softdelete = '0' OR softdelete IS NULL
	`)
	if err == nil {
		for _, item := range items {
			satkerMap[item.IDSatker] = item.NamaSatker
		}
	}
	return satkerMap
}

func (r *examinationRepository) GetExams(ctx context.Context, page, perPage, periodID int, search string) ([]*entity.Ujian, int64, error) {
	offset := (page - 1) * perPage

	whereClauses := []string{"(u.softdelete = '0' OR u.softdelete IS NULL)"}
	countWhere := []string{"(softdelete = '0' OR softdelete IS NULL)"}
	var args []interface{}
	argIdx := 1

	if periodID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("u.idperiode = $%d", argIdx))
		countWhere = append(countWhere, fmt.Sprintf("idperiode = $%d", argIdx))
		args = append(args, periodID)
		argIdx++
	}

	search = strings.TrimSpace(search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("u.namaujian ILIKE $%d", argIdx))
		countWhere = append(countWhere, fmt.Sprintf("namaujian ILIKE $%d", argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereSQL := " WHERE " + strings.Join(whereClauses, " AND ")
	countSQL := "SELECT COUNT(*) FROM cat.at_ujian WHERE " + strings.Join(countWhere, " AND ")

	var total int64
	err := r.db.GetContext(ctx, &total, countSQL, args...)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			u.idujian,
			u.namaujian,
			u.idperiode,
			COALESCE(p.namaperiode, '') as namaperiode,
			COALESCE(u.nilaiminimal, 0) as nilaiminimal,
			COALESCE(u.idsatker, '') as idsatker,
			'' as namasatker,
			COALESCE(u.kodejenis, '') as kodejenis,
			COALESCE(j.namajenis, '') as namajenis,
			COALESCE(u.ispercobaan, 0) as ispercobaan,
			u.keterangan,
			COALESCE(u.max_violations, 5) as max_violations
		FROM cat.at_ujian u
		LEFT JOIN cat.at_periode p ON p.idperiode = u.idperiode
		LEFT JOIN cat.at_jenisujian j ON j.kodejenis = u.kodejenis
	` + whereSQL + fmt.Sprintf(" ORDER BY u.idujian DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)

	args = append(args, perPage, offset)

	var exams []*entity.Ujian
	err = r.db.SelectContext(ctx, &exams, query, args...)
	if err == nil && len(exams) > 0 {
		satkerNames := r.getSatkerNames(ctx)
		for _, e := range exams {
			if name, ok := satkerNames[e.IDSatker]; ok {
				e.NamaSatker = name
			}
		}
	}
	return exams, total, err
}

func (r *examinationRepository) GetExamByID(ctx context.Context, examID int) (*entity.Ujian, error) {
	query := `
		SELECT 
			u.idujian,
			u.namaujian,
			u.idperiode,
			COALESCE(p.namaperiode, '') as namaperiode,
			COALESCE(u.nilaiminimal, 0) as nilaiminimal,
			COALESCE(u.idsatker, '') as idsatker,
			'' as namasatker,
			COALESCE(u.kodejenis, '') as kodejenis,
			COALESCE(j.namajenis, '') as namajenis,
			COALESCE(u.ispercobaan, 0) as ispercobaan,
			u.keterangan,
			COALESCE(u.max_violations, 5) as max_violations
		FROM cat.at_ujian u
		LEFT JOIN cat.at_periode p ON p.idperiode = u.idperiode
		LEFT JOIN cat.at_jenisujian j ON j.kodejenis = u.kodejenis
		WHERE u.idujian = $1 AND (u.softdelete = '0' OR u.softdelete IS NULL)
	`
	var exam entity.Ujian
	err := r.db.GetContext(ctx, &exam, query, examID)
	if err != nil {
		return nil, err
	}
	satkerNames := r.getSatkerNames(ctx)
	if name, ok := satkerNames[exam.IDSatker]; ok {
		exam.NamaSatker = name
	}
	return &exam, nil
}

func (r *examinationRepository) CreateExam(ctx context.Context, req *dto.CreateExamRequestDTO) (int, error) {
	maxV := req.MaxViolations
	if maxV <= 0 {
		maxV = 5
	}
	query := `
		INSERT INTO cat.at_ujian (
			namaujian, idperiode, nilaiminimal, idsatker, kodejenis, ispercobaan, keterangan, max_violations, softdelete
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '0')
		RETURNING idujian
	`
	keterangan := sql.NullString{String: req.Description, Valid: req.Description != ""}
	var id int
	err := r.db.QueryRowContext(ctx, query,
		req.ExamName,
		req.PeriodID,
		req.PassingGrade,
		req.IDSatker,
		req.KodeJenis,
		req.IsPercobaan,
		keterangan,
		maxV,
	).Scan(&id)
	return id, err
}

func (r *examinationRepository) UpdateExam(ctx context.Context, examID int, req *dto.UpdateExamRequestDTO) error {
	maxV := req.MaxViolations
	if maxV <= 0 {
		maxV = 5
	}
	query := `
		UPDATE cat.at_ujian
		SET namaujian = COALESCE(NULLIF($1, ''), namaujian),
		    idperiode = CASE WHEN $2 > 0 THEN $2 ELSE idperiode END,
		    nilaiminimal = $3,
		    idsatker = $4,
		    kodejenis = $5,
		    ispercobaan = $6,
		    keterangan = $7,
		    max_violations = $8,
		    t_updatetime = NOW()
		WHERE idujian = $9 AND (softdelete = '0' OR softdelete IS NULL)
	`
	keterangan := sql.NullString{String: req.Description, Valid: req.Description != ""}
	_, err := r.db.ExecContext(ctx, query,
		req.ExamName,
		req.PeriodID,
		req.PassingGrade,
		req.IDSatker,
		req.KodeJenis,
		req.IsPercobaan,
		keterangan,
		maxV,
		examID,
	)
	return err
}

func (r *examinationRepository) DeleteExam(ctx context.Context, examID int) error {
	query := `UPDATE cat.at_ujian SET softdelete = '1', t_updatetime = NOW() WHERE idujian = $1`
	_, err := r.db.ExecContext(ctx, query, examID)
	return err
}

func (r *examinationRepository) GetSchedules(ctx context.Context, examID, periodID int) ([]*entity.JadwalUjian, error) {
	query := `
		SELECT j.idjadwalujian, j.idujian, u.namaujian, 
		       COALESCE(u.idperiode, 0) as idperiode,
		       COALESCE(p.namaperiode, '') as namaperiode,
		       COALESCE((SELECT su.kodesoal FROM cat.at_soalujian su WHERE su.idjadwalujian = j.idjadwalujian LIMIT 1), 'SOAL_2026') as kodesoal, 
		       COALESCE(r.idruangujian, 1) as idruang, 
		       COALESCE(r.koderuang, 'Lab 1') as namaruang, 
		       COALESCE(r.tglmulai, NOW())::date as tglujian, 
		       COALESCE(r.waktumulai, '08:00') as jammulai, 
		       '09:30' as jamselesai, 
		       COALESCE(j.waktupengerjaan::integer, 90) as waktupengerjaan, 
		       COALESCE(j.token_ujian, '') as tokenujian, 
		       COALESCE(r.jumlahpeserta::integer, 40) as kapasitas,
		       COALESCE(j.max_violations, u.max_violations, 5) as max_violations
		FROM cat.at_jadwalujian j
		JOIN cat.at_ujian u ON u.idujian = j.idujian
		LEFT JOIN cat.at_periode p ON p.idperiode = u.idperiode
		LEFT JOIN cat.at_ruangujian r ON r.idjadwalujian = j.idjadwalujian
		WHERE (j.softdelete = '0' OR j.softdelete IS NULL)
	`
	if examID > 0 {
		query += fmt.Sprintf(" AND j.idujian = %d", examID)
	}
	if periodID > 0 {
		query += fmt.Sprintf(" AND u.idperiode = %d", periodID)
	}
	query += " ORDER BY j.idjadwalujian DESC"

	var schedules []*entity.JadwalUjian
	err := r.db.SelectContext(ctx, &schedules, query)
	return schedules, err
}

type StoredQuestionSnapshotItem struct {
	QuestionNumber    int                     `json:"question_number"`
	OriginalNoUrut    int                     `json:"original_nourut"`
	KodeSoal          string                  `json:"kodesoal"`
	QuestionText      string                  `json:"question_text"`
	QuestionMediaType string                  `json:"question_media_type,omitempty"`
	QuestionMediaURL  *string                 `json:"question_media_url,omitempty"`
	Options           []dto.QuestionOptionDTO `json:"options"`
	CorrectOptionKey  int                     `json:"correct_option_key"`
	CategoryTag       string                  `json:"category_tag"`
	OptionMapping     [5]int                  `json:"option_mapping"`
	QuestionType      string                  `json:"question_type,omitempty"` // STATIC, DYNAMIC_STIMULUS
	WeightCorrect     float64                 `json:"weight_correct"`
	WeightWrong       float64                 `json:"weight_wrong"`
	StimulusID        *int64                  `json:"stimulus_id,omitempty"`
	StimulusItemID    *int64                  `json:"stimulus_item_id,omitempty"`
	StimulusTitle     *string                 `json:"stimulus_title,omitempty"`
	StimulusText      *string                 `json:"stimulus_text,omitempty"`
	StimulusMediaType string                  `json:"stimulus_media_type,omitempty"`
	StimulusMediaURL  *string                 `json:"stimulus_media_url,omitempty"`
}

type StoredSessionSnapshot struct {
	ScheduleID       int                          `json:"schedule_id"`
	QuestionBankCode string                       `json:"question_bank_code"`
	TotalQuestions   int                          `json:"total_questions"`
	SnapshotVersion  int                          `json:"snapshot_version"`
	CapturedAt       string                       `json:"captured_at"`
	Questions        []StoredQuestionSnapshotItem `json:"questions"`
}

func (r *examinationRepository) StartExam(ctx context.Context, participantCode string, scheduleID int, sessionToken string, deviceID string, deviceFingerprint string) (*entity.JadwalUjian, int, string, error) {
	var sched entity.JadwalUjian
	query := `
		SELECT j.idjadwalujian, j.idujian, 
		       COALESCE((SELECT su.kodesoal FROM cat.at_soalujian su WHERE su.idjadwalujian = j.idjadwalujian LIMIT 1), 'SOAL_2026') as kodesoal, 
		       COALESCE(r.idruangujian, 1) as idruang, 
		       COALESCE(r.koderuang, 'Lab 1') as namaruang, 
		       COALESCE(r.tglmulai, NOW())::date as tglujian, 
		       COALESCE(r.waktumulai, '08:00') as jammulai, 
		       '09:30' as jamselesai, 
		       COALESCE(j.waktupengerjaan::integer, 90) as waktupengerjaan, 
		       COALESCE(j.token_ujian, '') as tokenujian, 
		       COALESCE(r.jumlahpeserta::integer, 40) as kapasitas,
		       COALESCE(j.max_violations, u.max_violations, 5) as max_violations
		FROM cat.at_jadwalujian j
		LEFT JOIN cat.at_ujian u ON u.idujian = j.idujian
		LEFT JOIN cat.at_ruangujian r ON r.idjadwalujian = j.idjadwalujian
		WHERE j.idjadwalujian = $1 AND (j.softdelete = '0' OR j.softdelete IS NULL)
	`
	err := r.db.GetContext(ctx, &sched, query, scheduleID)
	if err != nil {
		return nil, 0, "", fmt.Errorf("jadwal ujian tidak ditemukan")
	}

	now := time.Now()
	durationSec := sched.WaktuPengerjaan * 60

	// Check if previous session was finished or active
	var prevRow struct {
		TglMulai         *time.Time `db:"tglmulai"`
		TglSelesai       *time.Time `db:"tglselesai"`
		RemainingSeconds *int       `db:"remaining_seconds"`
	}
	_ = r.db.GetContext(ctx, &prevRow, `SELECT tglmulai, tglselesai, remaining_seconds FROM cat.at_jadwalpeserta WHERE idjadwalujian = $1 AND kodepeserta = $2`, scheduleID, participantCode)

	// BLOCK: If exam was already submitted (tglselesai IS NOT NULL), prevent re-taking
	if prevRow.TglSelesai != nil {
		return nil, 0, "", fmt.Errorf("ujian ini sudah pernah dikerjakan dan dikumpulkan. Anda tidak dapat mengulang ujian yang sama")
	}

	// ALWAYS validate session token (even when resuming) for security
	if sched.TokenUjian != "" {
		trimmedToken := strings.ToUpper(strings.TrimSpace(sessionToken))
		expectedToken := strings.ToUpper(strings.TrimSpace(sched.TokenUjian))
		if trimmedToken == "" || trimmedToken != expectedToken {
			return nil, 0, "", fmt.Errorf("token ujian salah atau tidak valid. Silakan tanyakan token aktif kepada pengawas ruangan")
		}
	}

	// Verification Check: Participant must have scanned barcode or been verified by proctor
	var verifyRow struct {
		IsVerified int `db:"is_verified"`
	}
	_ = r.db.GetContext(ctx, &verifyRow, `SELECT COALESCE(is_verified, 0) as is_verified FROM cat.at_jadwalpeserta WHERE idjadwalujian = $1 AND kodepeserta = $2`, scheduleID, participantCode)
	if verifyRow.IsVerified == 0 {
		return nil, 0, "", fmt.Errorf("peserta belum melakukan scan barcode kehadiran. Silakan scan kartu ujian atau verifikasi ke pengawas lab terlebih dahulu")
	}

	// Resume: keep existing timer. Fresh start: reset timer.
	shouldResetTimer := prevRow.TglMulai == nil

	updQuery := `
		UPDATE cat.at_jadwalpeserta
		SET tglmulai = CASE WHEN $6 = true THEN $1 ELSE COALESCE(tglmulai, $1) END,
		    remaining_seconds = CASE 
		        WHEN $6 = true THEN $7 
		        WHEN remaining_seconds IS NOT NULL AND remaining_seconds > 0 THEN LEAST(remaining_seconds, $7)
		        WHEN tglmulai IS NOT NULL THEN GREATEST(0, $7 - EXTRACT(EPOCH FROM ($1 - tglmulai))::integer)
		        ELSE $7
		    END,
		    is_locked = 0,
		    lock_reason = NULL,
		    last_heartbeat = $1,
		    client_device_id = COALESCE($4, client_device_id),
		    device_fingerprint = COALESCE($5, device_fingerprint),
		    reconnect_count = COALESCE(reconnect_count, 0) + CASE WHEN tglmulai IS NOT NULL AND $6 = false THEN 1 ELSE 0 END
		WHERE idjadwalujian = $2 AND kodepeserta = $3
	`
	_, _ = r.db.ExecContext(ctx, updQuery, now, scheduleID, participantCode, deviceID, deviceFingerprint, shouldResetTimer, durationSec)

	// Fetch exact tglmulai and remaining_seconds
	var sessionRow struct {
		TglMulai         *time.Time `db:"tglmulai"`
		RemainingSeconds *int       `db:"remaining_seconds"`
		IsLocked         int        `db:"is_locked"`
	}
	_ = r.db.GetContext(ctx, &sessionRow, `SELECT tglmulai, remaining_seconds, COALESCE(is_locked, 0) as is_locked FROM cat.at_jadwalpeserta WHERE idjadwalujian = $1 AND kodepeserta = $2`, scheduleID, participantCode)

	remainingSeconds := durationSec
	startedAtStr := now.Format(time.RFC3339)
	if sessionRow.RemainingSeconds != nil && *sessionRow.RemainingSeconds > 0 {
		remainingSeconds = *sessionRow.RemainingSeconds
	} else if sessionRow.TglMulai != nil {
		elapsed := int(now.Sub(*sessionRow.TglMulai).Seconds())
		remainingSeconds = durationSec - elapsed
		if remainingSeconds < 0 {
			remainingSeconds = 0
		}
	}
	// Safety cap: remaining_seconds must never exceed the exam's total duration
	if remainingSeconds > durationSec {
		remainingSeconds = durationSec
	}
	if sessionRow.TglMulai != nil {
		startedAtStr = sessionRow.TglMulai.Format(time.RFC3339)
	}

	// Ensure immutable session snapshot is built / loaded
	_, _ = r.buildAndSaveSessionSnapshot(ctx, scheduleID, participantCode, sched.KodeSoal)

	return &sched, remainingSeconds, startedAtStr, nil
}

func (r *examinationRepository) buildAndSaveSessionSnapshot(ctx context.Context, scheduleID int, participantCode, kodesoal string) (*dto.SessionQuestionsResponseDTO, error) {
	// Fetch configuration from cat.at_soalujian
	type soalUjianConfig struct {
		KodeSoal           string `db:"kodesoal"`
		JumlahPertanyaan   int    `db:"jumlahpertanyaan"`
		JumlahSoalStatis   int    `db:"jumlah_soal_statis"`
		JumlahKasusDinamis int    `db:"jumlah_kasus_dinamis"`
		NoUrut             int    `db:"nourut"`
	}
	var configs []soalUjianConfig
	_ = r.db.SelectContext(ctx, &configs, `
		SELECT kodesoal, 
		       COALESCE(jumlahpertanyaan, 0)::integer as jumlahpertanyaan, 
		       COALESCE(jumlah_soal_statis, 0)::integer as jumlah_soal_statis, 
		       COALESCE(jumlah_kasus_dinamis, 0)::integer as jumlah_kasus_dinamis, 
		       COALESCE(nourut, 1)::integer as nourut
		FROM cat.at_soalujian
		WHERE idjadwalujian = $1 AND (softdelete = '0' OR softdelete IS NULL)
		ORDER BY nourut ASC
	`, scheduleID)

	expectedTotal := 0
	hasDynamicQuotaConfigured := false
	for _, cfg := range configs {
		if cfg.JumlahKasusDinamis > 0 {
			hasDynamicQuotaConfigured = true
		}
		expectedTotal += cfg.JumlahPertanyaan
	}

	// Fetch participant state & check existing snapshot
	var participantState struct {
		TglMulai         *time.Time     `db:"tglmulai"`
		RemainingSeconds *int           `db:"remaining_seconds"`
		WaktuPengerjaan  int            `db:"waktupengerjaan"`
		SessionSnapshot  sql.NullString `db:"session_snapshot"`
		MaxViolations    int            `db:"max_violations"`
	}
	_ = r.db.GetContext(ctx, &participantState, `
		SELECT jp.tglmulai, jp.remaining_seconds, COALESCE(j.waktupengerjaan::integer, 90) as waktupengerjaan, jp.session_snapshot::text as session_snapshot,
		       COALESCE(j.max_violations, u.max_violations, 5) as max_violations
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_jadwalujian j ON j.idjadwalujian = jp.idjadwalujian
		LEFT JOIN cat.at_ujian u ON u.idujian = j.idujian
		WHERE jp.idjadwalujian = $1 AND jp.kodepeserta = $2
	`, scheduleID, participantCode)

	durationSec := participantState.WaktuPengerjaan * 60
	remainingSec := durationSec
	if participantState.RemainingSeconds != nil && *participantState.RemainingSeconds > 0 {
		remainingSec = *participantState.RemainingSeconds
	} else if participantState.TglMulai != nil {
		elapsed := int(time.Now().Sub(*participantState.TglMulai).Seconds())
		remainingSec = durationSec - elapsed
		if remainingSec < 0 {
			remainingSec = 0
		}
	}

	// Load saved answers
	userAnswers := make(map[int]dto.SavedUserAnswerDTO)
	type answerRow struct {
		NoSoal     int  `db:"nosoal"`
		Jawaban    int  `db:"jawaban"`
		IsDoubtful bool `db:"is_doubtful"`
	}
	var dbAnswers []answerRow
	_ = r.db.SelectContext(ctx, &dbAnswers, `
		SELECT nourutpeserta::integer as nosoal, 
		       COALESCE(jawabanpilih::integer, 0) as jawaban, 
		       COALESCE((snapshot_metadata->>'is_doubtful')::boolean, false) as is_doubtful
		FROM cat.at_jawabanpeserta
		WHERE idjadwalujian = $1 AND kodepeserta = $2 AND (softdelete = '0' OR softdelete IS NULL)
	`, scheduleID, participantCode)
	for _, ans := range dbAnswers {
		userAnswers[ans.NoSoal] = dto.SavedUserAnswerDTO{
			SelectedOptionKey: ans.Jawaban,
			IsDoubtful:        ans.IsDoubtful,
		}
	}

	log.Printf("[DEBUG_SHUFFLE] Participant %s Schedule %d: hasExistingSnapshot=%v (len=%d)", participantCode, scheduleID, participantState.SessionSnapshot.Valid, len(participantState.SessionSnapshot.String))
	if participantState.SessionSnapshot.Valid && len(participantState.SessionSnapshot.String) > 10 {
		var stored StoredSessionSnapshot
		if err := json.Unmarshal([]byte(participantState.SessionSnapshot.String), &stored); err == nil && len(stored.Questions) > 0 {
			// Check if snapshot is valid and question count matches expected configuration:
			// If expectedTotal is set, but the stored snapshot has a mismatched question count AND participant hasn't answered yet,
			// allow snapshot regeneration so that the configured quota is strictly respected.
			if expectedTotal > 0 && len(stored.Questions) != expectedTotal && len(userAnswers) == 0 {
				// Mismatched question count on unstarted/unanswered exam, proceed to regenerate below!
			} else {
				var clientList []dto.SessionQuestionItemDTO
				for _, q := range stored.Questions {
					wCorr := q.WeightCorrect
					if wCorr <= 0 {
						wCorr = 1.0
					}
					clientList = append(clientList, dto.SessionQuestionItemDTO{
						QuestionNumber:    q.QuestionNumber,
						QuestionText:      q.QuestionText,
						QuestionMediaType: q.QuestionMediaType,
						QuestionMediaURL:  q.QuestionMediaURL,
						Options:           q.Options,
						CategoryTag:       q.CategoryTag,
						QuestionType:      q.QuestionType,
						WeightCorrect:     wCorr,
						WeightWrong:       q.WeightWrong,
						StimulusID:        q.StimulusID,
						StimulusItemID:    q.StimulusItemID,
						StimulusTitle:     q.StimulusTitle,
						StimulusText:      q.StimulusText,
						StimulusMediaType: q.StimulusMediaType,
						StimulusMediaURL:  q.StimulusMediaURL,
					})
				}
				maxV := participantState.MaxViolations
				if maxV <= 0 {
					maxV = 5
				}
				return &dto.SessionQuestionsResponseDTO{
					ScheduleID:       stored.ScheduleID,
					QuestionBankCode: stored.QuestionBankCode,
					TotalQuestions:   stored.TotalQuestions,
					SnapshotVersion:  stored.SnapshotVersion,
					RemainingSeconds: remainingSec,
					MaxViolations:    maxV,
					UserAnswers:      userAnswers,
					Questions:        clientList,
				}, nil
			}
		}
	}

	// Helper to resolve media type and URL for questions
	resolveMediaHelper := func(img, aud, vid *string) (string, *string) {
		if img != nil && *img != "" {
			return "IMAGE", img
		}
		if aud != nil && *aud != "" {
			return "AUDIO", aud
		}
		if vid != nil && *vid != "" {
			return "VIDEO", vid
		}
		return "NONE", nil
	}

	type candidateQuestionItem struct {
		KodeSoal          string
		OriginalNoUrut    int
		QuestionText      string
		QuestionMediaType string
		QuestionMediaURL  *string
		Options           []dto.QuestionOptionDTO
		CorrectOptionKey  int
		CategoryTag       string
		QuestionType      string
		WeightCorrect     float64
		WeightWrong       float64
		StimulusID        *int64
		StimulusItemID    *int64
		StimulusTitle     *string
		StimulusText      *string
		StimulusMediaType string
		StimulusMediaURL  *string
	}

	type questionGroup struct {
		StimulusID *int64
		Questions  []candidateQuestionItem
	}

	var allSelectedGroups []questionGroup

	// If no configs in at_soalujian, fallback to default kodesoal
	targetConfigs := configs
	if len(targetConfigs) == 0 {
		targetConfigs = []soalUjianConfig{
			{KodeSoal: kodesoal, JumlahPertanyaan: 0, NoUrut: 1},
		}
	}

	for _, cfg := range targetConfigs {
		bankCode := cfg.KodeSoal
		if bankCode == "" {
			bankCode = kodesoal
		}

		var bankItems []candidateQuestionItem

		// 1. Fetch Static Questions from cat.at_pertanyaan
		type rawQ struct {
			NoUrut       int     `db:"nourut"`
			KodeSoal     string  `db:"kodesoal"`
			Pertanyaan   string  `db:"pertanyaan"`
			QImg         *string `db:"pertanyaanimage"`
			QAud         *string `db:"pertanyaanaudio"`
			QVid         *string `db:"pertanyaanvideo"`
			J1           string  `db:"jawaban1"`
			J1Img        *string `db:"jawaban1image"`
			J1Aud        *string `db:"jawaban1audio"`
			J1Vid        *string `db:"jawaban1video"`
			J2           string  `db:"jawaban2"`
			J2Img        *string `db:"jawaban2image"`
			J2Aud        *string `db:"jawaban2audio"`
			J2Vid        *string `db:"jawaban2video"`
			J3           string  `db:"jawaban3"`
			J3Img        *string `db:"jawaban3image"`
			J3Aud        *string `db:"jawaban3audio"`
			J3Vid        *string `db:"jawaban3video"`
			J4           string  `db:"jawaban4"`
			J4Img        *string `db:"jawaban4image"`
			J4Aud        *string `db:"jawaban4audio"`
			J4Vid        *string `db:"jawaban4video"`
			J5           string  `db:"jawaban5"`
			J5Img        *string `db:"jawaban5image"`
			J5Aud        *string `db:"jawaban5audio"`
			J5Vid        *string `db:"jawaban5video"`
			JawabanBenar int     `db:"jawabanbenar"`
			BobotBenar   float64 `db:"bobot_benar"`
			BobotSalah   float64 `db:"bobot_salah"`
			KataKunci    string  `db:"katakunci"`
		}
		var staticRows []rawQ
		_ = r.db.SelectContext(ctx, &staticRows, `
			SELECT nourut, kodesoal, COALESCE(pertanyaan, '') as pertanyaan,
			       pertanyaanimage, pertanyaanaudio, pertanyaanvideo,
			       COALESCE(jawaban1, '') as jawaban1, jawaban1image, jawaban1audio, jawaban1video,
			       COALESCE(jawaban2, '') as jawaban2, jawaban2image, jawaban2audio, jawaban2video,
			       COALESCE(jawaban3, '') as jawaban3, jawaban3image, jawaban3audio, jawaban3video,
			       COALESCE(jawaban4, '') as jawaban4, jawaban4image, jawaban4audio, jawaban4video,
			       COALESCE(jawaban5, '') as jawaban5, jawaban5image, jawaban5audio, jawaban5video,
			       COALESCE(jawabanbenar, 1) as jawabanbenar,
			       COALESCE(bobot_benar, COALESCE(bobot, 1.0)) as bobot_benar,
			       COALESCE(bobot_salah, 0.0) as bobot_salah,
			       COALESCE(katakunci, 'Umum') as katakunci
			FROM cat.at_pertanyaan
			WHERE kodesoal = $1 AND (softdelete = '0' OR softdelete IS NULL)
			ORDER BY nourut ASC
		`, bankCode)

		for _, q := range staticRows {
			qMType, qMURL := resolveMediaHelper(q.QImg, q.QAud, q.QVid)
			j1Type, j1URL := resolveMediaHelper(q.J1Img, q.J1Aud, q.J1Vid)
			j2Type, j2URL := resolveMediaHelper(q.J2Img, q.J2Aud, q.J2Vid)
			j3Type, j3URL := resolveMediaHelper(q.J3Img, q.J3Aud, q.J3Vid)
			j4Type, j4URL := resolveMediaHelper(q.J4Img, q.J4Aud, q.J4Vid)
			j5Type, j5URL := resolveMediaHelper(q.J5Img, q.J5Aud, q.J5Vid)

			opts := []dto.QuestionOptionDTO{
				{OptionKey: 1, OptionText: q.J1, MediaType: j1Type, MediaURL: j1URL},
				{OptionKey: 2, OptionText: q.J2, MediaType: j2Type, MediaURL: j2URL},
				{OptionKey: 3, OptionText: q.J3, MediaType: j3Type, MediaURL: j3URL},
				{OptionKey: 4, OptionText: q.J4, MediaType: j4Type, MediaURL: j4URL},
				{OptionKey: 5, OptionText: q.J5, MediaType: j5Type, MediaURL: j5URL},
			}

			wCorr := q.BobotBenar
			if wCorr <= 0 {
				wCorr = 1.0
			}

			bankItems = append(bankItems, candidateQuestionItem{
				KodeSoal:          bankCode,
				OriginalNoUrut:    q.NoUrut,
				QuestionText:      q.Pertanyaan,
				QuestionMediaType: qMType,
				QuestionMediaURL:  qMURL,
				Options:           opts,
				CorrectOptionKey:  q.JawabanBenar,
				CategoryTag:       q.KataKunci,
				QuestionType:      "STATIC",
				WeightCorrect:     wCorr,
				WeightWrong:       q.BobotSalah,
			})
		}

		// 2. Fetch Dynamic Stimulus Items from cat.cat_bank_stimulus + cat.cat_stimulus_items
		type rawStimRow struct {
			IDStimulus        int64   `db:"id_stimulus"`
			Title             string  `db:"title"`
			NarrativeText     string  `db:"narrative_text"`
			StimMediaType     string  `db:"stim_media_type"`
			StimMediaURL      *string `db:"stim_media_url"`
			CategoryTag       string  `db:"category_tag"`
			IDItem            int64   `db:"id_item"`
			ItemOrder         int     `db:"item_order"`
			QuestionText      string  `db:"question_text"`
			QuestionMediaType string  `db:"question_media_type"`
			QuestionMediaURL  *string `db:"question_media_url"`
			CorrectAnswer     string  `db:"correct_answer"`
			WeightCorrect     float64 `db:"weight_correct"`
			WeightWrong       float64 `db:"weight_wrong"`
		}
		var dynamicRows []rawStimRow
		_ = r.db.SelectContext(ctx, &dynamicRows, `
			SELECT s.id_stimulus, s.title, s.narrative_text, 
			       COALESCE(s.media_type, 'NONE') as stim_media_type, s.media_url as stim_media_url,
			       COALESCE(s.category_tag, 'Kasus Klinis') as category_tag,
			       i.id_item, i.item_order, i.question_text,
			       COALESCE(i.question_media_type, 'NONE') as question_media_type, i.question_media_url,
			       COALESCE(i.correct_answer, 'A') as correct_answer,
			       COALESCE(i.weight_correct, 1.0) as weight_correct,
			       COALESCE(i.weight_wrong, 0.0) as weight_wrong
			FROM cat.cat_stimulus_items i
			JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
			WHERE s.kodesoal = $1 AND (s.softdelete = '0' OR s.softdelete IS NULL) AND (i.softdelete = '0' OR i.softdelete IS NULL)
			ORDER BY s.sort_order ASC, s.id_stimulus ASC, i.item_order ASC, i.id_item ASC
		`, bankCode)

		if len(dynamicRows) > 0 {
			// Fetch options for all dynamic items
			type rawOptRow struct {
				IDOption    int64   `db:"id_option"`
				IDItem      int64   `db:"id_item"`
				OptionLabel string  `db:"option_label"`
				OptionText  string  `db:"option_text"`
				MediaType   string  `db:"media_type"`
				MediaURL    *string `db:"media_url"`
			}
			var optRows []rawOptRow
			_ = r.db.SelectContext(ctx, &optRows, `
				SELECT o.id_option, o.id_item, o.option_label, o.option_text,
				       COALESCE(o.media_type, 'NONE') as media_type, o.media_url
				FROM cat.cat_item_options o
				JOIN cat.cat_stimulus_items i ON i.id_item = o.id_item
				JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
				WHERE s.kodesoal = $1
				ORDER BY o.sort_order ASC, o.option_label ASC
			`, bankCode)

			optMap := make(map[int64][]rawOptRow)
			for _, o := range optRows {
				optMap[o.IDItem] = append(optMap[o.IDItem], o)
			}

			for _, d := range dynamicRows {
				itemOpts := optMap[d.IDItem]
				var optDTOs []dto.QuestionOptionDTO
				correctKey := 1
				cleanCorrect := strings.ToUpper(strings.TrimSpace(d.CorrectAnswer))

				if len(itemOpts) > 0 {
					for oIdx, op := range itemOpts {
						key := oIdx + 1
						cleanLabel := strings.ToUpper(strings.TrimSpace(op.OptionLabel))
						if cleanLabel == "A" {
							key = 1
						}
						if cleanLabel == "B" {
							key = 2
						}
						if cleanLabel == "C" {
							key = 3
						}
						if cleanLabel == "D" {
							key = 4
						}
						if cleanLabel == "E" {
							key = 5
						}

						if cleanCorrect == cleanLabel || cleanCorrect == fmt.Sprintf("%d", key) {
							correctKey = key
						}
						optDTOs = append(optDTOs, dto.QuestionOptionDTO{
							OptionKey:  key,
							OptionText: op.OptionText,
							MediaType:  op.MediaType,
							MediaURL:   op.MediaURL,
						})
					}
				} else {
					optDTOs = []dto.QuestionOptionDTO{
						{OptionKey: 1, OptionText: "Pilihan A"},
						{OptionKey: 2, OptionText: "Pilihan B"},
						{OptionKey: 3, OptionText: "Pilihan C"},
						{OptionKey: 4, OptionText: "Pilihan D"},
						{OptionKey: 5, OptionText: "Pilihan E"},
					}
					if cleanCorrect == "B" {
						correctKey = 2
					}
					if cleanCorrect == "C" {
						correctKey = 3
					}
					if cleanCorrect == "D" {
						correctKey = 4
					}
					if cleanCorrect == "E" {
						correctKey = 5
					}
				}

				stimID := d.IDStimulus
				itemID := d.IDItem
				stimTitle := d.Title
				stimText := d.NarrativeText

				wCorr := d.WeightCorrect
				if wCorr <= 0 {
					wCorr = 1.0
				}

				bankItems = append(bankItems, candidateQuestionItem{
					KodeSoal:          bankCode,
					OriginalNoUrut:    d.ItemOrder,
					QuestionText:      d.QuestionText,
					QuestionMediaType: d.QuestionMediaType,
					QuestionMediaURL:  d.QuestionMediaURL,
					Options:           optDTOs,
					CorrectOptionKey:  correctKey,
					CategoryTag:       d.CategoryTag,
					QuestionType:      "DYNAMIC_STIMULUS",
					WeightCorrect:     wCorr,
					WeightWrong:       d.WeightWrong,
					StimulusID:        &stimID,
					StimulusItemID:    &itemID,
					StimulusTitle:     &stimTitle,
					StimulusText:      &stimText,
					StimulusMediaType: d.StimMediaType,
					StimulusMediaURL:  d.StimMediaURL,
				})
			}
		}

		// Separate into static candidate list and dynamic stimulus groups (Model 1 Sampling)
		var staticList []candidateQuestionItem
		var dynamicStimMap = make(map[int64][]candidateQuestionItem)
		var dynamicStimOrder []int64

		for _, item := range bankItems {
			if item.StimulusID == nil {
				staticList = append(staticList, item)
			} else {
				sID := *item.StimulusID
				if _, exists := dynamicStimMap[sID]; !exists {
					dynamicStimOrder = append(dynamicStimOrder, sID)
				}
				dynamicStimMap[sID] = append(dynamicStimMap[sID], item)
			}
		}

		targetStatic := cfg.JumlahSoalStatis
		targetKasus := cfg.JumlahKasusDinamis

		// Backward compatibility: If only JumlahPertanyaan was specified under legacy config
		if targetStatic == 0 && targetKasus == 0 && cfg.JumlahPertanyaan > 0 {
			if len(dynamicStimOrder) > 0 && len(staticList) == 0 {
				targetKasus = cfg.JumlahPertanyaan
			} else if len(staticList) > 0 && len(dynamicStimOrder) == 0 {
				targetStatic = cfg.JumlahPertanyaan
			} else {
				targetStatic = cfg.JumlahPertanyaan
			}
		}

		// A. Sample Static Questions
		var sampledStaticGroups []questionGroup
		if targetStatic > 0 && len(staticList) > 0 {
			rBank := rand.New(rand.NewSource(time.Now().UnixNano()))
			rBank.Shuffle(len(staticList), func(i, j int) {
				staticList[i], staticList[j] = staticList[j], staticList[i]
			})
			takeCount := targetStatic
			if takeCount > len(staticList) {
				takeCount = len(staticList)
			}
			for i := 0; i < takeCount; i++ {
				sampledStaticGroups = append(sampledStaticGroups, questionGroup{
					StimulusID: nil,
					Questions:  []candidateQuestionItem{staticList[i]},
				})
			}
		} else if targetStatic <= 0 && len(staticList) > 0 && targetKasus <= 0 && cfg.JumlahPertanyaan <= 0 {
			for _, item := range staticList {
				sampledStaticGroups = append(sampledStaticGroups, questionGroup{
					StimulusID: nil,
					Questions:  []candidateQuestionItem{item},
				})
			}
		}

		// B. Sample Dynamic Clinical Cases (ATOMIC: each case includes all items contiguous and intact)
		var sampledDynamicGroups []questionGroup
		if targetKasus > 0 && len(dynamicStimOrder) > 0 {
			rBank := rand.New(rand.NewSource(time.Now().UnixNano()))
			rBank.Shuffle(len(dynamicStimOrder), func(i, j int) {
				dynamicStimOrder[i], dynamicStimOrder[j] = dynamicStimOrder[j], dynamicStimOrder[i]
			})
			takeKasus := targetKasus
			if takeKasus > len(dynamicStimOrder) {
				takeKasus = len(dynamicStimOrder)
			}
			for i := 0; i < takeKasus; i++ {
				sID := dynamicStimOrder[i]
				items := dynamicStimMap[sID]
				sampledDynamicGroups = append(sampledDynamicGroups, questionGroup{
					StimulusID: &sID,
					Questions:  items,
				})
			}
		} else if targetKasus <= 0 && len(dynamicStimOrder) > 0 && targetStatic <= 0 && cfg.JumlahPertanyaan <= 0 {
			for _, sID := range dynamicStimOrder {
				items := dynamicStimMap[sID]
				idCopy := sID
				sampledDynamicGroups = append(sampledDynamicGroups, questionGroup{
					StimulusID: &idCopy,
					Questions:  items,
				})
			}
		}

		allSelectedGroups = append(allSelectedGroups, sampledStaticGroups...)
		allSelectedGroups = append(allSelectedGroups, sampledDynamicGroups...)
	}

	// Last resort fallback if no questions were found at all
	if len(allSelectedGroups) == 0 {
		type fallbackRow struct {
			NoUrut       int     `db:"nourut"`
			KodeSoal     string  `db:"kodesoal"`
			Pertanyaan   string  `db:"pertanyaan"`
			QImg         *string `db:"pertanyaanimage"`
			QAud         *string `db:"pertanyaanaudio"`
			QVid         *string `db:"pertanyaanvideo"`
			J1           string  `db:"jawaban1"`
			J2           string  `db:"jawaban2"`
			J3           string  `db:"jawaban3"`
			J4           string  `db:"jawaban4"`
			J5           string  `db:"jawaban5"`
			JawabanBenar int     `db:"jawabanbenar"`
			KataKunci    string  `db:"katakunci"`
		}
		var fallbackStatic []fallbackRow
		_ = r.db.SelectContext(ctx, &fallbackStatic, `
			SELECT nourut, kodesoal, COALESCE(pertanyaan, '') as pertanyaan,
			       pertanyaanimage, pertanyaanaudio, pertanyaanvideo,
			       COALESCE(jawaban1, '') as jawaban1,
			       COALESCE(jawaban2, '') as jawaban2,
			       COALESCE(jawaban3, '') as jawaban3,
			       COALESCE(jawaban4, '') as jawaban4,
			       COALESCE(jawaban5, '') as jawaban5,
			       COALESCE(jawabanbenar, 1) as jawabanbenar,
			       COALESCE(katakunci, 'Umum') as katakunci
			FROM cat.at_pertanyaan
			WHERE softdelete = '0' OR softdelete IS NULL
			ORDER BY RANDOM() LIMIT 50
		`)
		for _, q := range fallbackStatic {
			qMType, qMURL := resolveMediaHelper(q.QImg, q.QAud, q.QVid)
			opts := []dto.QuestionOptionDTO{
				{OptionKey: 1, OptionText: q.J1},
				{OptionKey: 2, OptionText: q.J2},
				{OptionKey: 3, OptionText: q.J3},
				{OptionKey: 4, OptionText: q.J4},
				{OptionKey: 5, OptionText: q.J5},
			}
			allSelectedGroups = append(allSelectedGroups, questionGroup{
				StimulusID: nil,
				Questions: []candidateQuestionItem{
					{
						KodeSoal:          q.KodeSoal,
						OriginalNoUrut:    q.NoUrut,
						QuestionText:      q.Pertanyaan,
						QuestionMediaType: qMType,
						QuestionMediaURL:  qMURL,
						Options:           opts,
						CorrectOptionKey:  q.JawabanBenar,
						CategoryTag:       q.KataKunci,
						QuestionType:      "STATIC",
					},
				},
			})
		}
	}

	// 🎲 RANDOMIZE / SHUFFLE QUESTIONS:
	// Shuffle all question groups so every participant receives a distinct, randomized question order.
	// Questions belonging to the same stimulus remain logically contiguous.
	examRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	examRand.Shuffle(len(allSelectedGroups), func(i, j int) {
		allSelectedGroups[i], allSelectedGroups[j] = allSelectedGroups[j], allSelectedGroups[i]
	})

	// Flatten groups into final ordered questions
	var finalQuestions []candidateQuestionItem
	for _, g := range allSelectedGroups {
		finalQuestions = append(finalQuestions, g.Questions...)
	}

	// Strict Hard Cap: Only enforce if dynamic cases were NOT explicitly configured (to avoid cutting a clinical case in half)
	if !hasDynamicQuotaConfigured && expectedTotal > 0 && len(finalQuestions) > expectedTotal {
		finalQuestions = finalQuestions[:expectedTotal]
	}

	var storedItems []StoredQuestionSnapshotItem
	var clientItems []dto.SessionQuestionItemDTO

	for idx, q := range finalQuestions {
		qNum := idx + 1
		bankCode := q.KodeSoal
		if bankCode == "" {
			bankCode = kodesoal
		}

		wCorr := q.WeightCorrect
		if wCorr <= 0 {
			wCorr = 1.0
		}
		wWrng := q.WeightWrong

		// 🎲 RANDOMIZE / SHUFFLE OPTIONS (A-E):
		// Shuffle the options uniquely per participant so option letters (A, B, C, D, E) differ between participants.
		shuffledOptions := make([]dto.QuestionOptionDTO, len(q.Options))
		copy(shuffledOptions, q.Options)
		var optMapping [5]int
		for i := 0; i < 5; i++ {
			optMapping[i] = i + 1
		}
		mappedCorrectKey := q.CorrectOptionKey

		if len(shuffledOptions) > 1 {
			type optWithOrig struct {
				origKey int
				opt     dto.QuestionOptionDTO
			}
			origList := make([]optWithOrig, len(shuffledOptions))
			for i, o := range shuffledOptions {
				origList[i] = optWithOrig{origKey: o.OptionKey, opt: o}
			}
			examRand.Shuffle(len(origList), func(i, j int) {
				origList[i], origList[j] = origList[j], origList[i]
			})

			shuffledOptions = make([]dto.QuestionOptionDTO, len(origList))
			for newIdx, item := range origList {
				newKey := newIdx + 1
				shuffledOptions[newIdx] = dto.QuestionOptionDTO{
					OptionKey:  newKey,
					OptionText: item.opt.OptionText,
					MediaType:  item.opt.MediaType,
					MediaURL:   item.opt.MediaURL,
				}
				if item.origKey == q.CorrectOptionKey {
					mappedCorrectKey = newKey
				}
				if newIdx < 5 {
					optMapping[newIdx] = item.origKey
				}
			}
		}

		log.Printf("[DEBUG_SHUFFLE] Participant %s Q%d: origOpts=%d, mappedCorrectKey=%d, optMapping=%v", participantCode, qNum, len(q.Options), mappedCorrectKey, optMapping)

		storedItems = append(storedItems, StoredQuestionSnapshotItem{
			QuestionNumber:    qNum,
			OriginalNoUrut:    q.OriginalNoUrut,
			KodeSoal:          bankCode,
			QuestionText:      q.QuestionText,
			QuestionMediaType: q.QuestionMediaType,
			QuestionMediaURL:  q.QuestionMediaURL,
			Options:           shuffledOptions,
			CorrectOptionKey:  mappedCorrectKey,
			CategoryTag:       q.CategoryTag,
			OptionMapping:     optMapping,
			QuestionType:      q.QuestionType,
			WeightCorrect:     wCorr,
			WeightWrong:       wWrng,
			StimulusID:        q.StimulusID,
			StimulusItemID:    q.StimulusItemID,
			StimulusTitle:     q.StimulusTitle,
			StimulusText:      q.StimulusText,
			StimulusMediaType: q.StimulusMediaType,
			StimulusMediaURL:  q.StimulusMediaURL,
		})

		clientItems = append(clientItems, dto.SessionQuestionItemDTO{
			QuestionNumber:    qNum,
			QuestionText:      q.QuestionText,
			QuestionMediaType: q.QuestionMediaType,
			QuestionMediaURL:  q.QuestionMediaURL,
			Options:           shuffledOptions,
			CategoryTag:       q.CategoryTag,
			QuestionType:      q.QuestionType,
			WeightCorrect:     wCorr,
			WeightWrong:       wWrng,
			StimulusID:        q.StimulusID,
			StimulusItemID:    q.StimulusItemID,
			StimulusTitle:     q.StimulusTitle,
			StimulusText:      q.StimulusText,
			StimulusMediaType: q.StimulusMediaType,
			StimulusMediaURL:  q.StimulusMediaURL,
		})
	}

	snapshot := StoredSessionSnapshot{
		ScheduleID:       scheduleID,
		QuestionBankCode: kodesoal,
		TotalQuestions:   len(storedItems),
		SnapshotVersion:  1,
		CapturedAt:       time.Now().UTC().Format(time.RFC3339),
		Questions:        storedItems,
	}

	snapshotBytes, _ := json.Marshal(snapshot)
	_, _ = r.db.ExecContext(ctx, `UPDATE cat.at_jadwalpeserta SET session_snapshot = $1::jsonb WHERE idjadwalujian = $2 AND kodepeserta = $3`,
		string(snapshotBytes), scheduleID, participantCode)

	maxV := participantState.MaxViolations
	if maxV <= 0 {
		maxV = 5
	}
	return &dto.SessionQuestionsResponseDTO{
		ScheduleID:       scheduleID,
		QuestionBankCode: kodesoal,
		TotalQuestions:   len(clientItems),
		SnapshotVersion:  1,
		RemainingSeconds: remainingSec,
		MaxViolations:    maxV,
		UserAnswers:      userAnswers,
		Questions:        clientItems,
	}, nil
}

func (r *examinationRepository) GetSessionQuestions(ctx context.Context, participantCode string, scheduleID int) (*dto.SessionQuestionsResponseDTO, error) {
	var kodesoal string
	_ = r.db.GetContext(ctx, &kodesoal, `SELECT COALESCE(su.kodesoal, 'SOAL_2026') FROM cat.at_soalujian su WHERE su.idjadwalujian = $1 LIMIT 1`, scheduleID)
	if kodesoal == "" {
		kodesoal = "SOAL_2026"
	}
	return r.buildAndSaveSessionSnapshot(ctx, scheduleID, participantCode, kodesoal)
}

func (r *examinationRepository) SaveAnswer(ctx context.Context, participantCode string, scheduleID int, questionBankCode string, questionNumber int, selectedOption int, isDoubtful bool) error {
	// Check if session is locked or submitted
	var statusCheck struct {
		IsLocked         int            `db:"is_locked"`
		TglSelesai       *time.Time     `db:"tglselesai"`
		TglMulai         *time.Time     `db:"tglmulai"`
		WaktuMenit       int            `db:"waktupengerjaan"`
		RemainingSeconds *int           `db:"remaining_seconds"`
		SessionSnapshot  sql.NullString `db:"session_snapshot"`
	}
	checkQuery := `
		SELECT COALESCE(jp.is_locked, 0) as is_locked, jp.tglselesai, jp.tglmulai,
		       COALESCE(j.waktupengerjaan::integer, 90) as waktupengerjaan,
		       jp.remaining_seconds, jp.session_snapshot::text as session_snapshot
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_jadwalujian j ON j.idjadwalujian = jp.idjadwalujian
		WHERE jp.idjadwalujian = $1 AND jp.kodepeserta = $2
	`
	if err := r.db.GetContext(ctx, &statusCheck, checkQuery, scheduleID, participantCode); err == nil {
		if statusCheck.IsLocked == 1 {
			return fmt.Errorf("sesi ujian Anda telah dikunci oleh pengawas karena terindikasi pelanggaran")
		}
		if statusCheck.TglSelesai != nil {
			return fmt.Errorf("lembar ujian telah selesai dan dikunci, jawaban tidak dapat diubah")
		}
		if statusCheck.RemainingSeconds != nil && *statusCheck.RemainingSeconds <= 0 {
			return fmt.Errorf("waktu ujian telah habis (timeout)")
		}
		if statusCheck.RemainingSeconds == nil && statusCheck.TglMulai != nil {
			elapsed := time.Since(*statusCheck.TglMulai).Seconds()
			maxAllowed := float64(statusCheck.WaktuMenit*60 + 120) // grace period 120s
			if elapsed > maxAllowed {
				return fmt.Errorf("waktu ujian telah habis (timeout)")
			}
		}
	}

	jsonMeta := fmt.Sprintf(`{"is_doubtful": %t}`, isDoubtful)

	// Determine isbenar, score, actual bank and original nourut from session snapshot if available
	isBenar := 0
	scoreVal := 0.0
	actualBankCode := questionBankCode
	actualNoUrut := questionNumber

	if statusCheck.SessionSnapshot.Valid && len(statusCheck.SessionSnapshot.String) > 10 {
		var stored StoredSessionSnapshot
		if err := json.Unmarshal([]byte(statusCheck.SessionSnapshot.String), &stored); err == nil {
			for _, q := range stored.Questions {
				if q.QuestionNumber == questionNumber {
					if q.KodeSoal != "" {
						actualBankCode = q.KodeSoal
					}
					if q.OriginalNoUrut > 0 {
						actualNoUrut = q.OriginalNoUrut
					}
					if selectedOption > 0 && selectedOption == q.CorrectOptionKey {
						isBenar = 1
						wCorr := q.WeightCorrect
						if wCorr <= 0 {
							wCorr = 1.0
						}
						scoreVal = wCorr
					} else if selectedOption > 0 && q.WeightWrong != 0 {
						isBenar = 0
						scoreVal = q.WeightWrong
					}
					break
				}
			}
		}
	}

	// Atomic PostgreSQL UPSERT on primary key (kodepeserta, idjadwalujian, nourutpeserta)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cat.at_jawabanpeserta 
			(kodepeserta, idjadwalujian, nourutpeserta, kodesoal, nourut, jawabanpilih, isbenar, nilai, snapshot_metadata, softdelete, t_updatetime) 
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, '0', NOW())
		ON CONFLICT (kodepeserta, idjadwalujian, nourutpeserta) DO UPDATE
		SET jawabanpilih = EXCLUDED.jawabanpilih,
		    isbenar = EXCLUDED.isbenar,
		    nilai = EXCLUDED.nilai,
		    snapshot_metadata = EXCLUDED.snapshot_metadata,
		    kodesoal = EXCLUDED.kodesoal,
		    nourut = EXCLUDED.nourut,
		    softdelete = '0',
		    t_updatetime = NOW()
	`, participantCode, scheduleID, questionNumber, actualBankCode, actualNoUrut, selectedOption, isBenar, scoreVal, jsonMeta)
	return err
}

func (r *examinationRepository) RecordSecurityEvent(ctx context.Context, event *entity.SecurityEvent) (*dto.SecurityEventResponseDTO, error) {
	var maxViolations int = 5
	_ = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(j.max_violations, u.max_violations, 5)
		FROM cat.at_jadwalujian j
		LEFT JOIN cat.at_ujian u ON u.idujian = j.idujian
		WHERE j.idjadwalujian = $1
	`, event.IDJadwal).Scan(&maxViolations)
	if maxViolations <= 0 {
		maxViolations = 5
	}

	// 1. Identify if this is a violation event
	isViolationEvent := false
	switch event.EventType {
	case "TAB_SWITCH", "WINDOW_BLUR", "FULLSCREEN_EXIT", "FULLSCREEN_REQUIRED",
		"CAMERA_DISCONNECTED", "FORBIDDEN_SHORTCUT", "SCREENSHOT_ATTEMPT",
		"DEVTOOLS_ATTEMPT", "DEVTOOLS_OPEN", "PRINT_ATTEMPT", "COPY_ATTEMPT",
		"PASTE_ATTEMPT", "CUT_ATTEMPT", "MULTI_MONITOR", "VM_DETECTED",
		"UNTRUSTED_EVENT", "MULTI_FACE_DETECTED", "CAMERA_BLACKOUT":
		isViolationEvent = true
	}

	// 2. Cooldown debounce check (5 seconds filter)
	// Rapid consecutive browser events (e.g. blur + visibility + fullscreen exit) within 5s only count as 1 violation
	isCoolingDown := false
	if isViolationEvent {
		var recentCount int
		cooldownQuery := `
			SELECT COUNT(*) 
			FROM cat.at_security_events 
			WHERE idjadwal = $1 
			  AND kodepeserta = $2 
			  AND created_at >= (NOW() - INTERVAL '5 seconds')
			  AND event_type IN (
				  'TAB_SWITCH', 'WINDOW_BLUR', 'FULLSCREEN_EXIT', 'FULLSCREEN_REQUIRED',
				  'CAMERA_DISCONNECTED', 'FORBIDDEN_SHORTCUT', 'SCREENSHOT_ATTEMPT',
				  'DEVTOOLS_ATTEMPT', 'DEVTOOLS_OPEN', 'PRINT_ATTEMPT', 'COPY_ATTEMPT',
				  'PASTE_ATTEMPT', 'CUT_ATTEMPT', 'MULTI_MONITOR', 'VM_DETECTED',
				  'UNTRUSTED_EVENT', 'MULTI_FACE_DETECTED', 'CAMERA_BLACKOUT'
			  )
		`
		_ = r.db.QueryRowContext(ctx, cooldownQuery, event.IDJadwal, event.KodePeserta).Scan(&recentCount)
		if recentCount > 0 {
			isCoolingDown = true
		}
	}

	// 3. Insert audit event into cat.at_security_events (always logged for audit trail)
	insertQuery := `
		INSERT INTO cat.at_security_events (
			idjadwal, kodepeserta, event_type, severity, risk_score, 
			metadata, ip_address, user_agent, device_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, NOW())
	`
	metaStr := "{}"
	if event.Metadata != nil {
		metaStr = *event.Metadata
	}
	if isCoolingDown {
		trimmed := strings.TrimSpace(metaStr)
		if trimmed == "" || trimmed == "{}" {
			metaStr = `{"cooldown_suppressed":true}`
		} else if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
			metaStr = trimmed[:len(trimmed)-1] + `,"cooldown_suppressed":true}`
		}
	}
	_, _ = r.db.ExecContext(ctx, insertQuery,
		event.IDJadwal, event.KodePeserta, event.EventType, event.Severity, event.RiskScore,
		metaStr, event.IPAddress, event.UserAgent, event.DeviceID,
	)

	// 4. Calculate effective counter increments & risk score
	isTabSwitch := 0
	isFsExit := 0
	effectiveRiskScore := event.RiskScore

	if isCoolingDown {
		// Cooldown active: Suppress counter & risk increments for rapid event bursts within 5s
		isTabSwitch = 0
		isFsExit = 0
		effectiveRiskScore = 0
	} else {
		if isViolationEvent {
			isTabSwitch = 1
		}
		if event.EventType == "FULLSCREEN_EXIT" || event.EventType == "FULLSCREEN_REQUIRED" {
			isFsExit = 1
		}
	}

	// 5. Update with AUTO-LOCK: lock session when tab_switch_count >= maxViolations OR risk_score >= 81
	updateQuery := `
		UPDATE cat.at_jadwalpeserta
		SET risk_score = COALESCE(risk_score, 0) + $1,
		    tab_switch_count = COALESCE(tab_switch_count, 0) + $2,
		    fullscreen_exit_count = COALESCE(fullscreen_exit_count, 0) + $3,
		    risk_level = CASE 
		        WHEN (COALESCE(risk_score, 0) + $1) >= 81 THEN 'CRITICAL'
		        WHEN (COALESCE(risk_score, 0) + $1) >= 61 THEN 'HIGH_RISK'
		        WHEN (COALESCE(risk_score, 0) + $1) >= 41 THEN 'SUSPICIOUS'
		        WHEN (COALESCE(risk_score, 0) + $1) >= 21 THEN 'ATTENTION'
		        ELSE 'NORMAL'
		    END,
		    is_locked = CASE
		        WHEN COALESCE(is_locked, 0) = 1 THEN 1
		        WHEN (COALESCE(tab_switch_count, 0) + $2) >= $6 THEN 1
		        WHEN (COALESCE(risk_score, 0) + $1) >= 81 THEN 1
		        ELSE COALESCE(is_locked, 0)
		    END,
		    lock_reason = CASE
		        WHEN COALESCE(is_locked, 0) = 1 THEN lock_reason
		        WHEN (COALESCE(tab_switch_count, 0) + $2) >= $6 THEN 'Sesi dikunci otomatis: jumlah pelanggaran mencapai batas maksimal (' || $6::text || 'x). Hubungi pengawas untuk membuka kunci.'
		        WHEN (COALESCE(risk_score, 0) + $1) >= 81 THEN 'Sesi dikunci otomatis: skor risiko kecurangan mencapai level CRITICAL. Hubungi pengawas untuk membuka kunci.'
		        ELSE lock_reason
		    END
		WHERE idjadwalujian = $4 AND kodepeserta = $5
		RETURNING risk_score, risk_level, tab_switch_count, COALESCE(fullscreen_exit_count, 0), COALESCE(is_locked, 0)
	`
	var res dto.SecurityEventResponseDTO
	var prevLocked int
	// Read current lock status before update to detect auto-lock
	_ = r.db.QueryRowContext(ctx, `SELECT COALESCE(is_locked, 0) FROM cat.at_jadwalpeserta WHERE idjadwalujian = $1 AND kodepeserta = $2`, event.IDJadwal, event.KodePeserta).Scan(&prevLocked)

	err := r.db.QueryRowContext(ctx, updateQuery, effectiveRiskScore, isTabSwitch, isFsExit, event.IDJadwal, event.KodePeserta, maxViolations).
		Scan(&res.CurrentRiskScore, &res.RiskLevel, &res.TabSwitchCount, &res.FullscreenExitCount, &res.IsLocked)
	if err != nil {
		return &dto.SecurityEventResponseDTO{
			CurrentRiskScore: event.RiskScore,
			RiskLevel:        "NORMAL",
			TabSwitchCount:   isTabSwitch,
			IsLocked:         0,
			MaxViolations:    maxViolations,
		}, nil
	}

	// Detect if auto-lock just happened in this event
	res.AutoLocked = prevLocked == 0 && res.IsLocked == 1
	res.MaxViolations = maxViolations
	return &res, nil
}

func (r *examinationRepository) GetRecentSecurityEvents(ctx context.Context, scheduleID int, limit int) ([]*entity.SecurityEvent, error) {
	if limit <= 0 {
		limit = 30
	}
	query := `
		SELECT id, idjadwal, kodepeserta, event_type, severity, risk_score, 
		       metadata::text, ip_address, user_agent, device_id, created_at
		FROM cat.at_security_events
		WHERE idjadwal = $1
		ORDER BY id DESC
		LIMIT $2
	`
	var events []*entity.SecurityEvent
	err := r.db.SelectContext(ctx, &events, query, scheduleID, limit)
	return events, err
}

func (r *examinationRepository) ApplyProctorAction(ctx context.Context, scheduleID int, participantCode, action, reason, warning string) error {
	var query string
	switch action {
	case "LOCK_SESSION":
		query = `UPDATE cat.at_jadwalpeserta SET is_locked = 1, lock_reason = $1 WHERE idjadwalujian = $2 AND kodepeserta = $3`
		_, err := r.db.ExecContext(ctx, query, reason, scheduleID, participantCode)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{"action":"LOCK_SESSION","reason":%q,"note":"Sesi ujian dikunci manual oleh pengawas ruang"}`, reason)
		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO cat.at_security_events (idjadwal, kodepeserta, event_type, severity, risk_score, metadata, created_at)
			VALUES ($1, $2, 'PROCTOR_LOCK', 'CRITICAL', 50, $3::jsonb, NOW())
		`, scheduleID, participantCode, meta)
		return nil

	case "UNLOCK_SESSION":
		// Reset lock, tab_switch_count, fullscreen_exit_count, risk_score, risk_level, and clear warnings
		query = `
			UPDATE cat.at_jadwalpeserta 
			SET is_locked = 0, 
			    lock_reason = NULL, 
			    proctor_warning = NULL, 
			    tglselesai = NULL, 
			    tab_switch_count = 0, 
			    fullscreen_exit_count = 0, 
			    risk_score = 0, 
			    risk_level = 'NORMAL' 
			WHERE idjadwalujian = $1 AND kodepeserta = $2
		`
		_, err := r.db.ExecContext(ctx, query, scheduleID, participantCode)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{"action":"UNLOCK_SESSION","reason":%q,"note":"Kunci ujian dibuka oleh pengawas ruang (riwayat pelanggaran & skor risiko direset)"}`, reason)
		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO cat.at_security_events (idjadwal, kodepeserta, event_type, severity, risk_score, metadata, created_at)
			VALUES ($1, $2, 'PROCTOR_UNLOCK', 'INFO', 0, $3::jsonb, NOW())
		`, scheduleID, participantCode, meta)
		return nil

	case "SEND_WARNING":
		query = `UPDATE cat.at_jadwalpeserta SET proctor_warning = $1 WHERE idjadwalujian = $2 AND kodepeserta = $3`
		_, err := r.db.ExecContext(ctx, query, warning, scheduleID, participantCode)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{"action":"SEND_WARNING","warning":%q,"note":"Pengawas mengirimkan pesan peringatan kepada peserta"}`, warning)
		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO cat.at_security_events (idjadwal, kodepeserta, event_type, severity, risk_score, metadata, created_at)
			VALUES ($1, $2, 'PROCTOR_WARNING', 'WARNING', 10, $3::jsonb, NOW())
		`, scheduleID, participantCode, meta)
		return nil

	case "RESET_SESSION":
		query = `
			UPDATE cat.at_jadwalpeserta 
			SET is_locked = 0, 
			    lock_reason = NULL, 
			    proctor_warning = NULL, 
			    tglselesai = NULL, 
			    tab_switch_count = 0, 
			    fullscreen_exit_count = 0, 
			    risk_score = 0, 
			    risk_level = 'NORMAL', 
			    tglmulai = NOW(), 
			    remaining_seconds = NULL 
			WHERE idjadwalujian = $1 AND kodepeserta = $2
		`
		_, err := r.db.ExecContext(ctx, query, scheduleID, participantCode)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{"action":"RESET_SESSION","reason":%q,"note":"Sesi ujian direset ulang oleh pengawas ruang"}`, reason)
		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO cat.at_security_events (idjadwal, kodepeserta, event_type, severity, risk_score, metadata, created_at)
			VALUES ($1, $2, 'PROCTOR_RESET', 'WARNING', 0, $3::jsonb, NOW())
		`, scheduleID, participantCode, meta)
		return nil
	case "FORCE_SUBMIT":
		// 1. If dynamic exam session exists, finalize dynamic session first
		var hasDynamic int
		_ = r.db.GetContext(ctx, &hasDynamic, `SELECT COUNT(*) FROM cat.cat_participant_sessions_ext WHERE idjadwalujian = $1 AND kodepeserta = $2`, scheduleID, participantCode)
		if hasDynamic > 0 {
			_, _ = r.FinishDynamicExam(ctx, participantCode, scheduleID, "FORCE_SUBMIT")
		}

		// 2. Finalize static/snapshot exam
		_, _, _, _, _, _, _, err := r.FinishExam(ctx, participantCode, scheduleID, "FORCE_SUBMIT")

		// 3. Mark session explicitly completed and locked
		_, _ = r.db.ExecContext(ctx, `
			UPDATE cat.at_jadwalpeserta 
			SET is_locked = 1, 
			    lock_reason = 'Ujian diselesaikan paksa oleh pengawas ruang.', 
			    tglselesai = COALESCE(tglselesai, NOW()), 
			    submit_reason = 'FORCE_SUBMIT', 
			    remaining_seconds = 0 
			WHERE idjadwalujian = $1 AND kodepeserta = $2
		`, scheduleID, participantCode)

		return err
	default:
		return fmt.Errorf("aksi pengawas tidak dikenal: %s", action)
	}
}

func (r *examinationRepository) UpdateHeartbeatAndCheckStatus(ctx context.Context, scheduleID int, participantCode string, req *dto.HeartbeatRequestDTO) (*dto.HeartbeatResponseDTO, error) {
	now := time.Now()
	var row struct {
		IsLocked         int            `db:"is_locked"`
		LockReason       sql.NullString `db:"lock_reason"`
		ProctorWarning   sql.NullString `db:"proctor_warning"`
		TabSwitchCount   int            `db:"tab_switch_count"`
		RiskScore        int            `db:"risk_score"`
		WaktuPengerjaan  int            `db:"waktupengerjaan"`
		RemainingSeconds *int           `db:"remaining_seconds"`
		TglMulai         *time.Time     `db:"tglmulai"`
		TglSelesai       *time.Time     `db:"tglselesai"`
		SubmitReason     sql.NullString `db:"submit_reason"`
		DeadlineAt       *time.Time     `db:"exam_deadline_at"`
		MaxViolations    int            `db:"max_violations"`
	}
	query := `
		SELECT jp.is_locked, jp.lock_reason, jp.proctor_warning, 
		       COALESCE(jp.tab_switch_count, 0) as tab_switch_count, 
		       COALESCE(jp.risk_score, 0) as risk_score, 
		       COALESCE(j.waktupengerjaan::integer, 90) as waktupengerjaan,
		       jp.remaining_seconds, jp.tglmulai, jp.tglselesai, jp.submit_reason, j.exam_deadline_at,
		       COALESCE(j.max_violations, u.max_violations, 5) as max_violations
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_jadwalujian j ON j.idjadwalujian = jp.idjadwalujian
		LEFT JOIN cat.at_ujian u ON u.idujian = j.idujian
		WHERE jp.idjadwalujian = $1 AND jp.kodepeserta = $2
	`
	err := r.db.GetContext(ctx, &row, query, scheduleID, participantCode)
	if err != nil {
		remFallback := 3600
		if req != nil && req.ClientRemainingSeconds > 0 {
			remFallback = req.ClientRemainingSeconds
		}
		return &dto.HeartbeatResponseDTO{
			ServerTime:       now.Format("15:04:05"),
			RemainingSeconds: remFallback,
			IsLocked:         0,
			MaxViolations:    5,
		}, nil
	}

	maxViolations := row.MaxViolations
	if maxViolations <= 0 {
		maxViolations = 5
	}

	// 🛑 IMMEDIATE CHECK: If exam is already submitted/finished (e.g. by Proctor FORCE_SUBMIT or earlier completion),
	// instantly order client to terminate session and submit without letting client overwrite remaining seconds!
	if row.TglSelesai != nil {
		sReason := "COMPLETED"
		if row.SubmitReason.Valid && row.SubmitReason.String != "" {
			sReason = row.SubmitReason.String
		}
		lockMsg := "Ujian telah selesai."
		if sReason == "FORCE_SUBMIT" {
			lockMsg = "Ujian telah diselesaikan paksa oleh pengawas ruang."
		}

		_, _ = r.db.ExecContext(ctx, `
			UPDATE cat.at_jadwalpeserta 
			SET last_heartbeat = NOW(), remaining_seconds = 0, is_locked = 1 
			WHERE idjadwalujian = $1 AND kodepeserta = $2
		`, scheduleID, participantCode)

		return &dto.HeartbeatResponseDTO{
			ServerTime:       now.Format("15:04:05"),
			RemainingSeconds: 0,
			IsLocked:         1,
			LockReason:       &lockMsg,
			ShouldSubmit:     true,
			SubmitReason:     sReason,
			TabSwitchCount:   row.TabSwitchCount,
			RiskScore:        row.RiskScore,
			MaxViolations:    maxViolations,
		}, nil
	}

	// Auto-lock check: if tab_switch_count >= maxViolations or risk_score >= 81, ensure is_locked = 1
	isLocked := row.IsLocked
	lockReason := ""
	if row.LockReason.Valid {
		lockReason = row.LockReason.String
	}
	if (row.TabSwitchCount >= maxViolations || row.RiskScore >= 81) && isLocked == 0 {
		isLocked = 1
		lockReason = fmt.Sprintf("Sesi dikunci otomatis: jumlah pelanggaran mencapai batas maksimal (%dx). Hubungi pengawas untuk membuka kunci.", maxViolations)
		_, _ = r.db.ExecContext(ctx, `UPDATE cat.at_jadwalpeserta SET is_locked = 1, lock_reason = $1 WHERE idjadwalujian = $2 AND kodepeserta = $3`, lockReason, scheduleID, participantCode)
	}

	// Calculate remaining seconds
	durationSec := row.WaktuPengerjaan * 60
	remaining := durationSec
	if req != nil && req.ClientRemainingSeconds > 0 && req.ClientRemainingSeconds <= durationSec {
		remaining = req.ClientRemainingSeconds
	} else if row.RemainingSeconds != nil && *row.RemainingSeconds > 0 && *row.RemainingSeconds <= durationSec {
		remaining = *row.RemainingSeconds
	} else if row.TglMulai != nil {
		elapsed := int(now.Sub(*row.TglMulai).Seconds())
		remaining = durationSec - elapsed
		if remaining < 0 {
			remaining = 0
		}
	}

	// Calculate against absolute exam deadline if set (Hard Exam Deadline)
	if row.DeadlineAt != nil {
		secToDeadline := int(row.DeadlineAt.Sub(now).Seconds())
		if secToDeadline < remaining {
			remaining = secToDeadline
		}
	}

	shouldSubmit := false
	submitReason := ""
	if remaining <= 0 {
		remaining = 0
		shouldSubmit = true
		submitReason = "PARTICIPANT_TIMEOUT"
		// Proactively finalize in background to guarantee scores are calculated and saved
		go func(pCode string, sID int, reason string) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			_, _, _, _, _, _, _, _ = r.FinishExam(bgCtx, pCode, sID, reason)
		}(participantCode, scheduleID, submitReason)
	}

	// Persist verified remaining seconds and last_heartbeat to DB
	isPausedVal := false
	if req != nil && req.IsPaused {
		isPausedVal = true
	}
	_, _ = r.db.ExecContext(ctx, `
		UPDATE cat.at_jadwalpeserta 
		SET last_heartbeat = NOW(), remaining_seconds = $3, is_paused = $4 
		WHERE idjadwalujian = $1 AND kodepeserta = $2
	`, scheduleID, participantCode, remaining, isPausedVal)

	resp := &dto.HeartbeatResponseDTO{
		ServerTime:       now.Format("15:04:05"),
		RemainingSeconds: remaining,
		IsLocked:         isLocked,
		ShouldSubmit:     shouldSubmit,
		SubmitReason:     submitReason,
		TabSwitchCount:   row.TabSwitchCount,
		RiskScore:        row.RiskScore,
		MaxViolations:    maxViolations,
	}
	if lockReason != "" {
		resp.LockReason = &lockReason
	}
	if row.ProctorWarning.Valid {
		resp.ProctorWarning = &row.ProctorWarning.String
	}
	return resp, nil
}

func (r *examinationRepository) GetLiveMonitoring(ctx context.Context, scheduleID int) ([]*entity.LiveMonitorRow, error) {
	query := `
		SELECT jp.kodepeserta, p.nama, 1 as nourutpeserta, 
		       CASE WHEN jp.tglselesai IS NOT NULL THEN 'COMPLETED'
		            WHEN jp.tglmulai IS NOT NULL THEN 'IN_PROGRESS'
		            ELSE 'READY' END as statusujian,
		       COALESCE(
		           NULLIF(CASE WHEN jp.session_snapshot IS NOT NULL AND jsonb_typeof(jp.session_snapshot->'questions') = 'array' THEN jsonb_array_length(jp.session_snapshot->'questions') ELSE NULL END, 0),
		           NULLIF((
		               SELECT SUM(
		                   CASE 
		                       WHEN su.jumlah_kasus_dinamis > 0 THEN 
		                           COALESCE(su.jumlah_soal_statis, 0) + 
		                           COALESCE((
		                               SELECT COUNT(*)::integer 
		                               FROM cat.cat_stimulus_items si 
		                               JOIN cat.cat_bank_stimulus bs ON bs.id_stimulus = si.id_stimulus 
		                               WHERE bs.kodesoal = su.kodesoal AND (bs.softdelete = '0' OR bs.softdelete IS NULL)
		                           ), 0)
		                       ELSE su.jumlahpertanyaan 
		                   END
		               )::integer 
		               FROM cat.at_soalujian su 
		               WHERE su.idjadwalujian = jp.idjadwalujian AND (su.softdelete = '0' OR su.softdelete IS NULL)
		           ), 0),
		           NULLIF((SELECT SUM(su.jumlahpertanyaan)::integer FROM cat.at_soalujian su WHERE su.idjadwalujian = jp.idjadwalujian AND (su.softdelete = '0' OR su.softdelete IS NULL)), 0),
		           10
		       ) as total_questions,
		       COALESCE(ans.answered_count, 0) as jumlahterjawab,
		       GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) as risk_score,
		       CASE 
		           WHEN GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) >= 81 THEN 'CRITICAL'
		           WHEN GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) >= 61 THEN 'HIGH_RISK'
		           WHEN GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) >= 41 THEN 'SUSPICIOUS'
		           WHEN GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) >= 21 THEN 'ATTENTION'
		           ELSE 'NORMAL'
		       END as risk_level,
		       GREATEST(COALESCE(sec.total_tab_switches, 0), COALESCE(jp.tab_switch_count, 0)) as tab_switch_count,
		       GREATEST(COALESCE(sec.total_fullscreen_exits, 0), COALESCE(jp.fullscreen_exit_count, 0)) as fullscreen_exit_count,
		       COALESCE(jp.tab_switch_count, 0) as active_tab_switch_count,
		       COALESCE(jp.fullscreen_exit_count, 0) as active_fullscreen_exit_count,
		       COALESCE(sec.total_violations, 0) as total_violations,
		       COALESCE(sec.unlock_count, 0) as unlock_count,
		       COALESCE(jp.disconnect_count, 0) as disconnect_count,
		       COALESCE(jp.reconnect_count, 0) as reconnect_count,
		       COALESCE(jp.is_locked, 0) as is_locked,
		       jp.lock_reason,
		       jp.proctor_warning,
		       jp.last_heartbeat,
		       jp.tglmulai as waktumulai,
		       jp.tglselesai as waktuselesai,
		       jp.client_device_id,
		       COALESCE(jp.is_verified, 0) as is_verified,
		       jp.barcode_scanned_at
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_peserta p ON p.kodepeserta = jp.kodepeserta
		LEFT JOIN (
			SELECT kodepeserta, COUNT(*) as answered_count
			FROM cat.at_jawabanpeserta
			WHERE idjadwalujian = $1 AND (softdelete = '0' OR softdelete IS NULL)
			GROUP BY kodepeserta
		) ans ON ans.kodepeserta = jp.kodepeserta
		LEFT JOIN (
			SELECT idjadwal, kodepeserta,
			       COUNT(*) FILTER (
			           WHERE event_type IN (
			               'TAB_SWITCH', 'WINDOW_BLUR', 'FULLSCREEN_EXIT', 'FULLSCREEN_REQUIRED',
			               'CAMERA_DISCONNECTED', 'FORBIDDEN_SHORTCUT', 'SCREENSHOT_ATTEMPT',
			               'DEVTOOLS_ATTEMPT', 'DEVTOOLS_OPEN', 'PRINT_ATTEMPT', 'COPY_ATTEMPT',
			               'PASTE_ATTEMPT', 'CUT_ATTEMPT', 'MULTI_MONITOR', 'VM_DETECTED',
			               'UNTRUSTED_EVENT', 'MULTI_FACE_DETECTED', 'CAMERA_BLACKOUT'
			           ) AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       )::integer as total_violations,
			       COUNT(*) FILTER (
			           WHERE event_type IN ('TAB_SWITCH', 'WINDOW_BLUR') 
			             AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       )::integer as total_tab_switches,
			       COUNT(*) FILTER (
			           WHERE event_type IN ('FULLSCREEN_EXIT', 'FULLSCREEN_REQUIRED') 
			             AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       )::integer as total_fullscreen_exits,
			       COALESCE(SUM(risk_score) FILTER (
			           WHERE event_type IN (
			               'TAB_SWITCH', 'WINDOW_BLUR', 'FULLSCREEN_EXIT', 'FULLSCREEN_REQUIRED',
			               'CAMERA_DISCONNECTED', 'FORBIDDEN_SHORTCUT', 'SCREENSHOT_ATTEMPT',
			               'DEVTOOLS_ATTEMPT', 'DEVTOOLS_OPEN', 'PRINT_ATTEMPT', 'COPY_ATTEMPT',
			               'PASTE_ATTEMPT', 'CUT_ATTEMPT', 'MULTI_MONITOR', 'VM_DETECTED',
			               'UNTRUSTED_EVENT', 'MULTI_FACE_DETECTED', 'CAMERA_BLACKOUT'
			           ) AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       ), 0)::integer as cumulative_risk_score,
			       COUNT(*) FILTER (WHERE event_type = 'PROCTOR_UNLOCK')::integer as unlock_count
			FROM cat.at_security_events
			WHERE idjadwal = $1
			GROUP BY idjadwal, kodepeserta
		) sec ON sec.idjadwal = jp.idjadwalujian AND sec.kodepeserta = jp.kodepeserta
		WHERE jp.idjadwalujian = $1 AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
		ORDER BY jp.kodepeserta ASC
	`
	var rows []*entity.LiveMonitorRow
	err := r.db.SelectContext(ctx, &rows, query, scheduleID)
	return rows, err
}

func (r *examinationRepository) FinishExam(ctx context.Context, participantCode string, scheduleID int, submitReason string) (*entity.JadwalUjian, int, int, int, int, float64, string, error) {
	if submitReason == "" {
		submitReason = "MANUAL_SUBMIT"
	}

	// Idempotency check with transaction and row lock
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, 0, 0, 0, 0, 0, "", err
	}
	defer tx.Rollback()

	var existingCheck struct {
		TglSelesai      *time.Time     `db:"tglselesai"`
		Nilai           *float64       `db:"nilai"`
		SessionSnapshot sql.NullString `db:"session_snapshot"`
	}
	checkQ := `SELECT tglselesai, nilai, session_snapshot::text as session_snapshot FROM cat.at_jadwalpeserta WHERE idjadwalujian = $1 AND kodepeserta = $2`
	if err := tx.GetContext(ctx, &existingCheck, checkQ, scheduleID, participantCode); err == nil && existingCheck.TglSelesai != nil {
		// Already submitted: return idempotent score
		finalScore := 0.0
		if existingCheck.Nilai != nil {
			finalScore = *existingCheck.Nilai
		}
		status := "TL"
		if finalScore >= 60.0 {
			status = "L"
		}
		total := 10
		if existingCheck.SessionSnapshot.Valid && len(existingCheck.SessionSnapshot.String) > 10 {
			var stored StoredSessionSnapshot
			if err := json.Unmarshal([]byte(existingCheck.SessionSnapshot.String), &stored); err == nil && len(stored.Questions) > 0 {
				total = len(stored.Questions)
			}
		}
		correct := int(finalScore * float64(total) / 100.0)
		return nil, total, correct, total - correct, 0, finalScore, status, nil
	}

	var total, correct, wrong, empty int
	var finalScore float64
	var statusLulus string

	// 1. First priority: Grade against participant's frozen Session Snapshot (Exact subset & sequence)
	if existingCheck.SessionSnapshot.Valid && len(existingCheck.SessionSnapshot.String) > 10 {
		var stored StoredSessionSnapshot
		if err := json.Unmarshal([]byte(existingCheck.SessionSnapshot.String), &stored); err == nil && len(stored.Questions) > 0 {
			type ansRow struct {
				NoSoal  int `db:"nosoal"`
				Jawaban int `db:"jawaban"`
			}
			var userAnswers []ansRow
			_ = tx.SelectContext(ctx, &userAnswers, `
				SELECT nourutpeserta::integer as nosoal, COALESCE(jawabanpilih::integer, 0) as jawaban
				FROM cat.at_jawabanpeserta
				WHERE idjadwalujian = $1 AND kodepeserta = $2 AND (softdelete = '0' OR softdelete IS NULL)
			`, scheduleID, participantCode)

			ansMap := make(map[int]int)
			for _, a := range userAnswers {
				ansMap[a.NoSoal] = a.Jawaban
			}

			total = len(stored.Questions)
			totalWeight := 0.0
			earnedWeight := 0.0
			for _, q := range stored.Questions {
				userAns, answered := ansMap[q.QuestionNumber]
				wCorr := q.WeightCorrect
				if wCorr <= 0 {
					wCorr = 1.0
				}
				wWrng := q.WeightWrong
				totalWeight += wCorr

				isBenar := 0
				scoreVal := 0.0
				if !answered || userAns == 0 {
					empty++
				} else if userAns == q.CorrectOptionKey {
					correct++
					isBenar = 1
					scoreVal = wCorr
					earnedWeight += wCorr
				} else {
					wrong++
					scoreVal = wWrng
					earnedWeight += wWrng
				}

				if answered && userAns > 0 {
					_, _ = tx.ExecContext(ctx, `
						UPDATE cat.at_jawabanpeserta 
						SET isbenar = $1, nilai = $2, t_updatetime = NOW() 
						WHERE idjadwalujian = $3 AND kodepeserta = $4 AND nourutpeserta = $5
					`, isBenar, scoreVal, scheduleID, participantCode, q.QuestionNumber)
				}
			}

			if totalWeight > 0 {
				if earnedWeight < 0 {
					earnedWeight = 0
				}
				finalScore = math.Round((earnedWeight/totalWeight)*1000) / 10
			} else if total > 0 {
				finalScore = math.Round((float64(correct)/float64(total))*1000) / 10
			}
			statusLulus = "TL"
			if finalScore >= 60.0 {
				statusLulus = "L"
			}
		}
	}

	// 2. Fallback: Grade against master question bank if no session snapshot was captured
	if total == 0 {
		var bankCode string
		err = tx.GetContext(ctx, &bankCode, `SELECT COALESCE(kodesoal, 'SOAL_2026') FROM cat.at_soalujian WHERE idjadwalujian = $1 LIMIT 1`, scheduleID)
		if err != nil || bankCode == "" {
			bankCode = "SOAL_2026"
		}

		query := `
			SELECT p.nourut, a.jawabanpilih, p.jawabanbenar
			FROM cat.at_pertanyaan p
			LEFT JOIN cat.at_jawabanpeserta a ON a.kodesoal = p.kodesoal AND a.nourut = p.nourut AND a.idjadwalujian = $1 AND a.kodepeserta = $2
			WHERE p.kodesoal = $3 AND (p.softdelete = '0' OR p.softdelete IS NULL)
		`
		rows, err := tx.QueryContext(ctx, query, scheduleID, participantCode, bankCode)
		if err != nil {
			return nil, 0, 0, 0, 0, 0, "", err
		}
		defer rows.Close()

		for rows.Next() {
			total++
			var no int
			var userAns sql.NullInt64
			var key int
			if err := rows.Scan(&no, &userAns, &key); err == nil {
				if !userAns.Valid || userAns.Int64 == 0 {
					empty++
				} else if int(userAns.Int64) == key {
					correct++
				} else {
					wrong++
				}
			}
		}

		if total > 0 {
			finalScore = float64(correct) * (100.0 / float64(total))
		}
		statusLulus = "TL"
		if finalScore >= 60.0 {
			statusLulus = "L"
		}
	}

	now := time.Now()
	updateQuery := `
		UPDATE cat.at_jadwalpeserta
		SET tglselesai = $1, nilai = $2, submit_reason = $3, remaining_seconds = 0
		WHERE idjadwalujian = $4 AND kodepeserta = $5
	`
	_, _ = tx.ExecContext(ctx, updateQuery, now, finalScore, submitReason, scheduleID, participantCode)

	// Also sync to cat_participant_sessions_ext if exists
	sessExtQ := `
		UPDATE cat.cat_participant_sessions_ext
		SET session_status = 'SUBMITTED', completed_at = $1, total_score = $2,
		    total_correct = $3, total_wrong = $4, total_unanswered = $5,
		    pass_status = $6, submit_reason = $7, updated_at = NOW()
		WHERE idjadwalujian = $8 AND kodepeserta = $9
	`
	_, _ = tx.ExecContext(ctx, sessExtQ, now, finalScore, correct, wrong, empty, statusLulus, submitReason, scheduleID, participantCode)

	// Record submit audit event
	secEventQ := `
		INSERT INTO cat.at_security_events (idjadwal, kodepeserta, event_type, severity, risk_score, metadata, created_at)
		VALUES ($1, $2, $3, 'LOW', 0, $4::jsonb, NOW())
	`
	metaJson, _ := json.Marshal(map[string]interface{}{"reason": submitReason, "score": finalScore, "total": total, "correct": correct, "wrong": wrong, "empty": empty})
	_, _ = tx.ExecContext(ctx, secEventQ, scheduleID, participantCode, submitReason, string(metaJson))

	if err := tx.Commit(); err != nil {
		return nil, 0, 0, 0, 0, 0, "", err
	}

	return nil, total, correct, wrong, empty, finalScore, statusLulus, nil
}

func (r *examinationRepository) GetCollusionAnalysis(ctx context.Context, scheduleID int) (*dto.CollusionReportResponseDTO, error) {
	// 1. Fetch participants list
	type partRow struct {
		KodePeserta   string `db:"kodepeserta"`
		Nama          string `db:"nama"`
		NoUrutPeserta int    `db:"nourutpeserta"`
	}
	var participants []partRow
	partQuery := `
		SELECT jp.kodepeserta, p.nama, 1 as nourutpeserta
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_peserta p ON p.kodepeserta = jp.kodepeserta
		WHERE jp.idjadwalujian = $1 AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
		ORDER BY jp.kodepeserta ASC
	`
	_ = r.db.SelectContext(ctx, &participants, partQuery, scheduleID)

	// 2. Fetch question keys
	type qKey struct {
		NoUrut       int `db:"nourut"`
		JawabanBenar int `db:"jawabanbenar"`
	}
	var keys []qKey
	keyQuery := `
		SELECT p.nourut, p.jawabanbenar
		FROM cat.at_pertanyaan p
		JOIN cat.at_soalujian su ON su.kodesoal = p.kodesoal
		WHERE su.idjadwalujian = $1 AND (p.softdelete = '0' OR p.softdelete IS NULL)
	`
	_ = r.db.SelectContext(ctx, &keys, keyQuery, scheduleID)
	keyMap := make(map[int]int)
	for _, k := range keys {
		keyMap[k.NoUrut] = k.JawabanBenar
	}

	// 3. Fetch all answers
	type ansRow struct {
		KodePeserta  string `db:"kodepeserta"`
		NoUrut       int    `db:"nourut"`
		JawabanPilih int    `db:"jawabanpilih"`
	}
	var answers []ansRow
	ansQuery := `
		SELECT kodepeserta, nourut, jawabanpilih
		FROM cat.at_jawabanpeserta
		WHERE idjadwalujian = $1 AND (softdelete = '0' OR softdelete IS NULL)
	`
	_ = r.db.SelectContext(ctx, &answers, ansQuery, scheduleID)

	userAnswers := make(map[string]map[int]int)
	for _, a := range answers {
		if userAnswers[a.KodePeserta] == nil {
			userAnswers[a.KodePeserta] = make(map[int]int)
		}
		userAnswers[a.KodePeserta][a.NoUrut] = a.JawabanPilih
	}

	// 4. Compare all pairs
	var pairs []dto.CollusionPairDTO
	suspiciousCount := 0

	for i := 0; i < len(participants); i++ {
		for j := i + 1; j < len(participants); j++ {
			p1 := participants[i]
			p2 := participants[j]

			ans1 := userAnswers[p1.KodePeserta]
			ans2 := userAnswers[p2.KodePeserta]

			if len(ans1) == 0 || len(ans2) == 0 {
				continue
			}

			identical := 0
			totalCompared := 0
			identicalWrong := 0
			totalWrong := 0

			for qNo, opt1 := range ans1 {
				if opt2, exists := ans2[qNo]; exists && opt1 > 0 && opt2 > 0 {
					totalCompared++
					correctOpt := keyMap[qNo]

					if opt1 == opt2 {
						identical++
					}

					if opt1 != correctOpt && opt2 != correctOpt {
						totalWrong++
						if opt1 == opt2 {
							identicalWrong++
						}
					}
				}
			}

			if totalCompared > 0 {
				simPercent := (float64(identical) / float64(totalCompared)) * 100.0
				wrongSimPercent := 0.0
				if totalWrong > 0 {
					wrongSimPercent = (float64(identicalWrong) / float64(totalWrong)) * 100.0
				}

				riskStatus := "NORMAL"
				if simPercent >= 80.0 || wrongSimPercent >= 75.0 {
					riskStatus = "HIGH_COLLUSION"
					suspiciousCount++
				} else if simPercent >= 65.0 || wrongSimPercent >= 50.0 {
					riskStatus = "SUSPICIOUS"
					suspiciousCount++
				}

				pairs = append(pairs, dto.CollusionPairDTO{
					ParticipantCode1:         p1.KodePeserta,
					Name1:                    p1.Nama,
					DeskNumber1:              p1.NoUrutPeserta,
					ParticipantCode2:         p2.KodePeserta,
					Name2:                    p2.Nama,
					DeskNumber2:              p2.NoUrutPeserta,
					OverallSimilarityPercent: simPercent,
					WrongSimilarityPercent:   wrongSimPercent,
					IdenticalCount:           identical,
					TotalCompared:            totalCompared,
					IdenticalWrongCount:      identicalWrong,
					RiskStatus:               riskStatus,
				})
			}
		}
	}

	return &dto.CollusionReportResponseDTO{
		ScheduleID:           scheduleID,
		TotalParticipants:    len(participants),
		TotalPairsAnalyzed:   len(pairs),
		SuspiciousPairsCount: suspiciousCount,
		Pairs:                pairs,
	}, nil
}

// ============================================================================
// DYNAMIC MULTIMEDIA & MULTI-ITEM EXAMINATION METHODS (ASIA/JAKARTA TIMEZONE)
// ============================================================================

func (r *examinationRepository) GetDynamicExamSession(ctx context.Context, participantCode string, scheduleID int) (*dto.DynamicExamSessionResponseDTO, error) {
	locJakarta, _ := time.LoadLocation("Asia/Jakarta")
	if locJakarta == nil {
		locJakarta = time.FixedZone("WIB", 7*3600)
	}
	now := time.Now().In(locJakarta)

	// 1. Fetch Schedule and Exam Details
	var sched entity.JadwalUjian
	querySched := `
		SELECT j.idjadwalujian, j.idujian, u.namaujian,
		       COALESCE((SELECT su.kodesoal FROM cat.at_soalujian su WHERE su.idjadwalujian = j.idjadwalujian LIMIT 1), 'SOAL_2026') as kodesoal, 
		       COALESCE(r.idruangujian, 1) as idruang, 
		       COALESCE(r.koderuang, 'Lab 1') as namaruang, 
		       COALESCE(r.tglmulai, NOW())::date as tglujian, 
		       COALESCE(r.waktumulai, '08:00') as jammulai, 
		       '15:00' as jamselesai, 
		       COALESCE(j.waktupengerjaan::integer, 90) as waktupengerjaan, 
		       COALESCE(j.token_ujian, '') as tokenujian, 
		       COALESCE(r.jumlahpeserta::integer, 40) as kapasitas
		FROM cat.at_jadwalujian j
		JOIN cat.at_ujian u ON u.idujian = j.idujian
		LEFT JOIN cat.at_ruangujian r ON r.idjadwalujian = j.idjadwalujian
		WHERE j.idjadwalujian = $1 AND (j.softdelete = '0' OR j.softdelete IS NULL)
	`
	err := r.db.GetContext(ctx, &sched, querySched, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("jadwal ujian tidak ditemukan: %w", err)
	}

	// 2. Fetch or initialize Exam Schedule Extension (Time Window & Scoring Rules)
	var ext struct {
		WindowStart           time.Time `db:"window_start_time"`
		WindowEnd             time.Time `db:"window_end_time"`
		ScoringRule           string    `db:"scoring_rule"`
		DefaultCorrectScore   float64   `db:"default_correct_score"`
		DefaultWrongScore     float64   `db:"default_wrong_score"`
		AutoSubmitOnWindowEnd bool      `db:"auto_submit_on_window_end"`
		Timezone              string    `db:"timezone"`
	}
	queryExt := `
		SELECT window_start_time, window_end_time, scoring_rule, default_correct_score,
		       default_wrong_score, auto_submit_on_window_end, timezone
		FROM cat.cat_exam_schedules_ext
		WHERE idjadwalujian = $1
	`
	err = r.db.GetContext(ctx, &ext, queryExt, scheduleID)
	if err != nil {
		// Initialize default 08:00 - 15:00 window if not yet set
		windowStart := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, locJakarta)
		windowEnd := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, locJakarta)
		if now.Hour() >= 15 {
			windowEnd = now.Add(2 * time.Hour)
		}
		ext.WindowStart = windowStart
		ext.WindowEnd = windowEnd
		ext.ScoringRule = "STANDARD"
		ext.DefaultCorrectScore = 1.0
		ext.DefaultWrongScore = 0.0
		ext.AutoSubmitOnWindowEnd = true
		ext.Timezone = "Asia/Jakarta"

		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO cat.cat_exam_schedules_ext (idjadwalujian, window_start_time, window_end_time, scoring_rule, default_correct_score, default_wrong_score, auto_submit_on_window_end, timezone)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (idjadwalujian) DO NOTHING
		`, scheduleID, windowStart, windowEnd, ext.ScoringRule, ext.DefaultCorrectScore, ext.DefaultWrongScore, ext.AutoSubmitOnWindowEnd, ext.Timezone)
	}

	// 3. Time Window Validation
	if now.Before(ext.WindowStart) {
		return nil, fmt.Errorf("sesi ujian belum dibuka. Jadwal ujian baru dapat diakses mulai pukul %s WIB", ext.WindowStart.In(locJakarta).Format("15:04"))
	}
	if now.After(ext.WindowEnd) {
		return nil, fmt.Errorf("jendela waktu ujian telah berakhir pada pukul %s WIB", ext.WindowEnd.In(locJakarta).Format("15:04"))
	}

	// 4. Check Participant Session State
	var sess entity.ParticipantSessionExt
	sessQuery := `
		SELECT id_session, idjadwalujian, kodepeserta, session_status, started_at, completed_at,
		       calculated_deadline, session_snapshot, total_score, total_correct, total_wrong, total_unanswered, pass_status
		FROM cat.cat_participant_sessions_ext
		WHERE idjadwalujian = $1 AND kodepeserta = $2
	`
	err = r.db.GetContext(ctx, &sess, sessQuery, scheduleID, participantCode)

	// If already submitted -> Lock permanently
	if err == nil && (sess.SessionStatus == "SUBMITTED" || sess.CompletedAt != nil) {
		return nil, fmt.Errorf("Anda telah menyelesaikan ujian ini pada %s. Lembar ujian telah dikunci permanen", sess.CompletedAt.In(locJakarta).Format("02 Jan 2006 15:04 WIB"))
	}
	if err == nil && sess.SessionStatus == "EXPIRED_AUTO_SUBMIT" {
		return nil, fmt.Errorf("waktu ujian Anda telah habis dan telah di-submit otomatis oleh sistem")
	}

	durationSec := sched.WaktuPengerjaan * 60
	var startedAt time.Time
	var calculatedDeadline time.Time

	// Calculate Effective Deadline (min of started_at + duration, and window_end_time)
	if err == nil && sess.StartedAt != nil {
		startedAt = *sess.StartedAt
		if sess.CalculatedDeadline != nil {
			calculatedDeadline = *sess.CalculatedDeadline
		} else {
			durationDeadline := startedAt.Add(time.Duration(durationSec) * time.Second)
			if durationDeadline.After(ext.WindowEnd) {
				calculatedDeadline = ext.WindowEnd
			} else {
				calculatedDeadline = durationDeadline
			}
		}
	} else {
		// New Session Started Now
		startedAt = now
		durationDeadline := startedAt.Add(time.Duration(durationSec) * time.Second)
		if durationDeadline.After(ext.WindowEnd) {
			calculatedDeadline = ext.WindowEnd
		} else {
			calculatedDeadline = durationDeadline
		}
	}

	remainingSec := int(calculatedDeadline.Sub(now).Seconds())
	if remainingSec <= 0 {
		// Auto-submit immediately if expired
		_, _ = r.FinishDynamicExam(ctx, participantCode, scheduleID, "TIMEOUT_AUTO_SUBMIT")
		return nil, fmt.Errorf("waktu pengerjaan ujian telah berakhir")
	}

	// 5. Load or Build Dynamic Snapshot
	var stimuliList []dto.DynamicStimulusUnitDTO

	if err == nil && len(sess.SessionSnapshot) > 20 {
		_ = json.Unmarshal(sess.SessionSnapshot, &stimuliList)
	}

	if len(stimuliList) == 0 {
		// Fetch stimuli from cat_bank_stimulus for this question bank code
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
		var rawList []rawStimulus
		_ = r.db.SelectContext(ctx, &rawList, `
			SELECT id_stimulus, stimulus_code, title, narrative_text, media_type, media_url,
			       media_metadata, category_tag, difficulty_level
			FROM cat.cat_bank_stimulus
			WHERE kodesoal = $1 AND (softdelete = '0' OR softdelete IS NULL)
			ORDER BY sort_order ASC, id_stimulus ASC
		`, sched.KodeSoal)

		stimulusNumber := 1
		for _, raw := range rawList {
			// Fetch child items
			type rawItem struct {
				IDItem            int64   `db:"id_item"`
				ItemOrder         int     `db:"item_order"`
				QuestionText      string  `db:"question_text"`
				QuestionMediaType string  `db:"question_media_type"`
				QuestionMediaURL  *string `db:"question_media_url"`
				ItemType          string  `db:"item_type"`
				WeightCorrect     float64 `db:"weight_correct"`
				WeightWrong       float64 `db:"weight_wrong"`
			}
			var items []rawItem
			_ = r.db.SelectContext(ctx, &items, `
				SELECT id_item, item_order, question_text, question_media_type, question_media_url,
				       item_type, weight_correct, weight_wrong
				FROM cat.cat_stimulus_items
				WHERE id_stimulus = $1 AND (softdelete = '0' OR softdelete IS NULL)
				ORDER BY item_order ASC, id_item ASC
			`, raw.IDStimulus)

			var subItems []dto.DynamicSubItemDTO
			for _, it := range items {
				// Fetch options
				type rawOption struct {
					IDOption    int64   `db:"id_option"`
					OptionLabel string  `db:"option_label"`
					OptionText  string  `db:"option_text"`
					MediaType   string  `db:"media_type"`
					MediaURL    *string `db:"media_url"`
				}
				var opts []rawOption
				_ = r.db.SelectContext(ctx, &opts, `
					SELECT id_option, option_label, option_text, media_type, media_url
					FROM cat.cat_item_options
					WHERE id_item = $1
					ORDER BY sort_order ASC, option_label ASC
				`, it.IDItem)

				var optDTOs []dto.DynamicItemOptionDTO
				for _, op := range opts {
					optDTOs = append(optDTOs, dto.DynamicItemOptionDTO{
						IDOption:    op.IDOption,
						OptionLabel: op.OptionLabel,
						OptionText:  op.OptionText,
						MediaType:   op.MediaType,
						MediaURL:    op.MediaURL,
					})
				}

				subItems = append(subItems, dto.DynamicSubItemDTO{
					IDItem:            it.IDItem,
					ItemOrder:         it.ItemOrder,
					QuestionText:      it.QuestionText,
					QuestionMediaType: it.QuestionMediaType,
					QuestionMediaURL:  it.QuestionMediaURL,
					ItemType:          it.ItemType,
					WeightCorrect:     it.WeightCorrect,
					WeightWrong:       it.WeightWrong,
					Options:           optDTOs,
				})
			}

			if len(subItems) > 0 {
				var metaMap map[string]any
				_ = json.Unmarshal(raw.MediaMetadata, &metaMap)

				stimuliList = append(stimuliList, dto.DynamicStimulusUnitDTO{
					IDStimulus:     raw.IDStimulus,
					StimulusNumber: stimulusNumber,
					StimulusCode:   raw.StimulusCode,
					Title:          raw.Title,
					NarrativeText:  raw.NarrativeText,
					MediaType:      raw.MediaType,
					MediaURL:       raw.MediaURL,
					MediaMetadata:  metaMap,
					CategoryTag:    raw.CategoryTag,
					SubItems:       subItems,
				})
				stimulusNumber++
			}
		}

		// Save frozen snapshot
		snapshotBytes, _ := json.Marshal(stimuliList)
		upsertSessionQ := `
			INSERT INTO cat.cat_participant_sessions_ext (
				idjadwalujian, kodepeserta, session_status, started_at, calculated_deadline, session_snapshot, updated_at
			) VALUES ($1, $2, 'IN_PROGRESS', $3, $4, $5, NOW())
			ON CONFLICT (idjadwalujian, kodepeserta)
			DO UPDATE SET
				session_status = 'IN_PROGRESS',
				started_at = COALESCE(cat.cat_participant_sessions_ext.started_at, EXCLUDED.started_at),
				calculated_deadline = COALESCE(cat.cat_participant_sessions_ext.calculated_deadline, EXCLUDED.calculated_deadline),
				session_snapshot = COALESCE(cat.cat_participant_sessions_ext.session_snapshot, EXCLUDED.session_snapshot),
				updated_at = NOW()
		`
		_, _ = r.db.ExecContext(ctx, upsertSessionQ, scheduleID, participantCode, startedAt, calculatedDeadline, snapshotBytes)
	}

	// 6. Attach existing participant answers to sub-items
	type ansRow struct {
		IDItem          int64  `db:"id_item"`
		SelectedOptions string `db:"selected_options"`
		IsDoubtful      bool   `db:"is_doubtful"`
	}
	var existingAns []ansRow
	_ = r.db.SelectContext(ctx, &existingAns, `
		SELECT a.id_item, a.selected_options, a.is_doubtful
		FROM cat.cat_participant_answers_multi a
		JOIN cat.cat_participant_sessions_ext s ON s.id_session = a.id_session
		WHERE s.idjadwalujian = $1 AND s.kodepeserta = $2
	`, scheduleID, participantCode)

	ansMap := make(map[int64]ansRow)
	for _, a := range existingAns {
		ansMap[a.IDItem] = a
	}

	totalSubItems := 0
	for i := range stimuliList {
		for j := range stimuliList[i].SubItems {
			totalSubItems++
			itID := stimuliList[i].SubItems[j].IDItem
			if val, ok := ansMap[itID]; ok {
				sel := val.SelectedOptions
				stimuliList[i].SubItems[j].UserSelectedOption = &sel
				stimuliList[i].SubItems[j].IsDoubtful = val.IsDoubtful
			}
		}
	}

	return &dto.DynamicExamSessionResponseDTO{
		ScheduleID:       scheduleID,
		ExamName:         sched.NamaUjian,
		IsDynamic:        true,
		TimeWindowStart:  ext.WindowStart.In(locJakarta).Format(time.RFC3339),
		TimeWindowEnd:    ext.WindowEnd.In(locJakarta).Format(time.RFC3339),
		DurationMinutes:  sched.WaktuPengerjaan,
		RemainingSeconds: remainingSec,
		SessionStatus:    "IN_PROGRESS",
		ScoringRule:      ext.ScoringRule,
		ServerTimeWIB:    now.Format("15:04:05 WIB"),
		TotalStimuli:     len(stimuliList),
		TotalSubItems:    totalSubItems,
		Stimuli:          stimuliList,
	}, nil
}

func (r *examinationRepository) SaveDynamicAnswer(ctx context.Context, req *dto.SaveMultiAnswerDTO) error {
	locJakarta, _ := time.LoadLocation("Asia/Jakarta")
	if locJakarta == nil {
		locJakarta = time.FixedZone("WIB", 7*3600)
	}
	now := time.Now().In(locJakarta)

	// Validate Session
	var sess entity.ParticipantSessionExt
	err := r.db.GetContext(ctx, &sess, `
		SELECT id_session, session_status, calculated_deadline
		FROM cat.cat_participant_sessions_ext
		WHERE idjadwalujian = $1 AND kodepeserta = $2
	`, req.ScheduleID, req.ParticipantCode)
	if err != nil {
		return fmt.Errorf("sesi ujian tidak ditemukan: %w", err)
	}

	if sess.SessionStatus == "SUBMITTED" || sess.SessionStatus == "EXPIRED_AUTO_SUBMIT" {
		return fmt.Errorf("lembar ujian telah selesai dan dikunci, jawaban tidak dapat diubah")
	}

	if sess.CalculatedDeadline != nil && now.After(*sess.CalculatedDeadline) {
		_, _ = r.FinishDynamicExam(ctx, req.ParticipantCode, req.ScheduleID, "TIMEOUT_AUTO_SUBMIT")
		return fmt.Errorf("waktu ujian telah habis (timeout)")
	}

	// Fetch correct answer and weights
	type itemInfo struct {
		CorrectAnswer string  `db:"correct_answer"`
		WeightCorrect float64 `db:"weight_correct"`
		WeightWrong   float64 `db:"weight_wrong"`
	}
	var it itemInfo
	_ = r.db.GetContext(ctx, &it, `SELECT correct_answer, weight_correct, weight_wrong FROM cat.cat_stimulus_items WHERE id_item = $1`, req.IDItem)

	isCorrect := (req.SelectedOption != "" && req.SelectedOption == it.CorrectAnswer)
	awardedPoints := 0.0
	if req.SelectedOption != "" {
		if isCorrect {
			awardedPoints = it.WeightCorrect
			if awardedPoints == 0 {
				awardedPoints = 1.0
			}
		} else {
			awardedPoints = it.WeightWrong
		}
	}

	upsertAnswerQ := `
		INSERT INTO cat.cat_participant_answers_multi (
			id_session, id_stimulus, id_item, selected_options, is_doubtful, is_correct, awarded_points, answered_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (id_session, id_item)
		DO UPDATE SET
			selected_options = EXCLUDED.selected_options,
			is_doubtful = EXCLUDED.is_doubtful,
			is_correct = EXCLUDED.is_correct,
			awarded_points = EXCLUDED.awarded_points,
			answered_at = NOW()
	`
	_, err = r.db.ExecContext(ctx, upsertAnswerQ,
		sess.IDSession, req.IDStimulus, req.IDItem, req.SelectedOption, req.IsDoubtful, isCorrect, awardedPoints,
	)
	return err
}

func (r *examinationRepository) FinishDynamicExam(ctx context.Context, participantCode string, scheduleID int, submitReason string) (*dto.FinishExamResponseDTO, error) {
	if submitReason == "" {
		submitReason = "MANUAL_SUBMIT"
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var sess entity.ParticipantSessionExt
	err = tx.GetContext(ctx, &sess, `
		SELECT id_session, session_status, total_score, total_correct, total_wrong, total_unanswered, pass_status, completed_at, session_snapshot
		FROM cat.cat_participant_sessions_ext
		WHERE idjadwalujian = $1 AND kodepeserta = $2
		FOR UPDATE
	`, scheduleID, participantCode)
	if err != nil {
		return nil, fmt.Errorf("sesi ujian tidak ditemukan: %w", err)
	}

	if sess.SessionStatus == "SUBMITTED" && sess.CompletedAt != nil {
		// Idempotent return
		return &dto.FinishExamResponseDTO{
			ParticipantCode: participantCode,
			ScheduleID:      scheduleID,
			TotalQuestions:  sess.TotalCorrect + sess.TotalWrong + sess.TotalUnanswered,
			CorrectCount:    sess.TotalCorrect,
			WrongCount:      sess.TotalWrong,
			EmptyCount:      sess.TotalUnanswered,
			FinalScore:      sess.TotalScore,
			PassingStatus:   sess.PassStatus,
			SubmittedAt:     sess.CompletedAt.Format(time.RFC3339),
			SubmitReason:    submitReason,
		}, nil
	}

	// Fetch scoring rule
	var rule string
	_ = tx.GetContext(ctx, &rule, `SELECT scoring_rule FROM cat.cat_exam_schedules_ext WHERE idjadwalujian = $1`, scheduleID)
	if rule == "" {
		rule = "STANDARD"
	}

	// Calculate scores across all answered sub-items
	type scoreSummary struct {
		TotalAwarded float64 `db:"total_awarded"`
		TotalCorrect int     `db:"total_correct"`
		TotalWrong   int     `db:"total_wrong"`
	}
	var sum scoreSummary
	_ = tx.GetContext(ctx, &sum, `
		SELECT COALESCE(SUM(awarded_points), 0.0) as total_awarded,
		       COALESCE(COUNT(CASE WHEN is_correct = true THEN 1 END), 0) as total_correct,
		       COALESCE(COUNT(CASE WHEN is_correct = false AND selected_options IS NOT NULL AND selected_options != '' THEN 1 END), 0) as total_wrong
		FROM cat.cat_participant_answers_multi
		WHERE id_session = $1
	`, sess.IDSession)

	// Count total items from snapshot
	var stimuliList []dto.DynamicStimulusUnitDTO
	_ = json.Unmarshal(sess.SessionSnapshot, &stimuliList)
	totalItems := 0
	maxPossibleScore := 0.0
	for _, s := range stimuliList {
		for _, it := range s.SubItems {
			totalItems++
			w := it.WeightCorrect
			if w == 0 {
				w = 1.0
			}
			maxPossibleScore += w
		}
	}
	if totalItems == 0 {
		totalItems = 50
		maxPossibleScore = 50.0
	}

	totalEmpty := totalItems - (sum.TotalCorrect + sum.TotalWrong)
	if totalEmpty < 0 {
		totalEmpty = 0
	}

	finalScore := 0.0
	if maxPossibleScore > 0 {
		finalScore = (sum.TotalAwarded / maxPossibleScore) * 100.0
	}
	if finalScore < 0 {
		finalScore = 0.0
	}

	passStatus := "TL"
	if finalScore >= 60.0 {
		passStatus = "L"
	}

	now := time.Now()
	finalStatus := "SUBMITTED"
	if submitReason == "TIMEOUT_AUTO_SUBMIT" || submitReason == "EXPIRED_AUTO_SUBMIT" {
		finalStatus = "EXPIRED_AUTO_SUBMIT"
	}

	// Update session
	updateSessQ := `
		UPDATE cat.cat_participant_sessions_ext
		SET session_status = $1, completed_at = $2, total_score = $3,
		    total_correct = $4, total_wrong = $5, total_unanswered = $6,
		    pass_status = $7, submit_reason = $8, updated_at = NOW()
		WHERE id_session = $9
	`
	_, err = tx.ExecContext(ctx, updateSessQ,
		finalStatus, now, finalScore, sum.TotalCorrect, sum.TotalWrong, totalEmpty, passStatus, submitReason, sess.IDSession,
	)
	if err != nil {
		return nil, err
	}

	// Also update legacy table cat.at_jadwalpeserta for backward reporting compatibility
	_, _ = tx.ExecContext(ctx, `
		UPDATE cat.at_jadwalpeserta
		SET tglselesai = $1, nilai = $2, submit_reason = $3
		WHERE idjadwalujian = $4 AND kodepeserta = $5
	`, now, finalScore, submitReason, scheduleID, participantCode)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &dto.FinishExamResponseDTO{
		ParticipantCode: participantCode,
		ScheduleID:      scheduleID,
		TotalQuestions:  totalItems,
		CorrectCount:    sum.TotalCorrect,
		WrongCount:      sum.TotalWrong,
		EmptyCount:      totalEmpty,
		FinalScore:      finalScore,
		PassingStatus:   passStatus,
		SubmittedAt:     now.Format(time.RFC3339),
		SubmitReason:    submitReason,
	}, nil
}

func (r *examinationRepository) GetExamParticipants(ctx context.Context, examID int) ([]*entity.ExamParticipant, error) {
	query := `
		SELECT 
			p.kodepeserta,
			p.nama,
			p.email,
			w.namawilayah,
			pu.nilai,
			COALESCE(p.islogin, 0) as isloggedin,
			CASE WHEN jp.idjadwalujian IS NOT NULL THEN 1 ELSE 0 END as is_plotted,
			ju.nojadwal as session_number,
			COALESCE(r.namaruang, ru.koderuang, '') as room_name,
			CAST(ru.tglmulai AS varchar) as tgl_mulai,
			ru.waktumulai as waktu_mulai,
			ru.waktuselesai as waktu_selesai,
			COALESCE(ju.waktupengerjaan, '0060') as waktupengerjaan_str,
			COALESCE(jp.is_verified, 0) as is_verified,
			jp.barcode_scanned_at,
			jp.scanned_by
		FROM cat.at_pesertaujian pu
		JOIN cat.at_peserta p ON p.kodepeserta = pu.kodepeserta
		LEFT JOIN cat.ms_wilayah w ON (w.idwilayah = CAST(p.idkota AS TEXT) OR w.idwilayah = p.idkota_lama)
		LEFT JOIN cat.at_jadwalpeserta jp ON jp.kodepeserta = pu.kodepeserta 
		     AND jp.idjadwalujian IN (SELECT idjadwalujian FROM cat.at_jadwalujian WHERE idujian = pu.idujian AND (softdelete = '0' OR softdelete IS NULL))
		     AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
		LEFT JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian
		LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian
		LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
		WHERE pu.idujian = $1 AND (pu.softdelete = '0' OR pu.softdelete IS NULL)
		ORDER BY p.kodepeserta ASC
	`
	var participants []*entity.ExamParticipant
	err := r.db.SelectContext(ctx, &participants, query, examID)
	if err != nil {
		return nil, err
	}

	if len(participants) == 0 {
		fallbackQuery := `
			SELECT DISTINCT
				p.kodepeserta,
				p.nama,
				p.email,
				w.namawilayah,
				jp.nilai,
				COALESCE(p.islogin, 0) as isloggedin,
				1 as is_plotted,
				ju.nojadwal as session_number,
				COALESCE(r.namaruang, ru.koderuang, '') as room_name,
				CAST(ru.tglmulai AS varchar) as tgl_mulai,
				ru.waktumulai as waktu_mulai,
				ru.waktuselesai as waktu_selesai,
				COALESCE(ju.waktupengerjaan, '0060') as waktupengerjaan_str,
				COALESCE(jp.is_verified, 0) as is_verified,
				jp.barcode_scanned_at,
				jp.scanned_by
			FROM cat.at_jadwalpeserta jp
			JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian
			JOIN cat.at_peserta p ON p.kodepeserta = jp.kodepeserta
			LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian
			LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
			LEFT JOIN cat.ms_wilayah w ON (w.idwilayah = CAST(p.idkota AS TEXT) OR w.idwilayah = p.idkota_lama)
			WHERE ju.idujian = $1 AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
			ORDER BY p.kodepeserta ASC
		`
		_ = r.db.SelectContext(ctx, &participants, fallbackQuery, examID)
	}

	if participants == nil {
		participants = []*entity.ExamParticipant{}
	}

	return participants, nil
}

// =====================================================================
// Admin CRUD: Jadwal Sesi (at_jadwalujian)
// =====================================================================

func (r *examinationRepository) GetSessionsByExam(ctx context.Context, examID int) ([]*entity.JadwalSesi, error) {
	query := `
		SELECT
			j.idjadwalujian,
			j.idujian,
			COALESCE(j.nojadwal, 0)::int AS nojadwal,
			COALESCE(j.waktupengerjaan, '0060') AS waktupengerjaan,
			COALESCE(j.bobot, 0) AS bobot,
			COALESCE(j.tampilkannilai, 0)::int AS tampilkannilai,
			CAST(j.exam_deadline_at AS varchar) AS exam_deadline_at,
			COALESCE(j.grace_period_seconds, 60) AS grace_period_seconds,
			(SELECT COUNT(*) FROM cat.at_ruangujian ru WHERE ru.idjadwalujian = j.idjadwalujian AND (ru.softdelete = '0' OR ru.softdelete IS NULL))::int AS room_count,
			COALESCE(
				NULLIF(j.waktumulai, ''),
				(SELECT MIN(ru.waktumulai) FROM cat.at_ruangujian ru WHERE ru.idjadwalujian = j.idjadwalujian AND (ru.softdelete = '0' OR ru.softdelete IS NULL))
			) AS waktu_mulai,
			COALESCE(
				NULLIF(j.waktuselesai, ''),
				(SELECT MAX(COALESCE(ru.waktuselesai, '')) FROM cat.at_ruangujian ru WHERE ru.idjadwalujian = j.idjadwalujian AND (ru.softdelete = '0' OR ru.softdelete IS NULL))
			) AS waktu_selesai,
			COALESCE(
				CAST(j.tglmulai AS varchar),
				(SELECT CAST(MIN(ru.tglmulai) AS varchar) FROM cat.at_ruangujian ru WHERE ru.idjadwalujian = j.idjadwalujian AND (ru.softdelete = '0' OR ru.softdelete IS NULL))
			) AS tgl_mulai,
			COALESCE(
				CAST(j.tglselesai AS varchar),
				(SELECT CAST(MAX(ru.tglselesai) AS varchar) FROM cat.at_ruangujian ru WHERE ru.idjadwalujian = j.idjadwalujian AND (ru.softdelete = '0' OR ru.softdelete IS NULL))
			) AS tgl_selesai,
			COALESCE(j.token_ujian, '') AS token_ujian,
			COALESCE(j.max_violations, u.max_violations, 5)::int AS max_violations
		FROM cat.at_jadwalujian j
		LEFT JOIN cat.at_ujian u ON u.idujian = j.idujian
		WHERE j.idujian = $1
		  AND (j.softdelete = '0' OR j.softdelete IS NULL)
		ORDER BY j.nojadwal ASC
	`
	var sessions []*entity.JadwalSesi
	err := r.db.SelectContext(ctx, &sessions, query, examID)
	if sessions == nil {
		sessions = []*entity.JadwalSesi{}
	}
	return sessions, err
}

func (r *examinationRepository) CreateSession(ctx context.Context, req *dto.CreateScheduleRequestDTO) (int, error) {
	// Auto-increment nojadwal
	var nextNo int
	err := r.db.GetContext(ctx, &nextNo,
		`SELECT COALESCE(MAX(nojadwal), 0) + 1 FROM cat.at_jadwalujian WHERE idujian = $1 AND (softdelete = '0' OR softdelete IS NULL)`,
		req.ExamID)
	if err != nil {
		nextNo = 1
	}
	if req.NoJadwal > 0 {
		nextNo = req.NoJadwal
	}

	grace := req.GracePeriodSeconds
	if grace == 0 {
		grace = 60
	}

	maxViolations := req.MaxViolations
	if maxViolations <= 0 {
		var examMax int
		_ = r.db.GetContext(ctx, &examMax, `SELECT COALESCE(max_violations, 5) FROM cat.at_ujian WHERE idujian = $1`, req.ExamID)
		if examMax <= 0 {
			examMax = 5
		}
		maxViolations = examMax
	}

	// waktupengerjaan stored as char(4) e.g. "0060" = 60 minutes
	waktuStr := fmt.Sprintf("%04d", req.WaktuPengerjaan)

	var deadlineVal interface{}
	if req.ExamDeadlineAt != nil && *req.ExamDeadlineAt != "" {
		deadlineVal = *req.ExamDeadlineAt
	}

	var tglMulaiVal, tglSelesaiVal interface{}
	tglMulai := sanitizeDateStr(req.TglMulai)
	if tglMulai != "" {
		tglMulaiVal = tglMulai
	}
	tglSelesai := sanitizeDateStr(req.TglSelesai)
	if tglSelesai == "" && tglMulai != "" {
		tglSelesai = tglMulai
	}
	if tglSelesai != "" {
		tglSelesaiVal = tglSelesai
	}

	waktuMulai := req.WaktuMulai
	if len(waktuMulai) == 5 && waktuMulai[2] == ':' {
		waktuMulai = waktuMulai[:2] + waktuMulai[3:]
	}
	waktuSelesai := req.WaktuSelesai
	if len(waktuSelesai) == 5 && waktuSelesai[2] == ':' {
		waktuSelesai = waktuSelesai[:2] + waktuSelesai[3:]
	}
	var waktuMulaiVal, waktuSelesaiVal interface{}
	if waktuMulai != "" {
		waktuMulaiVal = waktuMulai
	}
	if waktuSelesai != "" {
		waktuSelesaiVal = waktuSelesai
	}

	if deadlineVal == nil && tglSelesai != "" && waktuSelesai != "" {
		formattedTime := waktuSelesai
		if len(formattedTime) == 4 {
			formattedTime = formattedTime[:2] + ":" + formattedTime[2:] + ":00"
		}
		deadlineVal = tglSelesai + " " + formattedTime
	}

	newToken := utils.GenerateRandomExamToken(6)
	for attempts := 0; attempts < 10; attempts++ {
		var exists int
		_ = r.db.GetContext(ctx, &exists, `SELECT 1 FROM cat.at_jadwalujian WHERE token_ujian = $1 AND (softdelete = '0' OR softdelete IS NULL) LIMIT 1`, newToken)
		if exists == 0 {
			break
		}
		newToken = utils.GenerateRandomExamToken(6)
	}

	var newID int
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO cat.at_jadwalujian
			(idujian, nojadwal, waktupengerjaan, bobot, tampilkannilai, exam_deadline_at, grace_period_seconds, tglmulai, tglselesai, waktumulai, waktuselesai, token_ujian, max_violations, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, '0', NOW(), 'i-jadwal')
		RETURNING idjadwalujian
	`, req.ExamID, nextNo, waktuStr, req.Bobot, req.TampilkanNilai, deadlineVal, grace, tglMulaiVal, tglSelesaiVal, waktuMulaiVal, waktuSelesaiVal, newToken, maxViolations).Scan(&newID)
	return newID, err
}

func (r *examinationRepository) UpdateSession(ctx context.Context, sessionID int, req *dto.UpdateSessionRequestDTO) error {
	grace := req.GracePeriodSeconds
	if grace == 0 {
		grace = 60
	}

	maxViolations := req.MaxViolations
	if maxViolations <= 0 {
		maxViolations = 5
	}

	waktuStr := fmt.Sprintf("%04d", req.WaktuPengerjaan)

	var deadlineVal interface{}
	if req.ExamDeadlineAt != nil && *req.ExamDeadlineAt != "" {
		deadlineVal = *req.ExamDeadlineAt
	}

	var tglMulaiVal, tglSelesaiVal interface{}
	tglMulai := sanitizeDateStr(req.TglMulai)
	if tglMulai != "" {
		tglMulaiVal = tglMulai
	}
	tglSelesai := sanitizeDateStr(req.TglSelesai)
	if tglSelesai == "" && tglMulai != "" {
		tglSelesai = tglMulai
	}
	if tglSelesai != "" {
		tglSelesaiVal = tglSelesai
	}

	waktuMulai := req.WaktuMulai
	if len(waktuMulai) == 5 && waktuMulai[2] == ':' {
		waktuMulai = waktuMulai[:2] + waktuMulai[3:]
	}
	waktuSelesai := req.WaktuSelesai
	if len(waktuSelesai) == 5 && waktuSelesai[2] == ':' {
		waktuSelesai = waktuSelesai[:2] + waktuSelesai[3:]
	}
	var waktuMulaiVal, waktuSelesaiVal interface{}
	if waktuMulai != "" {
		waktuMulaiVal = waktuMulai
	}
	if waktuSelesai != "" {
		waktuSelesaiVal = waktuSelesai
	}

	if deadlineVal == nil && tglSelesai != "" && waktuSelesai != "" {
		formattedTime := waktuSelesai
		if len(formattedTime) == 4 {
			formattedTime = formattedTime[:2] + ":" + formattedTime[2:] + ":00"
		}
		deadlineVal = tglSelesai + " " + formattedTime
	}

	query := `
		UPDATE cat.at_jadwalujian
		SET waktupengerjaan = $1,
		    bobot = $2,
		    tampilkannilai = $3,
		    exam_deadline_at = $4,
		    grace_period_seconds = $5,
		    tglmulai = $6,
		    tglselesai = $7,
		    waktumulai = $8,
		    waktuselesai = $9,
		    max_violations = $10,
		    t_updatetime = NOW(),
		    t_updateact = 'u-jadwal'
		WHERE idjadwalujian = $11
		  AND (softdelete = '0' OR softdelete IS NULL)
	`
	_, err := r.db.ExecContext(ctx, query, waktuStr, req.Bobot, req.TampilkanNilai, deadlineVal, grace, tglMulaiVal, tglSelesaiVal, waktuMulaiVal, waktuSelesaiVal, maxViolations, sessionID)
	if err == nil && tglMulaiVal != nil {
		// Update child room schedules if any
		_, _ = r.db.ExecContext(ctx, `
			UPDATE cat.at_ruangujian
			SET tglmulai = $1, tglselesai = $2, waktumulai = $3, waktuselesai = $4, t_updatetime = NOW()
			WHERE idjadwalujian = $5 AND (softdelete = '0' OR softdelete IS NULL)
		`, tglMulaiVal, tglSelesaiVal, waktuMulaiVal, waktuSelesaiVal, sessionID)
	}
	return err
}

func (r *examinationRepository) DeleteSession(ctx context.Context, sessionID int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE cat.at_jadwalujian SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-jadwal' WHERE idjadwalujian = $1`,
		sessionID)
	return err
}

func (r *examinationRepository) RefreshToken(ctx context.Context, scheduleID int) (string, error) {
	newToken := utils.GenerateRandomExamToken(6)
	for attempts := 0; attempts < 10; attempts++ {
		var exists int
		_ = r.db.GetContext(ctx, &exists, `SELECT 1 FROM cat.at_jadwalujian WHERE token_ujian = $1 AND idjadwalujian != $2 AND (softdelete = '0' OR softdelete IS NULL) LIMIT 1`, newToken, scheduleID)
		if exists == 0 {
			break
		}
		newToken = utils.GenerateRandomExamToken(6)
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE cat.at_jadwalujian
		SET token_ujian = $1, t_updatetime = NOW(), t_updateact = 'refresh-token'
		WHERE idjadwalujian = $2 AND (softdelete = '0' OR softdelete IS NULL)
	`, newToken, scheduleID)
	if err != nil {
		return "", fmt.Errorf("gagal memperbarui token ujian: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return "", fmt.Errorf("jadwal ujian tidak ditemukan")
	}

	return newToken, nil
}

// =====================================================================
// Admin CRUD: Ruang per Sesi (at_ruangujian)
// =====================================================================

func (r *examinationRepository) GetSessionRooms(ctx context.Context, sessionID int) ([]*entity.RuangUjian, error) {
	query := `
		SELECT
			ru.idruangujian,
			ru.idjadwalujian,
			ru.koderuang,
			COALESCE(r.namaruang, ru.koderuang) AS namaruang,
			COALESCE(CAST(ru.tglmulai AS varchar), '') AS tglmulai,
			COALESCE(CAST(ru.tglselesai AS varchar), '') AS tglselesai,
			COALESCE(ru.waktumulai, '') AS waktumulai_str,
			COALESCE(ru.waktuselesai, '') AS waktuselesai_str,
			COALESCE(j.waktupengerjaan, '0060') AS waktupengerjaan_str,
			COALESCE(ru.jumlahpeserta, 0)::int AS jumlahpeserta,
			COALESCE(ru.prioritas, 1)::int AS prioritas
		FROM cat.at_ruangujian ru
		LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
		LEFT JOIN cat.at_jadwalujian j ON j.idjadwalujian = ru.idjadwalujian
		WHERE ru.idjadwalujian = $1
		  AND (ru.softdelete = '0' OR ru.softdelete IS NULL)
		ORDER BY ru.prioritas ASC
	`
	var rooms []*entity.RuangUjian
	err := r.db.SelectContext(ctx, &rooms, query, sessionID)
	if rooms == nil {
		rooms = []*entity.RuangUjian{}
	}
	return rooms, err
}

func sanitizeDateStr(d string) string {
	d = strings.TrimSpace(d)
	parts := strings.Split(d, "-")
	if len(parts) == 3 && len(parts[0]) > 4 {
		// e.g. "92026" -> "2026"
		parts[0] = parts[0][len(parts[0])-4:]
		return strings.Join(parts, "-")
	}
	return d
}

func (r *examinationRepository) AddRoomToSession(ctx context.Context, req *dto.CreateRoomSessionRequestDTO) (int, error) {
	prio := req.Prioritas
	if prio == 0 {
		prio = 1
	}
	jumlah := req.JumlahPeserta

	tglMulai := sanitizeDateStr(req.TglMulai)
	tglSelesai := sanitizeDateStr(req.TglSelesai)
	if tglSelesai == "" {
		tglSelesai = tglMulai
	}

	// waktu format HHMM -> store as char(4)
	waktuMulai := req.WaktuMulai
	if len(waktuMulai) == 5 && waktuMulai[2] == ':' {
		waktuMulai = waktuMulai[:2] + waktuMulai[3:] // convert HH:MM -> HHMM
	}

	waktuSelesai := req.WaktuSelesai
	if len(waktuSelesai) == 5 && waktuSelesai[2] == ':' {
		waktuSelesai = waktuSelesai[:2] + waktuSelesai[3:] // convert HH:MM -> HHMM
	}

	var newID int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO cat.at_ruangujian
			(idjadwalujian, koderuang, tglmulai, tglselesai, waktumulai, waktuselesai, jumlahpeserta, prioritas, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, $3::date, $4::date, $5, $6, $7, $8, '0', NOW(), 'i-ruangujian')
		RETURNING idruangujian
	`, req.ScheduleID, req.KodeRuang, tglMulai, tglSelesai, waktuMulai, waktuSelesai, jumlah, prio).Scan(&newID)
	return newID, err
}

func (r *examinationRepository) UpdateSessionRoom(ctx context.Context, roomSessionID int, req *dto.UpdateRoomSessionRequestDTO) error {
	prio := req.Prioritas
	if prio == 0 {
		prio = 1
	}
	jumlah := req.JumlahPeserta

	tglMulai := sanitizeDateStr(req.TglMulai)
	tglSelesai := sanitizeDateStr(req.TglSelesai)
	if tglSelesai == "" {
		tglSelesai = tglMulai
	}

	// waktu format HHMM -> store as char(4)
	waktuMulai := req.WaktuMulai
	if len(waktuMulai) == 5 && waktuMulai[2] == ':' {
		waktuMulai = waktuMulai[:2] + waktuMulai[3:] // convert HH:MM -> HHMM
	}

	waktuSelesai := req.WaktuSelesai
	if len(waktuSelesai) == 5 && waktuSelesai[2] == ':' {
		waktuSelesai = waktuSelesai[:2] + waktuSelesai[3:] // convert HH:MM -> HHMM
	}

	query := `
		UPDATE cat.at_ruangujian
		SET koderuang = $1,
		    tglmulai = $2::date,
		    tglselesai = $3::date,
		    waktumulai = $4,
		    waktuselesai = $5,
		    jumlahpeserta = $6,
		    prioritas = $7,
		    t_updatetime = NOW(),
		    t_updateact = 'u-ruangujian'
		WHERE idruangujian = $8
		  AND (softdelete = '0' OR softdelete IS NULL)
	`
	_, err := r.db.ExecContext(ctx, query, req.KodeRuang, tglMulai, tglSelesai, waktuMulai, waktuSelesai, jumlah, prio, roomSessionID)
	return err
}

func (r *examinationRepository) DeleteSessionRoom(ctx context.Context, roomSessionID int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE cat.at_ruangujian SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-ruangujian' WHERE idruangujian = $1`,
		roomSessionID)
	return err
}

// =====================================================================
// Admin CRUD: Peserta Ujian (at_pesertaujian)
// =====================================================================

func (r *examinationRepository) AddExamParticipant(ctx context.Context, examID int, participantCode string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cat.at_pesertaujian (kodepeserta, idujian, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, '0', NOW(), 'i-pesertaujian')
		ON CONFLICT (kodepeserta, idujian) DO UPDATE
		SET softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-pesertaujian'
	`, participantCode, examID)
	return err
}

func (r *examinationRepository) RemoveExamParticipant(ctx context.Context, examID int, participantCode string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE cat.at_pesertaujian
		SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-pesertaujian'
		WHERE idujian = $1 AND kodepeserta = $2
	`, examID, participantCode)
	return err
}

func (r *examinationRepository) GetAvailableParticipants(ctx context.Context, examID int) ([]*entity.AvailableParticipant, error) {
	query := `
		SELECT p.kodepeserta, p.nama, p.email, p.hp
		FROM cat.at_peserta p
		JOIN cat.at_ujian u ON u.idperiode = p.idperiode
		WHERE u.idujian = $1
		  AND (p.softdelete = '0' OR p.softdelete IS NULL)
		  AND NOT EXISTS (
		      SELECT 1 FROM cat.at_pesertaujian pu
		      WHERE pu.kodepeserta = p.kodepeserta
		        AND pu.idujian = $1
		        AND (pu.softdelete = '0' OR pu.softdelete IS NULL)
		  )
		ORDER BY p.nama ASC
		LIMIT 200
	`
	var list []*entity.AvailableParticipant
	err := r.db.SelectContext(ctx, &list, query, examID)
	if list == nil {
		list = []*entity.AvailableParticipant{}
	}
	return list, err
}

// =====================================================================
// Admin: Plotting & Distribusi Peserta ke Sesi & Ruang (at_jadwalpeserta)
// =====================================================================

type roomSlot struct {
	SessionID     int
	RoomSessionID int
	Capacity      int
	Used          int
}

func (r *examinationRepository) AutoDistributeParticipants(ctx context.Context, examID int, overwrite bool) (*dto.AutoDistributeResponseDTO, error) {
	// 1. Ambil seluruh sesi aktif untuk examID
	var sessions []struct {
		IDJadwalUjian int `db:"idjadwalujian"`
		NoJadwal      int `db:"nojadwal"`
	}
	err := r.db.SelectContext(ctx, &sessions, `
		SELECT idjadwalujian, COALESCE(nojadwal, 0)::int AS nojadwal
		FROM cat.at_jadwalujian
		WHERE idujian = $1 AND (softdelete = '0' OR softdelete IS NULL)
		ORDER BY nojadwal ASC
	`, examID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat sesi ujian: %w", err)
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("belum ada sesi ujian yang dibuat untuk paket ujian ini")
	}

	// 2. Ambil seluruh ruang aktif per sesi
	var slots []roomSlot
	sessionIDsUsed := make(map[int]bool)
	roomIDsUsed := make(map[int]bool)

	for _, sess := range sessions {
		var rooms []struct {
			IDRuangUjian  int `db:"idruangujian"`
			JumlahPeserta int `db:"jumlahpeserta"`
			Prioritas     int `db:"prioritas"`
		}
		err = r.db.SelectContext(ctx, &rooms, `
			SELECT idruangujian, COALESCE(jumlahpeserta, 0)::int AS jumlahpeserta, COALESCE(prioritas, 1)::int AS prioritas
			FROM cat.at_ruangujian
			WHERE idjadwalujian = $1 AND (softdelete = '0' OR softdelete IS NULL)
			ORDER BY prioritas ASC, idruangujian ASC
		`, sess.IDJadwalUjian)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat ruang sesi %d: %w", sess.NoJadwal, err)
		}

		for _, rm := range rooms {
			cap := rm.JumlahPeserta
			if cap <= 0 {
				cap = 30 // default fallback
			}
			slots = append(slots, roomSlot{
				SessionID:     sess.IDJadwalUjian,
				RoomSessionID: rm.IDRuangUjian,
				Capacity:      cap,
				Used:          0,
			})
		}
	}

	if len(slots) == 0 {
		return nil, fmt.Errorf("belum ada ruang ujian yang dikonfigurasi pada sesi-sesi ujian ini")
	}

	// 3. Ambil seluruh peserta terdaftar di ujian
	var participantCodes []string
	err = r.db.SelectContext(ctx, &participantCodes, `
		SELECT kodepeserta
		FROM cat.at_pesertaujian
		WHERE idujian = $1 AND (softdelete = '0' OR softdelete IS NULL)
		ORDER BY kodepeserta ASC
	`, examID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat peserta ujian: %w", err)
	}
	if len(participantCodes) == 0 {
		return nil, fmt.Errorf("belum ada peserta terdaftar pada paket ujian ini")
	}

	// 4. Handle overwrite
	if overwrite {
		_, err = r.db.ExecContext(ctx, `
			DELETE FROM cat.at_jadwalpeserta
			WHERE idjadwalujian IN (SELECT idjadwalujian FROM cat.at_jadwalujian WHERE idujian = $1)
		`, examID)
		if err != nil {
			return nil, fmt.Errorf("gagal mereset plotting jadwal lama: %w", err)
		}
	} else {
		// Filter peserta yang sudah memiliki jadwal di ujian ini
		var assignedCodes []string
		_ = r.db.SelectContext(ctx, &assignedCodes, `
			SELECT DISTINCT jp.kodepeserta
			FROM cat.at_jadwalpeserta jp
			JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian
			WHERE ju.idujian = $1 AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
		`, examID)
		assignedMap := make(map[string]bool)
		for _, c := range assignedCodes {
			assignedMap[c] = true
		}

		// Count currently used capacity per room
		for i := range slots {
			var count int
			_ = r.db.GetContext(ctx, &count, `
				SELECT COUNT(*) FROM cat.at_jadwalpeserta
				WHERE idruangujian = $1 AND (softdelete = '0' OR softdelete IS NULL)
			`, slots[i].RoomSessionID)
			slots[i].Used = count
		}

		var unassigned []string
		for _, c := range participantCodes {
			if !assignedMap[c] {
				unassigned = append(unassigned, c)
			}
		}
		participantCodes = unassigned
	}

	// 5. Distribusikan peserta ke slots
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAssigned := 0
	unassignedCount := 0
	slotIdx := 0

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO cat.at_jadwalpeserta
			(kodepeserta, idjadwalujian, idruangujian, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, $3, '0', NOW(), 'auto-distribute')
		ON CONFLICT (kodepeserta, idjadwalujian) DO UPDATE
		SET idruangujian = EXCLUDED.idruangujian, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-auto-distribute'
	`)
	if err != nil {
		return nil, fmt.Errorf("gagal menyiapkan query insert jadwal: %w", err)
	}
	defer stmt.Close()

	for _, code := range participantCodes {
		// Cari slot yang masih punya sisa kapasitas
		for slotIdx < len(slots) && slots[slotIdx].Used >= slots[slotIdx].Capacity {
			slotIdx++
		}

		if slotIdx >= len(slots) {
			// Kapasitas seluruh ruang pada seluruh sesi telah penuh
			unassignedCount++
			continue
		}

		curSlot := &slots[slotIdx]
		_, err = stmt.ExecContext(ctx, code, curSlot.SessionID, curSlot.RoomSessionID)
		if err != nil {
			return nil, fmt.Errorf("gagal menugaskan peserta %s: %w", code, err)
		}

		curSlot.Used++
		sessionIDsUsed[curSlot.SessionID] = true
		roomIDsUsed[curSlot.RoomSessionID] = true
		totalAssigned++
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyimpan hasil plotting: %w", err)
	}

	msg := fmt.Sprintf("Berhasil mendistribusikan %d peserta ke dalam %d sesi dan %d ruang.",
		totalAssigned, len(sessionIDsUsed), len(roomIDsUsed))
	if unassignedCount > 0 {
		msg += fmt.Sprintf(" Peringatan: Terdapat %d peserta belum terplotting karena kapasitas ruang penuh.", unassignedCount)
	}

	return &dto.AutoDistributeResponseDTO{
		TotalAssigned:   totalAssigned,
		SessionsUsed:    len(sessionIDsUsed),
		RoomsUsed:       len(roomIDsUsed),
		UnassignedCount: unassignedCount,
		Message:         msg,
	}, nil
}

func (r *examinationRepository) GetRoomParticipants(ctx context.Context, roomSessionID int) ([]*entity.RoomParticipantRow, error) {
	query := `
		SELECT 
			p.kodepeserta,
			p.nama,
			p.email,
			w.namawilayah,
			jp.idjadwalujian,
			COALESCE(jp.idruangujian, 0) AS idruangujian,
			COALESCE(r.namaruang, ru.koderuang, '') AS namaruang,
			COALESCE(p.islogin, 0) AS islogin,
			COALESCE(jp.is_verified, 0) AS is_verified,
			jp.barcode_scanned_at
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_peserta p ON p.kodepeserta = jp.kodepeserta
		LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian
		LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
		LEFT JOIN cat.ms_wilayah w ON (w.idwilayah = CAST(p.idkota AS TEXT) OR w.idwilayah = p.idkota_lama)
		WHERE jp.idruangujian = $1 AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
		ORDER BY p.nama ASC
	`
	var list []*entity.RoomParticipantRow
	err := r.db.SelectContext(ctx, &list, query, roomSessionID)
	if list == nil {
		list = []*entity.RoomParticipantRow{}
	}
	return list, err
}

func (r *examinationRepository) AddRoomParticipants(ctx context.Context, roomSessionID int, participantCodes []string) error {
	// Ambil idjadwalujian dan idujian dari roomSessionID
	var info struct {
		IDJadwalUjian int `db:"idjadwalujian"`
		IDUjian       int `db:"idujian"`
	}
	err := r.db.GetContext(ctx, &info, `
		SELECT ru.idjadwalujian, ju.idujian
		FROM cat.at_ruangujian ru
		JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = ru.idjadwalujian
		WHERE ru.idruangujian = $1 AND (ru.softdelete = '0' OR ru.softdelete IS NULL)
	`, roomSessionID)
	if err != nil {
		return fmt.Errorf("ruang sesi tidak ditemukan: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Statement untuk menghapus alokasi sesi lama pada paket ujian ini (mencegah double sesi)
	delStmt, err := tx.PreparexContext(ctx, `
		DELETE FROM cat.at_jadwalpeserta
		WHERE kodepeserta = $1
		  AND idjadwalujian IN (
		      SELECT idjadwalujian FROM cat.at_jadwalujian WHERE idujian = $2
		  )
	`)
	if err != nil {
		return err
	}
	defer delStmt.Close()

	// 2. Statement untuk memasukkan ke sesi dan ruang yang baru
	insStmt, err := tx.PreparexContext(ctx, `
		INSERT INTO cat.at_jadwalpeserta
			(kodepeserta, idjadwalujian, idruangujian, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, $3, '0', NOW(), 'i-roomparticipant')
		ON CONFLICT (kodepeserta, idjadwalujian) DO UPDATE
		SET idruangujian = EXCLUDED.idruangujian, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-roomparticipant'
	`)
	if err != nil {
		return err
	}
	defer insStmt.Close()

	for _, code := range participantCodes {
		if code == "" {
			continue
		}
		// Hapus penugasan sesi lama di ujian ini terlebih dahulu (clean transfer)
		_, err = delStmt.ExecContext(ctx, code, info.IDUjian)
		if err != nil {
			return fmt.Errorf("gagal memindahkan peserta %s: %w", code, err)
		}

		_, err = insStmt.ExecContext(ctx, code, info.IDJadwalUjian, roomSessionID)
		if err != nil {
			return fmt.Errorf("gagal menugaskan peserta %s: %w", code, err)
		}
	}

	return tx.Commit()
}

func (r *examinationRepository) RemoveRoomParticipant(ctx context.Context, roomSessionID int, participantCode string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM cat.at_jadwalpeserta
		WHERE idruangujian = $1 AND kodepeserta = $2
	`, roomSessionID, participantCode)
	return err
}

func (r *examinationRepository) VerifyParticipantBarcode(ctx context.Context, scheduleID int, participantCode string, barcode string) (*dto.VerifyBarcodeResponseDTO, error) {
	barcode = strings.TrimSpace(barcode)
	participantCode = strings.TrimSpace(participantCode)
	if barcode == "" {
		return nil, fmt.Errorf("data barcode tidak boleh kosong")
	}

	// Fetch participant details in this schedule
	var row struct {
		KodePeserta string `db:"kodepeserta"`
		Nama        string `db:"nama"`
		IsVerified  int    `db:"is_verified"`
		NamaRuang   string `db:"namaruang"`
	}
	query := `
		SELECT jp.kodepeserta, p.nama, COALESCE(jp.is_verified, 0) as is_verified,
		       COALESCE(r.namaruang, ru.koderuang, 'Lab Ujian') as namaruang
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_peserta p ON p.kodepeserta = jp.kodepeserta
		LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian
		LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
		WHERE jp.idjadwalujian = $1 AND jp.kodepeserta = $2 AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
	`
	err := r.db.GetContext(ctx, &row, query, scheduleID, participantCode)
	if err != nil {
		return nil, fmt.Errorf("peserta tidak terdaftar pada sesi ujian ini")
	}

	cleanBarcode := strings.TrimSpace(barcode)
	if cleanBarcode != participantCode && !strings.Contains(cleanBarcode, participantCode) {
		return nil, fmt.Errorf("barcode tidak sesuai dengan kartu ujian peserta (%s)", participantCode)
	}

	now := time.Now()
	updQ := `
		UPDATE cat.at_jadwalpeserta 
		SET is_verified = 1, barcode_scanned_at = $1, scanned_by = 'PARTICIPANT', barcode_data = $2
		WHERE idjadwalujian = $3 AND kodepeserta = $4
	`
	_, err = r.db.ExecContext(ctx, updQ, now, barcode, scheduleID, participantCode)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui status verifikasi: %w", err)
	}

	return &dto.VerifyBarcodeResponseDTO{
		ParticipantCode: row.KodePeserta,
		Name:            row.Nama,
		IsVerified:      1,
		ScannedAt:       now.Format(time.RFC3339),
		ScannedBy:       "PARTICIPANT",
		Message:         "Verifikasi barcode kehadiran berhasil. Akses ujian dibuka.",
		RoomName:        row.NamaRuang,
	}, nil
}

func (r *examinationRepository) VerifyProctorBarcode(ctx context.Context, scheduleID int, barcode string, proctorUser string) (*dto.VerifyBarcodeResponseDTO, error) {
	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return nil, fmt.Errorf("data barcode tidak boleh kosong")
	}
	if proctorUser == "" {
		proctorUser = "PROCTOR"
	}

	var row struct {
		KodePeserta       string  `db:"kodepeserta"`
		Nama              string  `db:"nama"`
		IsVerified        int     `db:"is_verified"`
		NamaRuang         string  `db:"namaruang"`
		NoJadwal          int     `db:"nojadwal"`
		ExistingScannedAt *string `db:"existing_scanned_at"`
		ExistingScannedBy *string `db:"existing_scanned_by"`
	}
	query := `
		SELECT jp.kodepeserta, p.nama, COALESCE(jp.is_verified, 0) as is_verified,
		       COALESCE(r.namaruang, ru.koderuang, 'Lab Ujian') as namaruang,
		       COALESCE(ju.nojadwal, 1) as nojadwal,
		       CAST(jp.barcode_scanned_at AS text) as existing_scanned_at,
		       jp.scanned_by as existing_scanned_by
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_peserta p ON p.kodepeserta = jp.kodepeserta
		LEFT JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian
		LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian
		LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
		WHERE jp.idjadwalujian = $1 
		  AND (
		      jp.barcode_data = $2 
		      OR 'CBT-' || jp.idjadwalujian || '-' || jp.kodepeserta = $2
		      OR jp.kodepeserta = $2
		  )
		  AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &row, query, scheduleID, barcode)
	if err != nil {
		return nil, fmt.Errorf("peserta dengan barcode '%s' tidak valid atau tidak terdaftar pada sesi ujian ini", barcode)
	}

	alreadyVerified := row.IsVerified == 1
	now := time.Now()
	updQ := `
		UPDATE cat.at_jadwalpeserta 
		SET is_verified = 1, barcode_scanned_at = COALESCE(barcode_scanned_at, $1), 
		    scanned_by = COALESCE(scanned_by, $2), barcode_data = $3
		WHERE idjadwalujian = $4 AND kodepeserta = $5
	`
	_, err = r.db.ExecContext(ctx, updQ, now, proctorUser, barcode, scheduleID, row.KodePeserta)
	if err != nil {
		return nil, fmt.Errorf("gagal memverifikasi kehadiran peserta: %w", err)
	}

	// Calculate session attendance stats
	var statRow struct {
		Total    int `db:"total"`
		Verified int `db:"verified"`
	}
	statQ := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN is_verified = 1 THEN 1 END) as verified
		FROM cat.at_jadwalpeserta
		WHERE idjadwalujian = $1 AND (softdelete = '0' OR softdelete IS NULL)
	`
	_ = r.db.GetContext(ctx, &statRow, statQ, scheduleID)
	pct := 0.0
	if statRow.Total > 0 {
		pct = (float64(statRow.Verified) / float64(statRow.Total)) * 100
	}

	msg := fmt.Sprintf("Presensi %s (%s) berhasil diverifikasi.", row.Nama, row.KodePeserta)
	if alreadyVerified {
		prevTime := "sebelumnya"
		if row.ExistingScannedAt != nil {
			prevTime = *row.ExistingScannedAt
		}
		msg = fmt.Sprintf("PERINGATAN DUPLIKAT: Peserta %s sudah pernah diverifikasi (%s).", row.Nama, prevTime)
	}

	return &dto.VerifyBarcodeResponseDTO{
		ParticipantCode: row.KodePeserta,
		Name:            row.Nama,
		IsVerified:      1,
		AlreadyVerified: alreadyVerified,
		ScannedAt:       now.Format(time.RFC3339),
		ScannedBy:       proctorUser,
		Message:         msg,
		RoomName:        row.NamaRuang,
		SessionNumber:   row.NoJadwal,
		Stats: &dto.AttendanceStatsDTO{
			TotalParticipants: statRow.Total,
			VerifiedCount:     statRow.Verified,
			UnverifiedCount:   statRow.Total - statRow.Verified,
			AttendancePct:     pct,
		},
	}, nil
}

func (r *examinationRepository) GetSessionAttendanceSummary(ctx context.Context, scheduleID int) (*dto.AttendanceSummaryResponseDTO, error) {
	var sessionInfo struct {
		ScheduleID int    `db:"idjadwalujian"`
		SessionNo  int    `db:"nojadwal"`
		ExamName   string `db:"namaujian"`
	}
	sessQ := `
		SELECT ju.idjadwalujian, COALESCE(ju.nojadwal, 1) as nojadwal, COALESCE(u.namaujian, '') as namaujian
		FROM cat.at_jadwalujian ju
		JOIN cat.at_ujian u ON u.idujian = ju.idujian
		WHERE ju.idjadwalujian = $1
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &sessionInfo, sessQ, scheduleID); err != nil {
		return nil, fmt.Errorf("sesi ujian tidak ditemukan: %w", err)
	}

	type partRow struct {
		ParticipantCode  string  `db:"kodepeserta"`
		Name             string  `db:"nama"`
		RoomName         string  `db:"namaruang"`
		IsVerified       int     `db:"is_verified"`
		BarcodeScannedAt *string `db:"barcode_scanned_at"`
		ScannedBy        *string `db:"scanned_by"`
	}

	partsQ := `
		SELECT 
			jp.kodepeserta,
			p.nama,
			COALESCE(r.namaruang, ru.koderuang, 'Lab Ujian') as namaruang,
			COALESCE(jp.is_verified, 0) as is_verified,
			CAST(jp.barcode_scanned_at AS text) as barcode_scanned_at,
			jp.scanned_by
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_peserta p ON p.kodepeserta = jp.kodepeserta
		LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian
		LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
		WHERE jp.idjadwalujian = $1 AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
		ORDER BY jp.is_verified DESC, jp.barcode_scanned_at DESC NULLS LAST, p.nama ASC
	`
	var rows []partRow
	_ = r.db.SelectContext(ctx, &rows, partsQ, scheduleID)

	total := len(rows)
	verified := 0
	participants := make([]dto.AttendanceParticipantItemDTO, 0, total)
	for idx, pr := range rows {
		if pr.IsVerified == 1 {
			verified++
		}
		participants = append(participants, dto.AttendanceParticipantItemDTO{
			ParticipantCode:  pr.ParticipantCode,
			Name:             pr.Name,
			RoomName:         pr.RoomName,
			DeskNumber:       idx + 1,
			SessionNumber:    sessionInfo.SessionNo,
			IsVerified:       pr.IsVerified,
			BarcodeScannedAt: pr.BarcodeScannedAt,
			ScannedBy:        pr.ScannedBy,
		})
	}

	pct := 0.0
	if total > 0 {
		pct = (float64(verified) / float64(total)) * 100
	}

	return &dto.AttendanceSummaryResponseDTO{
		ScheduleID: sessionInfo.ScheduleID,
		SessionNo:  sessionInfo.SessionNo,
		ExamName:   sessionInfo.ExamName,
		Stats: dto.AttendanceStatsDTO{
			TotalParticipants: total,
			VerifiedCount:     verified,
			UnverifiedCount:   total - verified,
			AttendancePct:     pct,
		},
		Participants: participants,
	}, nil
}

func (r *examinationRepository) SetParticipantVerificationStatus(ctx context.Context, scheduleID int, participantCode string, isVerified int, verifier string) error {
	var err error
	if isVerified == 1 {
		now := time.Now()
		_, err = r.db.ExecContext(ctx, `
			UPDATE cat.at_jadwalpeserta 
			SET is_verified = 1, barcode_scanned_at = $1, scanned_by = $2
			WHERE idjadwalujian = $3 AND kodepeserta = $4
		`, now, verifier, scheduleID, participantCode)
	} else {
		_, err = r.db.ExecContext(ctx, `
			UPDATE cat.at_jadwalpeserta 
			SET is_verified = 0, barcode_scanned_at = NULL, scanned_by = NULL
			WHERE idjadwalujian = $1 AND kodepeserta = $2
		`, scheduleID, participantCode)
	}
	return err
}

func (r *examinationRepository) GetParticipantSchedules(ctx context.Context, participantCode string) ([]*dto.ParticipantScheduleDTO, error) {
	type rawSched struct {
		ScheduleID       int        `db:"schedule_id"`
		ExamID           int        `db:"exam_id"`
		ExamName         string     `db:"exam_name"`
		QuestionBankCode string     `db:"question_bank_code"`
		RoomID           int        `db:"room_id"`
		RoomName         string     `db:"room_name"`
		ExamDate         time.Time  `db:"exam_date"`
		EndDate          time.Time  `db:"end_date"`
		StartTime        string     `db:"start_time"`
		EndTime          string     `db:"end_time"`
		DurationMinutes  int        `db:"duration_minutes"`
		SessionToken     string     `db:"session_token"`
		Capacity         int        `db:"capacity"`
		ScheduleStatus   string     `db:"schedule_status"`
		IsVerified       int        `db:"is_verified"`
		BarcodeScannedAt *time.Time `db:"barcode_scanned_at"`
		ExamBarcode      string     `db:"exam_barcode"`
		HasStarted       bool       `db:"has_started"`
		IsFinished       bool       `db:"is_finished"`
		RemainingSeconds int        `db:"remaining_seconds"`
		MaxViolations    int        `db:"max_violations"`
		IsEP             bool       `db:"is_ep"`
		ExamType         string     `db:"exam_type"`
	}

	schedQuery := `
		SELECT j.idjadwalujian as schedule_id, j.idujian as exam_id, u.namaujian as exam_name, 
		       COALESCE((SELECT su.kodesoal FROM cat.at_soalujian su WHERE su.idjadwalujian = j.idjadwalujian LIMIT 1), 'SOAL_2026') as question_bank_code, 
		       COALESCE(r.idruangujian, 1) as room_id, 
		       COALESCE(r.koderuang, 'Lab 1') as room_name, 
		       COALESCE(r.tglmulai, j.tglmulai, NOW())::date as exam_date, 
		       COALESCE(r.tglselesai, j.tglselesai, r.tglmulai, j.tglmulai, NOW())::date as end_date, 
		       COALESCE(nullif(trim(r.waktumulai), ''), nullif(trim(j.waktumulai), ''), '00:00') as start_time, 
		       COALESCE(nullif(trim(r.waktuselesai), ''), nullif(trim(j.waktuselesai), ''), '23:59') as end_time, 
		       COALESCE(j.waktupengerjaan::integer, 90) as duration_minutes, 
		       COALESCE(j.token_ujian, '') as session_token, 
		       COALESCE(r.jumlahpeserta::integer, 40) as capacity,
		       CASE
		         WHEN (NOW() AT TIME ZONE 'Asia/Jakarta') < 
		              (COALESCE(r.tglmulai, j.tglmulai, CURRENT_DATE)::date + 
		               CASE 
		                 WHEN length(trim(COALESCE(r.waktumulai, j.waktumulai, '00:00'))) = 4 AND trim(COALESCE(r.waktumulai, j.waktumulai, '00:00')) !~ ':' 
		                 THEN (substring(trim(COALESCE(r.waktumulai, j.waktumulai, '00:00')) from 1 for 2) || ':' || substring(trim(COALESCE(r.waktumulai, j.waktumulai, '00:00')) from 3 for 2))::time 
		                 ELSE COALESCE(nullif(trim(COALESCE(r.waktumulai, j.waktumulai, '')), ''), '00:00')::time 
		               END)
		           THEN 'UPCOMING'
		         WHEN (NOW() AT TIME ZONE 'Asia/Jakarta') > 
		              (COALESCE(r.tglselesai, j.tglselesai, r.tglmulai, j.tglmulai, CURRENT_DATE)::date + 
		               CASE 
		                 WHEN length(trim(COALESCE(r.waktuselesai, j.waktuselesai, '23:59'))) = 4 AND trim(COALESCE(r.waktuselesai, j.waktuselesai, '23:59')) !~ ':' 
		                 THEN (substring(trim(COALESCE(r.waktuselesai, j.waktuselesai, '23:59')) from 1 for 2) || ':' || substring(trim(COALESCE(r.waktuselesai, j.waktuselesai, '23:59')) from 3 for 2))::time 
		                 ELSE COALESCE(nullif(trim(COALESCE(r.waktuselesai, j.waktuselesai, '')), ''), '23:59')::time 
		               END)
		           THEN 'EXPIRED'
		         ELSE 'ACTIVE'
		       END as schedule_status,
		       COALESCE(jp.is_verified, 0) as is_verified,
		       jp.barcode_scanned_at,
		       COALESCE(jp.barcode_data, 'CBT-' || jp.idjadwalujian || '-' || jp.kodepeserta) AS exam_barcode,
		       (jp.tglmulai IS NOT NULL) as has_started,
		       (jp.tglselesai IS NOT NULL) as is_finished,
		       COALESCE(jp.remaining_seconds, 0) as remaining_seconds,
		       COALESCE(j.max_violations, u.max_violations, 5) as max_violations,
		       false as is_ep,
		       'REGULAR' as exam_type
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_jadwalujian j ON j.idjadwalujian = jp.idjadwalujian AND (j.softdelete = '0' OR j.softdelete IS NULL)
		JOIN cat.at_ujian u ON u.idujian = j.idujian AND (u.softdelete = '0' OR u.softdelete IS NULL)
		LEFT JOIN cat.at_ruangujian r ON r.idruangujian = jp.idruangujian
		WHERE jp.kodepeserta = $1 AND (jp.softdelete = '0' OR jp.softdelete IS NULL)

		UNION ALL

		SELECT s.id_ep_schedule as schedule_id, s.id_ep_exam as exam_id, e.exam_name as exam_name,
		       'TOEFL_EP' as question_bank_code,
		       s.id_ruang as room_id,
		       COALESCE(r.namaruang, ar.koderuang, 'Ruang Ujian ' || s.id_ruang) as room_name,
		       s.exam_date::date as exam_date,
		       s.exam_date::date as end_date,
		       s.start_time::text as start_time,
		       s.end_time::text as end_time,
		       115 as duration_minutes,
		       COALESCE(s.session_token, 'TOKEN') as session_token,
		       COALESCE(s.capacity, 40) as capacity,
		       CASE
		         WHEN (NOW() AT TIME ZONE 'Asia/Jakarta') < (s.exam_date::date + COALESCE(nullif(trim(s.start_time::text), ''), '00:00')::time) THEN 'UPCOMING'
		         WHEN (NOW() AT TIME ZONE 'Asia/Jakarta') > (s.exam_date::date + COALESCE(nullif(trim(s.end_time::text), ''), '23:59')::time) THEN 'EXPIRED'
		         ELSE 'ACTIVE'
		       END as schedule_status,
		       1 as is_verified,
		       NULL::timestamp as barcode_scanned_at,
		       'EP-' || s.id_ep_schedule || '-' || sp.kodepeserta AS exam_barcode,
		       (sp.session_status != 'NOT_STARTED') as has_started,
		       (sp.session_status = 'SUBMITTED' OR sp.finished_at IS NOT NULL) as is_finished,
		       0 as remaining_seconds,
		       5 as max_violations,
		       true as is_ep,
		       'TOEFL_EP' as exam_type
		FROM cat.ep_schedule_participants sp
		JOIN cat.ep_schedules s ON s.id_ep_schedule = sp.id_ep_schedule AND s.is_active = true
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		LEFT JOIN cat.at_ruangujian ar ON ar.idruangujian = s.id_ruang
		LEFT JOIN cat.at_ruang r ON r.koderuang = ar.koderuang OR r.koderuang = s.id_ruang::text
		WHERE sp.kodepeserta = $1 AND (sp.softdelete = '0' OR sp.softdelete IS NULL)
	`

	var rows []rawSched
	err := r.db.SelectContext(ctx, &rows, schedQuery, participantCode)
	if err != nil {
		return nil, err
	}

	var result []*dto.ParticipantScheduleDTO
	for _, sc := range rows {
		var scannedAtStr *string
		if sc.BarcodeScannedAt != nil {
			s := sc.BarcodeScannedAt.Format(time.RFC3339)
			scannedAtStr = &s
		}
		result = append(result, &dto.ParticipantScheduleDTO{
			ScheduleID:       sc.ScheduleID,
			ExamID:           sc.ExamID,
			ExamName:         sc.ExamName,
			QuestionBankCode: sc.QuestionBankCode,
			RoomID:           sc.RoomID,
			RoomName:         sc.RoomName,
			ExamDate:         sc.ExamDate.Format(time.RFC3339),
			EndDate:          sc.EndDate.Format(time.RFC3339),
			StartTime:        formatTimeOnly(sc.StartTime),
			EndTime:          formatTimeOnly(sc.EndTime),
			DurationMinutes:  sc.DurationMinutes,
			SessionToken:     sc.SessionToken,
			Capacity:         sc.Capacity,
			ScheduleStatus:   sc.ScheduleStatus,
			IsVerified:       sc.IsVerified,
			BarcodeScannedAt: scannedAtStr,
			ExamBarcode:      sc.ExamBarcode,
			HasStarted:       sc.HasStarted,
			IsFinished:       sc.IsFinished,
			RemainingSeconds: sc.RemainingSeconds,
			MaxViolations:    sc.MaxViolations,
			IsEP:             sc.IsEP,
			ExamType:         sc.ExamType,
		})
	}
	return result, nil
}

func formatTimeOnly(t string) string {
	t = strings.TrimSpace(t)
	if len(t) == 4 && !strings.Contains(t, ":") {
		return t[:2] + ":" + t[2:]
	}
	if strings.Contains(t, ":") {
		parts := strings.Split(t, ":")
		if len(parts) >= 2 {
			hh := parts[0]
			if len(hh) == 1 {
				hh = "0" + hh
			}
			mm := parts[1]
			if len(mm) == 1 {
				mm = "0" + mm
			}
			return hh + ":" + mm
		}
	}
	if t == "" {
		return "00:00"
	}
	return t
}

// === Bank Soal per Sesi ===

func (r *examinationRepository) GetSessionBankSoal(ctx context.Context, sessionID int) ([]*entity.SoalUjian, error) {
	query := `
		SELECT
			su.idjadwalujian,
			su.kodesoal,
			COALESCE(s.namasoal, '') AS namasoal,
			COALESCE(su.nourut, 0) AS nourut,
			COALESCE(su.jumlahpertanyaan, 0) AS jumlahpertanyaan,
			COALESCE(su.jumlah_soal_statis, 0) AS jumlah_soal_statis,
			COALESCE(su.jumlah_kasus_dinamis, 0) AS jumlah_kasus_dinamis,
			COALESCE(su.bobot, 0) AS bobot,
			su.kodeskor,
			COALESCE(qcount.total, 0) AS total_available,
			COALESCE(qcount.total_statis, 0) AS total_statis_available,
			COALESCE(qcount.total_stimulus, 0) AS total_stimulus_available
		FROM cat.at_soalujian su
		JOIN cat.at_soal s ON s.kodesoal = su.kodesoal
		LEFT JOIN (
			SELECT q.kodesoal, 
			       (SELECT COUNT(*) FROM cat.at_pertanyaan p WHERE p.kodesoal = q.kodesoal AND (p.softdelete = '0' OR p.softdelete IS NULL))
			       +
			       (SELECT COUNT(*) FROM cat.cat_stimulus_items si JOIN cat.cat_bank_stimulus bs ON bs.id_stimulus = si.id_stimulus WHERE bs.kodesoal = q.kodesoal AND (bs.softdelete = '0' OR bs.softdelete IS NULL) AND (si.softdelete = '0' OR si.softdelete IS NULL))
			       AS total,
			       (SELECT COUNT(*) FROM cat.at_pertanyaan p WHERE p.kodesoal = q.kodesoal AND (p.softdelete = '0' OR p.softdelete IS NULL)) AS total_statis,
			       (SELECT COUNT(*) FROM cat.cat_bank_stimulus bs WHERE bs.kodesoal = q.kodesoal AND (bs.softdelete = '0' OR bs.softdelete IS NULL)) AS total_stimulus
			FROM cat.at_soal q
			WHERE (q.softdelete = '0' OR q.softdelete IS NULL)
		) qcount ON qcount.kodesoal = su.kodesoal
		WHERE su.idjadwalujian = $1 AND su.softdelete = '0'
		ORDER BY su.nourut, su.kodesoal
	`
	var results []*entity.SoalUjian
	err := r.db.SelectContext(ctx, &results, query, sessionID)
	return results, err
}

func (r *examinationRepository) AddBankSoalToSession(ctx context.Context, sessionID int, req *dto.AddBankSoalToSessionDTO) error {
	kodeSkor := req.KodeSkor
	if kodeSkor == "" {
		kodeSkor = "1"
	}

	// Get next nourut
	var maxUrut int
	_ = r.db.GetContext(ctx, &maxUrut, `SELECT COALESCE(MAX(nourut), 0) FROM cat.at_soalujian WHERE idjadwalujian = $1 AND softdelete = '0'`, sessionID)

	statisCount := req.JumlahSoalStatis
	dinamisCount := req.JumlahKasusDinamis
	totalQ := req.JumlahPertanyaan

	if statisCount == 0 && dinamisCount == 0 && totalQ > 0 {
		var hasDinamis int
		_ = r.db.GetContext(ctx, &hasDinamis, `SELECT COUNT(*) FROM cat.cat_bank_stimulus WHERE kodesoal = $1 AND (softdelete = '0' OR softdelete IS NULL)`, req.KodeSoal)
		if hasDinamis > 0 {
			dinamisCount = totalQ
		} else {
			statisCount = totalQ
		}
	} else if totalQ <= 0 {
		totalQ = statisCount + dinamisCount
	}

	query := `
		INSERT INTO cat.at_soalujian (idjadwalujian, kodesoal, nourut, jumlahpertanyaan, jumlah_soal_statis, jumlah_kasus_dinamis, bobot, kodeskor, softdelete)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '0')
		ON CONFLICT (idjadwalujian, kodesoal) DO UPDATE SET
			jumlahpertanyaan = EXCLUDED.jumlahpertanyaan,
			jumlah_soal_statis = EXCLUDED.jumlah_soal_statis,
			jumlah_kasus_dinamis = EXCLUDED.jumlah_kasus_dinamis,
			bobot = EXCLUDED.bobot,
			kodeskor = EXCLUDED.kodeskor,
			nourut = EXCLUDED.nourut,
			softdelete = '0'
	`
	_, err := r.db.ExecContext(ctx, query, sessionID, req.KodeSoal, maxUrut+1, totalQ, statisCount, dinamisCount, req.Bobot, kodeSkor)
	if err != nil {
		return err
	}

	// Invalidate unstarted participant snapshots so updated bank soal or quotas take effect immediately
	_, _ = r.db.ExecContext(ctx, `
		UPDATE cat.at_jadwalpeserta 
		SET session_snapshot = NULL 
		WHERE idjadwalujian = $1 
		  AND (SELECT COUNT(*) FROM cat.at_jawabanpeserta WHERE idjadwalujian = $1 AND kodepeserta = cat.at_jadwalpeserta.kodepeserta AND (softdelete = '0' OR softdelete IS NULL)) = 0
	`, sessionID)

	return nil
}

func (r *examinationRepository) RemoveBankSoalFromSession(ctx context.Context, sessionID int, kodeSoal string) error {
	cleanKode := strings.TrimSpace(kodeSoal)
	// Try softdelete first (standard pattern in Poltekkes CAT to prevent FK constraint violations)
	res, err := r.db.ExecContext(ctx,
		`UPDATE cat.at_soalujian 
		 SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-soalujian' 
		 WHERE idjadwalujian = $1 AND (kodesoal = $2 OR TRIM(kodesoal) = $2 OR LOWER(TRIM(kodesoal)) = LOWER($2))`,
		sessionID, cleanKode)
	if err != nil {
		// If update fails, attempt DELETE
		_, err = r.db.ExecContext(ctx,
			`DELETE FROM cat.at_soalujian WHERE idjadwalujian = $1 AND (kodesoal = $2 OR TRIM(kodesoal) = $2)`,
			sessionID, cleanKode)
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		// If 0 rows updated (e.g. was already softdeleted or case mismatch), try hard DELETE
		_, _ = r.db.ExecContext(ctx,
			`DELETE FROM cat.at_soalujian WHERE idjadwalujian = $1 AND (kodesoal = $2 OR TRIM(kodesoal) = $2)`,
			sessionID, cleanKode)
	}

	// Invalidate unstarted participant snapshots so updated bank soal or quotas take effect immediately
	_, _ = r.db.ExecContext(ctx, `
		UPDATE cat.at_jadwalpeserta 
		SET session_snapshot = NULL 
		WHERE idjadwalujian = $1 
		  AND (SELECT COUNT(*) FROM cat.at_jawabanpeserta WHERE idjadwalujian = $1 AND kodepeserta = cat.at_jadwalpeserta.kodepeserta AND (softdelete = '0' OR softdelete IS NULL)) = 0
	`, sessionID)

	return nil
}

func (r *examinationRepository) ImportSipenmaruToExam(ctx context.Context, examID int, req *dto.ExamImportSipenmaruRequestDTO) (*dto.ExamImportSipenmaruResponseDTO, error) {
	resp := &dto.ExamImportSipenmaruResponseDTO{
		Errors: make([]string, 0),
	}

	// 1. Resolve CBT period ID if 0
	cbtPeriodID := req.CBTPeriodID
	if cbtPeriodID <= 0 {
		_ = r.db.GetContext(ctx, &cbtPeriodID, "SELECT idperiode FROM cat.at_ujian WHERE idujian = $1", examID)
	}
	if cbtPeriodID <= 0 {
		return nil, fmt.Errorf("periode ujian tidak ditemukan untuk ujian #%d", examID)
	}

	// 2. Query candidates from pendaftaran.pd_pendaftar
	baseQuery := `
		SELECT 
			p.idpendaftar,
			COALESCE(p.namapendaftar, '') AS nama,
			COALESCE(p.jk, 'L') AS jk,
			COALESCE(p.hp, '') AS hp,
			COALESCE(p.email, '') AS email,
			COALESCE(p.alamat, '') AS alamat,
			p.idkota,
			COALESCE(p.idperiode::text, '') AS id_periode,
			COALESCE(p.idsistemkuliah, 0) AS id_sistem_kuliah,
			COALESCE(p.idjalurpendaftaran, 0) AS id_jalur_pendaftaran,
			COALESCE(p.idgelombang, 0) AS id_gelombang,
			p.idunit1::text AS id_unit1,
			p.idunit2::text AS id_unit2,
			COALESCE(p.isadministrasi, 0) AS is_administrasi,
			COALESCE(p.noujian, '') AS no_ujian_sipenmaru,
			'' AS existing_kode_peserta
		FROM pendaftaran.pd_pendaftar p
		WHERE (p.softdelete = '0' OR p.softdelete IS NULL)
	`
	var conditions []string
	var args []interface{}
	argIdx := 1

	pmbPeriodStr := req.GetPMBPeriodString()
	if pmbPeriodStr != "" {
		conditions = append(conditions, fmt.Sprintf("p.idperiode::text = $%d", argIdx))
		args = append(args, pmbPeriodStr)
		argIdx++
	}
	if req.IDSistemKuliah > 0 {
		conditions = append(conditions, fmt.Sprintf("p.idsistemkuliah = $%d", argIdx))
		args = append(args, req.IDSistemKuliah)
		argIdx++
	}
	if req.IDJalurPendaftaran > 0 {
		conditions = append(conditions, fmt.Sprintf("p.idjalurpendaftaran = $%d", argIdx))
		args = append(args, req.IDJalurPendaftaran)
		argIdx++
	}
	if req.IDGelombang > 0 {
		conditions = append(conditions, fmt.Sprintf("p.idgelombang = $%d", argIdx))
		args = append(args, req.IDGelombang)
		argIdx++
	}
	if req.OnlyAdministrasi {
		conditions = append(conditions, "p.isadministrasi = 1")
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	type candRow struct {
		IDPendaftar         string  `db:"idpendaftar"`
		Nama                string  `db:"nama"`
		JK                  string  `db:"jk"`
		HP                  string  `db:"hp"`
		Email               string  `db:"email"`
		Alamat              string  `db:"alamat"`
		IDKota              *int    `db:"idkota"`
		IDPeriode           string  `db:"id_periode"`
		IDSistemKuliah      int     `db:"id_sistem_kuliah"`
		IDJalurPendaftaran  int     `db:"id_jalur_pendaftaran"`
		IDGelombang         int     `db:"id_gelombang"`
		IDUnit1             *string `db:"id_unit1"`
		IDUnit2             *string `db:"id_unit2"`
		IsAdministrasi      int     `db:"is_administrasi"`
		NoUjianSipenmaru    string  `db:"no_ujian_sipenmaru"`
		ExistingKodePeserta string  `db:"existing_kode_peserta"`
	}

	var candidates []candRow
	err := r.siakadDB.SelectContext(ctx, &candidates, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data pendaftar: %w", err)
	}

	// Lookup already registered CBT participant codes from CBT database
	if len(candidates) > 0 {
		var existingParticipants []struct {
			IDPendaftar string `db:"idpendaftar"`
			KodePeserta string `db:"kodepeserta"`
		}
		_ = r.db.SelectContext(ctx, &existingParticipants, `
			SELECT COALESCE(idpendaftar, '') as idpendaftar, kodepeserta 
			FROM cat.at_peserta 
			WHERE idpendaftar IS NOT NULL AND idpendaftar != '' AND (softdelete = '0' OR softdelete IS NULL)
		`)
		existMap := make(map[string]string, len(existingParticipants))
		for _, ep := range existingParticipants {
			existMap[ep.IDPendaftar] = ep.KodePeserta
		}
		for i := range candidates {
			if code, ok := existMap[candidates[i].IDPendaftar]; ok {
				candidates[i].ExistingKodePeserta = code
			}
		}
	}

	selectedMap := make(map[string]bool)
	hasSelection := len(req.SelectedPendaftarIDs) > 0
	for _, id := range req.SelectedPendaftarIDs {
		selectedMap[id] = true
	}

	yearPrefix := fmt.Sprintf("%02d", time.Now().Year()%100)
	if pmbPeriodStr != "" && len(pmbPeriodStr) >= 4 {
		yearPrefix = pmbPeriodStr[2:4]
	}

	for _, c := range candidates {
		if hasSelection && !selectedMap[c.IDPendaftar] {
			continue
		}

		resp.TotalProcessed++

		var kodepeserta string
		if c.ExistingKodePeserta != "" {
			kodepeserta = c.ExistingKodePeserta
			resp.SkippedCount++
		} else {
			// Check if already in at_peserta
			_ = r.db.GetContext(ctx, &kodepeserta, "SELECT kodepeserta FROM cat.at_peserta WHERE idpendaftar = $1 AND (softdelete = '0' OR softdelete IS NULL)", c.IDPendaftar)
			if kodepeserta != "" {
				resp.SkippedCount++
			} else {
				// Generate new kodepeserta
				noujian := c.NoUjianSipenmaru
				if noujian != "" {
					if len(noujian) >= 12 {
						kodepeserta = noujian
					} else {
						kodepeserta = yearPrefix + noujian
					}
				} else {
					p1 := "00"
					if c.IDUnit1 != nil && *c.IDUnit1 != "" {
						var kd string
						_ = r.siakadDB.GetContext(ctx, &kd, "SELECT COALESCE(kodedaftar, '') FROM ref.ms_unit WHERE idunit::text = $1", *c.IDUnit1)
						if kd != "" {
							if len(kd) >= 2 {
								p1 = kd[:2]
							} else {
								p1 = fmt.Sprintf("%02s", kd)
							}
						}
					}
					p2 := "00"
					if c.IDUnit2 != nil && *c.IDUnit2 != "" {
						var kd string
						_ = r.siakadDB.GetContext(ctx, &kd, "SELECT COALESCE(kodedaftar, '') FROM ref.ms_unit WHERE idunit::text = $1", *c.IDUnit2)
						if kd != "" {
							if len(kd) >= 2 {
								p2 = kd[:2]
							} else {
								p2 = fmt.Sprintf("%02s", kd)
							}
						}
					}

					var maxSeq int
					_ = r.db.GetContext(ctx, &maxSeq, `
						SELECT COALESCE(MAX(
							CASE WHEN length(kodepeserta) >= 4 THEN SUBSTRING(kodepeserta FROM length(kodepeserta)-3 FOR 4)::int ELSE 0 END
						), 0) FROM cat.at_peserta WHERE idperiode = $1
					`, cbtPeriodID)
					seq := maxSeq + resp.ImportedCount + 1

					noujian = fmt.Sprintf("%s%s001%02d%04d", p1, p2, c.IDJalurPendaftaran, seq)
					kodepeserta = fmt.Sprintf("%s%s", yearPrefix, noujian)
				}

				if len(kodepeserta) > 20 {
					kodepeserta = kodepeserta[:20]
				}

				hasher := md5.New()
				hasher.Write([]byte(kodepeserta))
				hashedPassword := hex.EncodeToString(hasher.Sum(nil))

				_, insErr := r.db.ExecContext(ctx, `
					INSERT INTO cat.at_peserta (
						kodepeserta, idperiode, nama, jk, hp, email, alamat, idkota,
						password, hint, sumberdata, isvalid, isaktif, islogin,
						kodereferensi, idpendaftar, softdelete
					) VALUES (
						$1, $2, $3, $4, $5, $6, $7, $8,
						$9, $10, 'P', 1, 0, 1,
						$11, $12, '0'
					)
					ON CONFLICT (kodepeserta) DO UPDATE SET
						nama = EXCLUDED.nama,
						idpendaftar = EXCLUDED.idpendaftar,
						softdelete = '0'
				`, kodepeserta, cbtPeriodID, c.Nama, c.JK, c.HP, c.Email, c.Alamat, c.IDKota,
					hashedPassword, kodepeserta, c.IDPendaftar, c.IDPendaftar)
				if insErr != nil {
					resp.ErrorCount++
					resp.Errors = append(resp.Errors, fmt.Sprintf("%s (%s): %v", c.Nama, c.IDPendaftar, insErr))
					continue
				}

				if c.NoUjianSipenmaru == "" {
					_, _ = r.siakadDB.ExecContext(ctx, "UPDATE pendaftaran.pd_pendaftar SET noujian = $1 WHERE idpendaftar = $2 AND (noujian IS NULL OR noujian = '')", noujian, c.IDPendaftar)
				}

				resp.ImportedCount++
			}
		}

		// Register to exam (cat.at_pesertaujian)
		_, examErr := r.db.ExecContext(ctx, `
			INSERT INTO cat.at_pesertaujian (kodepeserta, idujian, softdelete, t_updatetime, t_updateact)
			VALUES ($1, $2, '0', NOW(), 'i-pesertaujian')
			ON CONFLICT (kodepeserta, idujian) DO UPDATE
			SET softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-pesertaujian'
		`, kodepeserta, examID)
		if examErr == nil {
			resp.ExamRegisteredCount++
		}
	}

	// Auto-plotting if requested
	if req.AutoPlotting && resp.ExamRegisteredCount > 0 {
		distResp, distErr := r.AutoDistributeParticipants(ctx, examID, false)
		if distErr == nil && distResp != nil {
			resp.SessionPlottedCount = distResp.TotalAssigned
		}
	}

	return resp, nil
}

func (r *examinationRepository) GetCompletedParticipants(ctx context.Context, examID int, scheduleID int) ([]*dto.CompletedParticipantDTO, error) {
	var whereClause string
	var args []interface{}

	if scheduleID > 0 {
		whereClause = `jp.idjadwalujian = $1 AND jp.tglselesai IS NOT NULL AND (jp.softdelete = '0' OR jp.softdelete IS NULL)`
		args = append(args, scheduleID)
	} else if examID > 0 {
		whereClause = `ju.idujian = $1 AND jp.tglselesai IS NOT NULL AND (jp.softdelete = '0' OR jp.softdelete IS NULL)`
		args = append(args, examID)
	} else {
		whereClause = `jp.tglselesai IS NOT NULL AND (jp.softdelete = '0' OR jp.softdelete IS NULL)`
	}

	query := fmt.Sprintf(`
		SELECT 
			jp.kodepeserta,
			p.nama,
			p.email,
			jp.idjadwalujian as schedule_id,
			ju.nojadwal as session_number,
			COALESCE(r.namaruang, ru.koderuang, '') as room_name,
			jp.tglmulai,
			jp.tglselesai,
			COALESCE(jp.nilai, 0.0) as final_score,
			CASE WHEN COALESCE(jp.nilai, 0.0) >= COALESCE(u.nilaiminimal, 60.0) THEN 'LULUS' ELSE 'TIDAK LULUS' END as passing_status,
			GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) as risk_score,
			CASE 
			    WHEN GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) >= 81 THEN 'CRITICAL'
			    WHEN GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) >= 61 THEN 'HIGH_RISK'
			    WHEN GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) >= 41 THEN 'SUSPICIOUS'
			    WHEN GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) >= 21 THEN 'ATTENTION'
			    ELSE 'NORMAL'
			END as risk_level,
			GREATEST(COALESCE(sec.total_tab_switches, 0), COALESCE(jp.tab_switch_count, 0)) as tab_switch_count,
			GREATEST(COALESCE(sec.total_fullscreen_exits, 0), COALESCE(jp.fullscreen_exit_count, 0)) as fullscreen_exit_count,
			COALESCE(sec.total_violations, 0) as total_violations,
			COALESCE(sec.unlock_count, 0) as unlock_count,
			COALESCE(jp.disconnect_count, 0) as disconnect_count,
			COALESCE(jp.is_locked, 0) as is_locked,
			jp.lock_reason,
			COALESCE(jp.is_verified, 0) as is_verified,
			COALESCE(ans.total_answered, 0) as total_answered,
			COALESCE(ans.correct_count, 0) as correct_count,
			COALESCE(ans.wrong_count, 0) as wrong_count,
			COALESCE(ans.empty_count, 0) as empty_count
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_peserta p ON p.kodepeserta = jp.kodepeserta
		JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian
		JOIN cat.at_ujian u ON u.idujian = ju.idujian
		LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian
		LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
		LEFT JOIN (
			SELECT a.idjadwalujian, a.kodepeserta,
			       COUNT(CASE WHEN a.jawabanpilih > 0 THEN 1 END) as total_answered,
			       COUNT(CASE WHEN a.isbenar = 1 OR a.jawabanpilih = p.jawabanbenar THEN 1 END) as correct_count,
			       COUNT(CASE WHEN a.jawabanpilih > 0 AND (a.isbenar = 0 OR a.jawabanpilih != p.jawabanbenar) THEN 1 END) as wrong_count,
			       COUNT(CASE WHEN a.jawabanpilih = 0 OR a.jawabanpilih IS NULL THEN 1 END) as empty_count
			FROM cat.at_jawabanpeserta a
			JOIN cat.at_pertanyaan p ON p.kodesoal = a.kodesoal AND p.nourut = a.nourut
			WHERE (a.softdelete = '0' OR a.softdelete IS NULL)
			GROUP BY a.idjadwalujian, a.kodepeserta
		) ans ON ans.idjadwalujian = jp.idjadwalujian AND ans.kodepeserta = jp.kodepeserta
		LEFT JOIN (
			SELECT idjadwal, kodepeserta,
			       COUNT(*) FILTER (
			           WHERE event_type IN (
			               'TAB_SWITCH', 'WINDOW_BLUR', 'FULLSCREEN_EXIT', 'FULLSCREEN_REQUIRED',
			               'CAMERA_DISCONNECTED', 'FORBIDDEN_SHORTCUT', 'SCREENSHOT_ATTEMPT',
			               'DEVTOOLS_ATTEMPT', 'DEVTOOLS_OPEN', 'PRINT_ATTEMPT', 'COPY_ATTEMPT',
			               'PASTE_ATTEMPT', 'CUT_ATTEMPT', 'MULTI_MONITOR', 'VM_DETECTED',
			               'UNTRUSTED_EVENT', 'MULTI_FACE_DETECTED', 'CAMERA_BLACKOUT'
			           ) AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       )::integer as total_violations,
			       COUNT(*) FILTER (
			           WHERE event_type IN ('TAB_SWITCH', 'WINDOW_BLUR') 
			             AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       )::integer as total_tab_switches,
			       COUNT(*) FILTER (
			           WHERE event_type IN ('FULLSCREEN_EXIT', 'FULLSCREEN_REQUIRED') 
			             AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       )::integer as total_fullscreen_exits,
			       COALESCE(SUM(risk_score) FILTER (
			           WHERE event_type IN (
			               'TAB_SWITCH', 'WINDOW_BLUR', 'FULLSCREEN_EXIT', 'FULLSCREEN_REQUIRED',
			               'CAMERA_DISCONNECTED', 'FORBIDDEN_SHORTCUT', 'SCREENSHOT_ATTEMPT',
			               'DEVTOOLS_ATTEMPT', 'DEVTOOLS_OPEN', 'PRINT_ATTEMPT', 'COPY_ATTEMPT',
			               'PASTE_ATTEMPT', 'CUT_ATTEMPT', 'MULTI_MONITOR', 'VM_DETECTED',
			               'UNTRUSTED_EVENT', 'MULTI_FACE_DETECTED', 'CAMERA_BLACKOUT'
			           ) AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       ), 0)::integer as cumulative_risk_score,
			       COUNT(*) FILTER (WHERE event_type = 'PROCTOR_UNLOCK')::integer as unlock_count
			FROM cat.at_security_events
			GROUP BY idjadwal, kodepeserta
		) sec ON sec.idjadwal = jp.idjadwalujian AND sec.kodepeserta = jp.kodepeserta
		WHERE %s
		ORDER BY jp.tglselesai DESC, jp.nilai DESC
	`, whereClause)

	type rawCompletedRow struct {
		KodePeserta         string     `db:"kodepeserta"`
		Nama                string     `db:"nama"`
		Email               *string    `db:"email"`
		ScheduleID          int        `db:"schedule_id"`
		SessionNumber       int        `db:"session_number"`
		RoomName            string     `db:"room_name"`
		TglMulai            *time.Time `db:"tglmulai"`
		TglSelesai          *time.Time `db:"tglselesai"`
		FinalScore          float64    `db:"final_score"`
		PassingStatus       string     `db:"passing_status"`
		RiskScore           int        `db:"risk_score"`
		RiskLevel           string     `db:"risk_level"`
		TabSwitchCount      int        `db:"tab_switch_count"`
		FullscreenExitCount int        `db:"fullscreen_exit_count"`
		TotalViolations     int        `db:"total_violations"`
		UnlockCount         int        `db:"unlock_count"`
		DisconnectCount     int        `db:"disconnect_count"`
		IsLocked            int        `db:"is_locked"`
		LockReason          *string    `db:"lock_reason"`
		IsVerified          int        `db:"is_verified"`
		TotalAnswered       int        `db:"total_answered"`
		CorrectCount        int        `db:"correct_count"`
		WrongCount          int        `db:"wrong_count"`
		EmptyCount          int        `db:"empty_count"`
	}

	var rows []rawCompletedRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	var results []*dto.CompletedParticipantDTO
	for _, row := range rows {
		duration := 0
		var startedStr, finishedStr *string
		if row.TglMulai != nil {
			s := row.TglMulai.Format("2006-01-02 15:04:05")
			startedStr = &s
		}
		if row.TglSelesai != nil {
			f := row.TglSelesai.Format("2006-01-02 15:04:05")
			finishedStr = &f
			if row.TglMulai != nil {
				duration = int(row.TglSelesai.Sub(*row.TglMulai).Minutes())
				if duration < 0 {
					duration = 0
				}
			}
		}

		results = append(results, &dto.CompletedParticipantDTO{
			ParticipantCode:     row.KodePeserta,
			Name:                row.Nama,
			Email:               row.Email,
			ScheduleID:          row.ScheduleID,
			SessionNumber:       row.SessionNumber,
			RoomName:            row.RoomName,
			StartedAt:           startedStr,
			FinishedAt:          finishedStr,
			DurationMinutes:     duration,
			FinalScore:          row.FinalScore,
			PassingStatus:       row.PassingStatus,
			RiskScore:           row.RiskScore,
			RiskLevel:           row.RiskLevel,
			TabSwitchCount:      row.TabSwitchCount,
			FullscreenExitCount: row.FullscreenExitCount,
			TotalViolations:     row.TotalViolations,
			UnlockCount:         row.UnlockCount,
			DisconnectCount:     row.DisconnectCount,
			IsLocked:            row.IsLocked,
			LockReason:          row.LockReason,
			TotalAnswered:       row.TotalAnswered,
			CorrectCount:        row.CorrectCount,
			WrongCount:          row.WrongCount,
			EmptyCount:          row.EmptyCount,
			IsVerified:          row.IsVerified,
		})
	}

	return results, nil
}

func (r *examinationRepository) GetParticipantCompletedExams(ctx context.Context, participantCode string) ([]*dto.ParticipantExamSummaryDTO, error) {
	query := `
		SELECT 
			jp.idjadwalujian as schedule_id,
			ju.idujian as exam_id,
			u.namaujian as exam_name,
			COALESCE(p.namaperiode, '') as period_name,
			COALESCE(ju.nojadwal, 1) as session_number,
			COALESCE(r.namaruang, ru.koderuang, 'Ruang Utama') as room_name,
			COALESCE(CAST(ru.tglmulai AS varchar), CAST(ju.tglmulai AS varchar), CAST(CURRENT_DATE AS varchar)) as exam_date,
			jp.tglmulai,
			jp.tglselesai,
			COALESCE(jp.nilai, 0.0) as final_score,
			COALESCE(u.nilaiminimal, 60.0) as passing_grade,
			CASE WHEN COALESCE(jp.nilai, 0.0) >= COALESCE(u.nilaiminimal, 60.0) THEN 'LULUS' ELSE 'TIDAK LULUS' END as passing_status,
			COALESCE(ju.tampilkannilai, 1) as tampilkan_nilai,
			GREATEST(COALESCE(sec.cumulative_risk_score, 0), COALESCE(jp.risk_score, 0)) as risk_score,
			GREATEST(COALESCE(sec.total_tab_switches, 0), COALESCE(jp.tab_switch_count, 0)) as tab_switch_count,
			COALESCE(jp.is_locked, 0) as is_locked,
			COALESCE(ans.total_answered, 0) as total_answered,
			COALESCE(ans.total_questions, 0) as total_questions
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian
		JOIN cat.at_ujian u ON u.idujian = ju.idujian
		LEFT JOIN cat.at_periode p ON p.idperiode = u.idperiode
		LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian
		LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
		LEFT JOIN (
			SELECT a.idjadwalujian, a.kodepeserta,
			       COUNT(CASE WHEN a.jawabanpilih > 0 THEN 1 END) as total_answered,
			       COUNT(*) as total_questions
			FROM cat.at_jawabanpeserta a
			WHERE (a.softdelete = '0' OR a.softdelete IS NULL)
			GROUP BY a.idjadwalujian, a.kodepeserta
		) ans ON ans.idjadwalujian = jp.idjadwalujian AND ans.kodepeserta = jp.kodepeserta
		LEFT JOIN (
			SELECT idjadwal, kodepeserta,
			       COUNT(*) FILTER (
			           WHERE event_type IN ('TAB_SWITCH', 'WINDOW_BLUR') 
			             AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       )::integer as total_tab_switches,
			       COALESCE(SUM(risk_score) FILTER (
			           WHERE event_type IN (
			               'TAB_SWITCH', 'WINDOW_BLUR', 'FULLSCREEN_EXIT', 'FULLSCREEN_REQUIRED',
			               'CAMERA_DISCONNECTED', 'FORBIDDEN_SHORTCUT', 'SCREENSHOT_ATTEMPT',
			               'DEVTOOLS_ATTEMPT', 'DEVTOOLS_OPEN', 'PRINT_ATTEMPT', 'COPY_ATTEMPT',
			               'PASTE_ATTEMPT', 'CUT_ATTEMPT', 'MULTI_MONITOR', 'VM_DETECTED',
			               'UNTRUSTED_EVENT', 'MULTI_FACE_DETECTED', 'CAMERA_BLACKOUT'
			           ) AND NOT (COALESCE(metadata->>'cooldown_suppressed', 'false')::boolean)
			       ), 0)::integer as cumulative_risk_score
			FROM cat.at_security_events
			WHERE kodepeserta = $1
			GROUP BY idjadwal, kodepeserta
		) sec ON sec.idjadwal = jp.idjadwalujian AND sec.kodepeserta = jp.kodepeserta
		WHERE jp.kodepeserta = $1
		  AND (jp.tglselesai IS NOT NULL OR jp.nilai > 0)
		  AND (jp.softdelete = '0' OR jp.softdelete IS NULL)

		UNION ALL

		SELECT 
			s.id_ep_schedule as schedule_id,
			s.id_ep_exam as exam_id,
			e.exam_name as exam_name,
			'TOEFL & TOEIC / English Proficiency' as period_name,
			1 as session_number,
			COALESCE(r.namaruang, ar.koderuang, 'Ruang Ujian ' || s.id_ruang) as room_name,
			COALESCE(s.exam_date::text, CURRENT_DATE::text) as exam_date,
			sp.started_at as tglmulai,
			sp.finished_at as tglselesai,
			COALESCE(sc.total_scaled_score, 0)::float8 as final_score,
			COALESCE(e.passing_score, 450)::float8 as passing_grade,
			COALESCE(sc.passing_status, 'SELESAI') as passing_status,
			1 as tampilkan_nilai,
			COALESCE(sp.risk_score, 0)::integer as risk_score,
			COALESCE(sp.tab_switch_count, 0)::integer as tab_switch_count,
			CASE WHEN sp.is_locked THEN 1 ELSE 0 END as is_locked,
			COALESCE((SELECT COUNT(*) FROM cat.cat_participant_answers_multi a WHERE a.id_session = sp.id_participant AND a.selected_options != '')::integer, 0) as total_answered,
			COALESCE((SELECT COUNT(*) FROM cat.ep_exam_sections es JOIN cat.cat_bank_stimulus bs ON bs.kodesoal = es.kodesoal JOIN cat.cat_stimulus_items si ON si.id_stimulus = bs.id_stimulus WHERE es.id_ep_exam = s.id_ep_exam)::integer, 0) as total_questions
		FROM cat.ep_schedule_participants sp
		JOIN cat.ep_schedules s ON s.id_ep_schedule = sp.id_ep_schedule
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		LEFT JOIN cat.at_ruangujian ar ON ar.idruangujian = s.id_ruang
		LEFT JOIN cat.at_ruang r ON r.koderuang = ar.koderuang OR r.koderuang = s.id_ruang::text
		LEFT JOIN cat.ep_scores sc ON sc.id_participant = sp.id_participant
		WHERE sp.kodepeserta = $1
		  AND (sp.session_status = 'SUBMITTED' OR sp.finished_at IS NOT NULL)
		  AND (sp.softdelete = '0' OR sp.softdelete IS NULL)

		ORDER BY tglselesai DESC NULLS LAST, tglmulai DESC NULLS LAST
	`

	type rawSummaryRow struct {
		ScheduleID     int        `db:"schedule_id"`
		ExamID         int        `db:"exam_id"`
		ExamName       string     `db:"exam_name"`
		PeriodName     string     `db:"period_name"`
		SessionNumber  int        `db:"session_number"`
		RoomName       string     `db:"room_name"`
		ExamDate       string     `db:"exam_date"`
		TglMulai       *time.Time `db:"tglmulai"`
		TglSelesai     *time.Time `db:"tglselesai"`
		FinalScore     float64    `db:"final_score"`
		PassingGrade   float64    `db:"passing_grade"`
		PassingStatus  string     `db:"passing_status"`
		TampilkanNilai int        `db:"tampilkan_nilai"`
		RiskScore      int        `db:"risk_score"`
		TabSwitchCount int        `db:"tab_switch_count"`
		IsLocked       int        `db:"is_locked"`
		TotalAnswered  int        `db:"total_answered"`
		TotalQuestions int        `db:"total_questions"`
	}

	var rows []rawSummaryRow
	if err := r.db.SelectContext(ctx, &rows, query, participantCode); err != nil {
		return nil, fmt.Errorf("gagal memuat riwayat ujian peserta: %w", err)
	}

	var results []*dto.ParticipantExamSummaryDTO
	for _, row := range rows {
		item := &dto.ParticipantExamSummaryDTO{
			ScheduleID:     row.ScheduleID,
			ExamID:         row.ExamID,
			ExamName:       row.ExamName,
			PeriodName:     row.PeriodName,
			SessionNumber:  row.SessionNumber,
			RoomName:       row.RoomName,
			ExamDate:       row.ExamDate,
			FinalScore:     row.FinalScore,
			PassingGrade:   row.PassingGrade,
			PassingStatus:  row.PassingStatus,
			TampilkanNilai: row.TampilkanNilai,
			RiskScore:      row.RiskScore,
			TabSwitchCount: row.TabSwitchCount,
			IsLocked:       row.IsLocked,
			TotalAnswered:  row.TotalAnswered,
			TotalQuestions: row.TotalQuestions,
		}
		if row.TglMulai != nil {
			s := row.TglMulai.Format("2006-01-02 15:04:05")
			item.StartedAt = &s
		}
		if row.TglSelesai != nil {
			f := row.TglSelesai.Format("2006-01-02 15:04:05")
			item.FinishedAt = &f
			if row.TglMulai != nil {
				dur := int(row.TglSelesai.Sub(*row.TglMulai).Minutes())
				if dur < 0 {
					dur = 0
				}
				item.DurationMinutes = dur
			}
		}
		results = append(results, item)
	}
	if results == nil {
		results = []*dto.ParticipantExamSummaryDTO{}
	}
	return results, nil
}

func (r *examinationRepository) GetParticipantExamResultDetail(ctx context.Context, scheduleID int, participantCode string) (*dto.ParticipantExamResultDetailDTO, error) {
	// 1. Fetch header & participant info
	headerQ := `
		SELECT 
			jp.kodepeserta,
			p.nama,
			p.email,
			jp.idjadwalujian as schedule_id,
			ju.idujian as exam_id,
			u.namaujian as exam_name,
			ju.nojadwal as session_number,
			COALESCE(r.namaruang, ru.koderuang, '') as room_name,
			jp.tglmulai,
			jp.tglselesai,
			COALESCE(jp.nilai, 0.0) as final_score,
			COALESCE(u.nilaiminimal, 60.0) as passing_grade,
			CASE WHEN COALESCE(jp.nilai, 0.0) >= COALESCE(u.nilaiminimal, 60.0) THEN 'LULUS' ELSE 'TIDAK LULUS' END as passing_status,
			COALESCE(jp.risk_score, 0) as risk_score,
			COALESCE(jp.risk_level, 'NORMAL') as risk_level,
			COALESCE(jp.is_locked, 0) as is_locked,
			jp.lock_reason,
			jp.proctor_warning,
			COALESCE(jp.tab_switch_count, 0) as tab_switch_count,
			COALESCE(jp.fullscreen_exit_count, 0) as fullscreen_exit_count,
			COALESCE(jp.disconnect_count, 0) as disconnect_count,
			COALESCE(jp.reconnect_count, 0) as reconnect_count,
			COALESCE(jp.is_verified, 0) as is_verified,
			jp.barcode_scanned_at,
			jp.client_device_id,
			jp.session_snapshot::text as session_snapshot
		FROM cat.at_jadwalpeserta jp
		JOIN cat.at_peserta p ON p.kodepeserta = jp.kodepeserta
		JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian
		JOIN cat.at_ujian u ON u.idujian = ju.idujian
		LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian
		LEFT JOIN cat.at_ruang r ON r.koderuang = ru.koderuang
		WHERE jp.idjadwalujian = $1 AND jp.kodepeserta = $2
	`

	var h struct {
		KodePeserta         string         `db:"kodepeserta"`
		Nama                string         `db:"nama"`
		Email               *string        `db:"email"`
		ScheduleID          int            `db:"schedule_id"`
		ExamID              int            `db:"exam_id"`
		ExamName            string         `db:"exam_name"`
		SessionNumber       int            `db:"session_number"`
		RoomName            string         `db:"room_name"`
		TglMulai            *time.Time     `db:"tglmulai"`
		TglSelesai          *time.Time     `db:"tglselesai"`
		FinalScore          float64        `db:"final_score"`
		PassingGrade        float64        `db:"passing_grade"`
		PassingStatus       string         `db:"passing_status"`
		RiskScore           int            `db:"risk_score"`
		RiskLevel           string         `db:"risk_level"`
		IsLocked            int            `db:"is_locked"`
		LockReason          *string        `db:"lock_reason"`
		ProctorWarning      *string        `db:"proctor_warning"`
		TabSwitchCount      int            `db:"tab_switch_count"`
		FullscreenExitCount int            `db:"fullscreen_exit_count"`
		DisconnectCount     int            `db:"disconnect_count"`
		ReconnectCount      int            `db:"reconnect_count"`
		IsVerified          int            `db:"is_verified"`
		BarcodeScannedAt    *time.Time     `db:"barcode_scanned_at"`
		ClientDeviceID      *string        `db:"client_device_id"`
		SessionSnapshot     sql.NullString `db:"session_snapshot"`
	}

	if err := r.db.GetContext(ctx, &h, headerQ, scheduleID, participantCode); err != nil {
		return r.getEPParticipantExamResultDetail(ctx, scheduleID, participantCode)
	}

	result := &dto.ParticipantExamResultDetailDTO{}
	result.Participant.ParticipantCode = h.KodePeserta
	result.Participant.Name = h.Nama
	result.Participant.Email = h.Email
	result.Participant.ScheduleID = h.ScheduleID
	result.Participant.ExamID = h.ExamID
	result.Participant.ExamName = h.ExamName
	result.Participant.SessionNumber = h.SessionNumber
	result.Participant.RoomName = h.RoomName
	result.Participant.IsVerified = h.IsVerified
	result.Participant.ClientDeviceID = h.ClientDeviceID
	result.Participant.ExamMode = "STANDARD"

	if h.TglMulai != nil {
		s := h.TglMulai.Format("2006-01-02 15:04:05")
		result.Participant.StartedAt = &s
	}
	if h.TglSelesai != nil {
		f := h.TglSelesai.Format("2006-01-02 15:04:05")
		result.Participant.FinishedAt = &f
		if h.TglMulai != nil {
			dur := int(h.TglSelesai.Sub(*h.TglMulai).Minutes())
			if dur < 0 {
				dur = 0
			}
			result.Participant.DurationMinutes = dur
		}
	}
	if h.BarcodeScannedAt != nil {
		b := h.BarcodeScannedAt.Format("2006-01-02 15:04:05")
		result.Participant.BarcodeScannedAt = &b
	}

	result.Score.FinalScore = h.FinalScore
	result.Score.PassingGrade = h.PassingGrade
	result.Score.PassingStatus = h.PassingStatus
	result.Score.SubmitReason = "MANUAL_SUBMIT"

	result.Violations.RiskScore = h.RiskScore
	result.Violations.RiskLevel = h.RiskLevel
	result.Violations.IsLocked = h.IsLocked
	result.Violations.LockReason = h.LockReason
	result.Violations.ProctorWarning = h.ProctorWarning
	result.Violations.TabSwitchCount = h.TabSwitchCount
	result.Violations.FullscreenExitCount = h.FullscreenExitCount
	result.Violations.DisconnectCount = h.DisconnectCount
	result.Violations.ReconnectCount = h.ReconnectCount

	// 2. Fetch security audit events
	type rawSecEvent struct {
		ID              int64     `db:"id"`
		ScheduleID      int       `db:"schedule_id"`
		ParticipantCode string    `db:"participant_code"`
		EventType       string    `db:"event_type"`
		Severity        string    `db:"severity"`
		RiskScore       int       `db:"risk_score"`
		Metadata        *string   `db:"metadata"`
		IPAddress       string    `db:"ip_address"`
		CreatedAt       time.Time `db:"created_at"`
	}
	var rawEvents []rawSecEvent
	_ = r.db.SelectContext(ctx, &rawEvents, `
		SELECT id, idjadwal as schedule_id, kodepeserta as participant_code,
		       event_type, severity, risk_score, metadata::text as metadata,
		       COALESCE(ip_address, '') as ip_address, created_at
		FROM cat.at_security_events
		WHERE idjadwal = $1 AND kodepeserta = $2
		ORDER BY created_at ASC
	`, scheduleID, participantCode)

	// Classify submit-type events (NOT violations) separately from security violations
	submitEventTypes := map[string]bool{
		"MANUAL_SUBMIT": true, "PARTICIPANT_TIMEOUT": true, "FORCE_SUBMIT": true,
		"DEADLINE_SUBMIT": true, "EXAM_DEADLINE": true,
	}
	violationTypeMap := map[string]bool{
		"TAB_SWITCH": true, "WINDOW_BLUR": true, "FULLSCREEN_EXIT": true, "FULLSCREEN_REQUIRED": true,
		"CAMERA_DISCONNECTED": true, "FORBIDDEN_SHORTCUT": true, "SCREENSHOT_ATTEMPT": true,
		"DEVTOOLS_ATTEMPT": true, "DEVTOOLS_OPEN": true, "PRINT_ATTEMPT": true, "COPY_ATTEMPT": true,
		"PASTE_ATTEMPT": true, "CUT_ATTEMPT": true, "MULTI_MONITOR": true, "VM_DETECTED": true,
		"UNTRUSTED_EVENT": true, "MULTI_FACE_DETECTED": true, "CAMERA_BLACKOUT": true,
	}

	cumRiskScore := 0
	totTabSwitches := 0
	totFsExits := 0
	totViolations := 0
	unlockCount := 0

	var evList = make([]dto.SecurityEventItemDTO, 0)
	for _, rev := range rawEvents {
		var metaMap map[string]interface{}
		if rev.Metadata != nil && *rev.Metadata != "" {
			_ = json.Unmarshal([]byte(*rev.Metadata), &metaMap)
		}

		if rev.EventType == "PROCTOR_UNLOCK" {
			unlockCount++
		}

		isSuppressed := false
		if metaMap != nil {
			if supp, ok := metaMap["cooldown_suppressed"].(bool); ok && supp {
				isSuppressed = true
			}
		}

		if violationTypeMap[rev.EventType] && !isSuppressed {
			totViolations++
			cumRiskScore += rev.RiskScore
			if rev.EventType == "TAB_SWITCH" || rev.EventType == "WINDOW_BLUR" {
				totTabSwitches++
			}
			if rev.EventType == "FULLSCREEN_EXIT" || rev.EventType == "FULLSCREEN_REQUIRED" {
				totFsExits++
			}
		}

		// Extract submit reason from submit-type events but do NOT include them in violations list
		if submitEventTypes[rev.EventType] {
			result.Score.SubmitReason = rev.EventType
			continue // Skip: submit events are not security violations
		}
		evList = append(evList, dto.SecurityEventItemDTO{
			ID:              rev.ID,
			ScheduleID:      rev.ScheduleID,
			ParticipantCode: rev.ParticipantCode,
			ParticipantName: h.Nama,
			EventType:       rev.EventType,
			Severity:        rev.Severity,
			RiskScore:       rev.RiskScore,
			Metadata:        metaMap,
			IPAddress:       rev.IPAddress,
			CreatedAt:       rev.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	result.Violations.Events = evList

	// Apply cumulative violation counts
	finalRisk := cumRiskScore
	if h.RiskScore > finalRisk {
		finalRisk = h.RiskScore
	}
	finalRiskLevel := "NORMAL"
	if finalRisk >= 81 {
		finalRiskLevel = "CRITICAL"
	} else if finalRisk >= 61 {
		finalRiskLevel = "HIGH_RISK"
	} else if finalRisk >= 41 {
		finalRiskLevel = "SUSPICIOUS"
	} else if finalRisk >= 21 {
		finalRiskLevel = "ATTENTION"
	}

	result.Violations.RiskScore = finalRisk
	result.Violations.RiskLevel = finalRiskLevel
	if totTabSwitches > h.TabSwitchCount {
		result.Violations.TabSwitchCount = totTabSwitches
	} else {
		result.Violations.TabSwitchCount = h.TabSwitchCount
	}
	if totFsExits > h.FullscreenExitCount {
		result.Violations.FullscreenExitCount = totFsExits
	} else {
		result.Violations.FullscreenExitCount = h.FullscreenExitCount
	}
	result.Violations.ActiveTabSwitchCount = h.TabSwitchCount
	result.Violations.ActiveFullscreenExitCount = h.FullscreenExitCount
	result.Violations.TotalViolations = totViolations
	result.Violations.UnlockCount = unlockCount

	// 3. Priority: If participant has an immutable SessionSnapshot in cat.at_jadwalpeserta, use it directly!
	if h.SessionSnapshot.Valid && len(h.SessionSnapshot.String) > 10 {
		var stored StoredSessionSnapshot
		if err := json.Unmarshal([]byte(h.SessionSnapshot.String), &stored); err == nil && len(stored.Questions) > 0 {
			result.Participant.ExamMode = "STANDARD"

			type ansRow struct {
				NoSoal     int  `db:"nosoal"`
				Jawaban    int  `db:"jawaban"`
				IsDoubtful bool `db:"is_doubtful"`
			}
			var userAnswers []ansRow
			_ = r.db.SelectContext(ctx, &userAnswers, `
				SELECT nourutpeserta::integer as nosoal, 
				       COALESCE(jawabanpilih::integer, 0) as jawaban,
				       COALESCE((snapshot_metadata->>'is_doubtful')::boolean, false) as is_doubtful
				FROM cat.at_jawabanpeserta
				WHERE idjadwalujian = $1 AND kodepeserta = $2 AND (softdelete = '0' OR softdelete IS NULL)
			`, scheduleID, participantCode)

			ansMap := make(map[int]ansRow)
			for _, a := range userAnswers {
				ansMap[a.NoSoal] = a
			}

			letterMap := map[int]string{1: "A", 2: "B", 3: "C", 4: "D", 5: "E"}
			correctCount := 0
			wrongCount := 0
			emptyCount := 0
			totalWeight := 0.0
			earnedWeight := 0.0

			for _, q := range stored.Questions {
				correctKey := letterMap[q.CorrectOptionKey]
				if correctKey == "" {
					correctKey = fmt.Sprintf("%d", q.CorrectOptionKey)
				}

				ans, answered := ansMap[q.QuestionNumber]
				userKey := ""
				if answered && ans.Jawaban > 0 {
					userKey = letterMap[ans.Jawaban]
					if userKey == "" {
						userKey = fmt.Sprintf("%d", ans.Jawaban)
					}
				}

				wCorr := q.WeightCorrect
				if wCorr <= 0 {
					wCorr = 1.0
				}
				wWrng := q.WeightWrong
				totalWeight += wCorr

				status := "EMPTY"
				score := 0.0
				if !answered || ans.Jawaban == 0 {
					status = "EMPTY"
					emptyCount++
				} else if ans.Jawaban == q.CorrectOptionKey {
					status = "CORRECT"
					score = wCorr
					correctCount++
					earnedWeight += wCorr
				} else {
					status = "WRONG"
					score = wWrng
					wrongCount++
					if wWrng != 0 {
						earnedWeight += wWrng
					}
				}

				var optsReview []dto.QuestionOptionReviewDTO
				for _, op := range q.Options {
					optLetter := letterMap[op.OptionKey]
					if optLetter == "" {
						optLetter = fmt.Sprintf("%d", op.OptionKey)
					}
					isChosen := answered && ans.Jawaban == op.OptionKey
					isCorrect := op.OptionKey == q.CorrectOptionKey
					optsReview = append(optsReview, dto.QuestionOptionReviewDTO{
						OptionKey:    optLetter,
						OptionText:   op.OptionText,
						MediaURL:     op.MediaURL,
						IsUserChoice: isChosen,
						IsCorrect:    isCorrect,
					})
				}

				qMediaType := q.QuestionMediaType
				if qMediaType == "" {
					qMediaType = "NONE"
				}

				result.Questions = append(result.Questions, dto.QuestionAnswerReviewItemDTO{
					QuestionNumber:    q.QuestionNumber,
					StimulusTitle:     q.StimulusTitle,
					StimulusText:      q.StimulusText,
					StimulusMediaType: &qMediaType,
					StimulusMediaURL:  q.StimulusMediaURL,
					QuestionText:      q.QuestionText,
					QuestionMediaType: &qMediaType,
					QuestionMediaURL:  q.QuestionMediaURL,
					Options:           optsReview,
					UserSelectedKey:   userKey,
					CorrectKey:        correctKey,
					Status:            status,
					IsDoubtful:        ans.IsDoubtful,
					WeightCorrect:     wCorr,
					WeightWrong:       wWrng,
					ScoreObtained:     score,
				})
			}

			result.Score.TotalQuestions = len(result.Questions)
			result.Score.CorrectCount = correctCount
			result.Score.WrongCount = wrongCount
			result.Score.EmptyCount = emptyCount
			result.Score.TotalWeight = totalWeight
			result.Score.EarnedWeight = math.Max(0, earnedWeight)
			if result.Score.TotalQuestions > 0 {
				result.Score.AccuracyPercent = math.Round((float64(correctCount)/float64(result.Score.TotalQuestions))*1000) / 10
			}

			// Self-Healing Calculation: If finalScore is 0 or unsubmitted, calculate dynamically from earnedWeight/totalWeight
			calculatedScore := 0.0
			if totalWeight > 0 {
				calculatedScore = math.Round((math.Max(0, earnedWeight)/totalWeight)*1000) / 10
			} else if result.Score.TotalQuestions > 0 {
				calculatedScore = math.Round((float64(correctCount)/float64(result.Score.TotalQuestions))*1000) / 10
			}

			if (result.Score.FinalScore == 0.0 && calculatedScore > 0) || h.TglSelesai == nil {
				result.Score.FinalScore = calculatedScore
				// Persist healed score & completion time to database to prevent stale zero score
				_, _ = r.db.ExecContext(ctx, `
					UPDATE cat.at_jadwalpeserta 
					SET nilai = $1, 
					    tglselesai = COALESCE(tglselesai, NOW()),
					    submit_reason = COALESCE(submit_reason, 'AUTO_HEAL_SUBMIT')
					WHERE idjadwalujian = $2 AND kodepeserta = $3
				`, calculatedScore, scheduleID, participantCode)
			}

			if result.Score.FinalScore >= result.Score.PassingGrade {
				result.Score.PassingStatus = "LULUS"
			} else {
				result.Score.PassingStatus = "TIDAK LULUS"
			}

			return result, nil
		}
	}

	// 4. Check for legacy Dynamic exam session
	var dynamicSess struct {
		IDSession       int64          `db:"id_session"`
		SessionSnapshot sql.NullString `db:"session_snapshot"`
		SubmitReason    *string        `db:"submit_reason"`
	}
	err := r.db.GetContext(ctx, &dynamicSess, `
		SELECT id_session, session_snapshot::text as session_snapshot, submit_reason
		FROM cat.cat_participant_sessions_ext
		WHERE idjadwalujian = $1 AND kodepeserta = $2
	`, scheduleID, participantCode)

	if err == nil && dynamicSess.SessionSnapshot.Valid && dynamicSess.SessionSnapshot.String != "" {
		result.Participant.ExamMode = "DYNAMIC"
		if dynamicSess.SubmitReason != nil && *dynamicSess.SubmitReason != "" {
			result.Score.SubmitReason = *dynamicSess.SubmitReason
		}

		var stimuli []dto.DynamicStimulusUnitDTO
		_ = json.Unmarshal([]byte(dynamicSess.SessionSnapshot.String), &stimuli)

		// Fetch dynamic answers
		type dynAnsRow struct {
			IDStimulus     int64   `db:"id_stimulus"`
			IDItem         int64   `db:"id_item"`
			SelectedOption *string `db:"selected_options"`
			IsDoubtful     bool    `db:"is_doubtful"`
			IsCorrect      *bool   `db:"is_correct"`
			AwardedPoints  float64 `db:"awarded_points"`
		}
		var userAnswers []dynAnsRow
		_ = r.db.SelectContext(ctx, &userAnswers, `
			SELECT id_stimulus, id_item, selected_options, is_doubtful, is_correct, awarded_points
			FROM cat.cat_participant_answers_multi
			WHERE id_session = $1
		`, dynamicSess.IDSession)

		ansMap := make(map[int64]dynAnsRow)
		for _, a := range userAnswers {
			ansMap[a.IDItem] = a
		}

		qNo := 1
		correctCount := 0
		wrongCount := 0
		emptyCount := 0

		for _, st := range stimuli {
			for _, item := range st.SubItems {
				qItem := dto.QuestionAnswerReviewItemDTO{
					QuestionNumber:    qNo,
					StimulusTitle:     &st.Title,
					StimulusText:      &st.NarrativeText,
					StimulusMediaType: &st.MediaType,
					StimulusMediaURL:  st.MediaURL,
					QuestionText:      item.QuestionText,
					QuestionMediaType: &item.QuestionMediaType,
					QuestionMediaURL:  item.QuestionMediaURL,
					WeightCorrect:     item.WeightCorrect,
					WeightWrong:       item.WeightWrong,
					Options:           []dto.QuestionOptionReviewDTO{},
				}

				ans, hasAns := ansMap[int64(item.IDItem)]
				userSel := ""
				if hasAns && ans.SelectedOption != nil {
					userSel = strings.ToUpper(strings.TrimSpace(*ans.SelectedOption))
				}
				qItem.UserSelectedKey = userSel
				qItem.IsDoubtful = hasAns && ans.IsDoubtful

				for _, opt := range item.Options {
					label := strings.ToUpper(strings.TrimSpace(opt.OptionLabel))
					isChosen := userSel != "" && userSel == label
					// Check correctness
					isOptCorrect := false
					if hasAns && ans.IsCorrect != nil && *ans.IsCorrect && isChosen {
						isOptCorrect = true
					}
					qItem.Options = append(qItem.Options, dto.QuestionOptionReviewDTO{
						OptionKey:    label,
						OptionText:   opt.OptionText,
						MediaURL:     opt.MediaURL,
						IsUserChoice: isChosen,
						IsCorrect:    isOptCorrect,
					})
				}

				if hasAns && ans.IsCorrect != nil {
					if *ans.IsCorrect {
						qItem.Status = "CORRECT"
						qItem.ScoreObtained = ans.AwardedPoints
						correctCount++
					} else if userSel == "" {
						qItem.Status = "EMPTY"
						qItem.ScoreObtained = 0
						emptyCount++
					} else {
						qItem.Status = "WRONG"
						qItem.ScoreObtained = ans.AwardedPoints
						wrongCount++
					}
				} else if userSel == "" {
					qItem.Status = "EMPTY"
					emptyCount++
				} else {
					qItem.Status = "WRONG"
					wrongCount++
				}

				result.Questions = append(result.Questions, qItem)
				qNo++
			}
		}

		result.Score.TotalQuestions = len(result.Questions)
		result.Score.CorrectCount = correctCount
		result.Score.WrongCount = wrongCount
		result.Score.EmptyCount = emptyCount
		if result.Score.TotalQuestions > 0 {
			result.Score.AccuracyPercent = math.Round((float64(correctCount)/float64(result.Score.TotalQuestions))*1000) / 10
		}

		return result, nil
	}

	// 4. Standard Exam: Fetch question bank and questions
	var bankCode string
	_ = r.db.GetContext(ctx, &bankCode, `SELECT COALESCE(kodesoal, 'SOAL_2026') FROM cat.at_soalujian WHERE idjadwalujian = $1 ORDER BY nourut ASC LIMIT 1`, scheduleID)
	if bankCode == "" {
		bankCode = "SOAL_2026"
	}

	type rawQuestionRow struct {
		NoUrut          int     `db:"nourut"`
		Pertanyaan      string  `db:"pertanyaan"`
		Pertanyaan2     string  `db:"pertanyaan2"`
		PertanyaanImage *string `db:"pertanyaanimage"`
		PertanyaanAudio *string `db:"pertanyaanaudio"`
		PertanyaanVideo *string `db:"pertanyaanvideo"`
		Jawaban1        string  `db:"jawaban1"`
		Jawaban1Image   *string `db:"jawaban1image"`
		Jawaban2        string  `db:"jawaban2"`
		Jawaban2Image   *string `db:"jawaban2image"`
		Jawaban3        string  `db:"jawaban3"`
		Jawaban3Image   *string `db:"jawaban3image"`
		Jawaban4        string  `db:"jawaban4"`
		Jawaban4Image   *string `db:"jawaban4image"`
		Jawaban5        string  `db:"jawaban5"`
		Jawaban5Image   *string `db:"jawaban5image"`
		JawabanBenar    int     `db:"jawabanbenar"`
		BobotBenar      float64 `db:"bobot_benar"`
		BobotSalah      float64 `db:"bobot_salah"`
		JawabanPilih    int     `db:"jawabanpilih"`
		IsBenar         int     `db:"isbenar"`
		Nilai           float64 `db:"nilai"`
	}

	rawQ := `
		SELECT 
			p.nourut,
			COALESCE(p.pertanyaan, '') as pertanyaan,
			COALESCE(p.pertanyaan2, '') as pertanyaan2,
			p.pertanyaanimage, p.pertanyaanaudio, p.pertanyaanvideo,
			COALESCE(p.jawaban1, '') as jawaban1, p.jawaban1image,
			COALESCE(p.jawaban2, '') as jawaban2, p.jawaban2image,
			COALESCE(p.jawaban3, '') as jawaban3, p.jawaban3image,
			COALESCE(p.jawaban4, '') as jawaban4, p.jawaban4image,
			COALESCE(p.jawaban5, '') as jawaban5, p.jawaban5image,
			COALESCE(p.jawabanbenar, 1) as jawabanbenar,
			COALESCE(p.bobot_benar, 1.0) as bobot_benar,
			COALESCE(p.bobot_salah, 0.0) as bobot_salah,
			COALESCE(a.jawabanpilih, 0) as jawabanpilih,
			COALESCE(a.isbenar, 0) as isbenar,
			COALESCE(a.nilai, 0.0) as nilai
		FROM cat.at_pertanyaan p
		LEFT JOIN cat.at_jawabanpeserta a ON a.kodesoal = p.kodesoal AND a.nourut = p.nourut AND a.idjadwalujian = $1 AND a.kodepeserta = $2
		WHERE p.kodesoal = $3 AND (p.softdelete = '0' OR p.softdelete IS NULL)
		ORDER BY p.nourut ASC
	`

	var qRows []rawQuestionRow
	_ = r.db.SelectContext(ctx, &qRows, rawQ, scheduleID, participantCode, bankCode)

	letterMap := map[int]string{1: "A", 2: "B", 3: "C", 4: "D", 5: "E"}
	correctCount := 0
	wrongCount := 0
	emptyCount := 0

	for _, qr := range qRows {
		correctKey := letterMap[qr.JawabanBenar]
		userKey := ""
		if qr.JawabanPilih > 0 {
			userKey = letterMap[qr.JawabanPilih]
		}

		status := "EMPTY"
		score := 0.0
		if qr.JawabanPilih == 0 {
			status = "EMPTY"
			emptyCount++
		} else if qr.JawabanPilih == qr.JawabanBenar || qr.IsBenar == 1 {
			status = "CORRECT"
			score = qr.BobotBenar
			if score == 0 {
				score = 1.0
			}
			correctCount++
		} else {
			status = "WRONG"
			score = qr.BobotSalah
			wrongCount++
		}

		options := []dto.QuestionOptionReviewDTO{
			{OptionKey: "A", OptionText: qr.Jawaban1, MediaURL: qr.Jawaban1Image, IsUserChoice: qr.JawabanPilih == 1, IsCorrect: qr.JawabanBenar == 1},
			{OptionKey: "B", OptionText: qr.Jawaban2, MediaURL: qr.Jawaban2Image, IsUserChoice: qr.JawabanPilih == 2, IsCorrect: qr.JawabanBenar == 2},
			{OptionKey: "C", OptionText: qr.Jawaban3, MediaURL: qr.Jawaban3Image, IsUserChoice: qr.JawabanPilih == 3, IsCorrect: qr.JawabanBenar == 3},
			{OptionKey: "D", OptionText: qr.Jawaban4, MediaURL: qr.Jawaban4Image, IsUserChoice: qr.JawabanPilih == 4, IsCorrect: qr.JawabanBenar == 4},
			{OptionKey: "E", OptionText: qr.Jawaban5, MediaURL: qr.Jawaban5Image, IsUserChoice: qr.JawabanPilih == 5, IsCorrect: qr.JawabanBenar == 5},
		}

		mediaType := "NONE"
		mediaURL := ""
		if qr.PertanyaanImage != nil && *qr.PertanyaanImage != "" {
			mediaType = "IMAGE"
			mediaURL = *qr.PertanyaanImage
		} else if qr.PertanyaanAudio != nil && *qr.PertanyaanAudio != "" {
			mediaType = "AUDIO"
			mediaURL = *qr.PertanyaanAudio
		} else if qr.PertanyaanVideo != nil && *qr.PertanyaanVideo != "" {
			mediaType = "VIDEO"
			mediaURL = *qr.PertanyaanVideo
		}

		var mURLPtr *string
		if mediaURL != "" {
			mURLPtr = &mediaURL
		}

		result.Questions = append(result.Questions, dto.QuestionAnswerReviewItemDTO{
			QuestionNumber:    qr.NoUrut,
			QuestionText:      qr.Pertanyaan,
			QuestionMediaType: &mediaType,
			QuestionMediaURL:  mURLPtr,
			Options:           options,
			UserSelectedKey:   userKey,
			CorrectKey:        correctKey,
			Status:            status,
			IsDoubtful:        false,
			WeightCorrect:     qr.BobotBenar,
			WeightWrong:       qr.BobotSalah,
			ScoreObtained:     score,
		})
	}

	result.Score.TotalQuestions = len(result.Questions)
	result.Score.CorrectCount = correctCount
	result.Score.WrongCount = wrongCount
	result.Score.EmptyCount = emptyCount
	if result.Score.TotalQuestions > 0 {
		result.Score.AccuracyPercent = math.Round((float64(correctCount)/float64(result.Score.TotalQuestions))*1000) / 10
	}

	return result, nil
}

func (r *examinationRepository) getEPParticipantExamResultDetail(ctx context.Context, scheduleID int, participantCode string) (*dto.ParticipantExamResultDetailDTO, error) {
	var h struct {
		IDParticipant       int        `db:"id_participant"`
		ScheduleID          int        `db:"schedule_id"`
		ExamID              int        `db:"exam_id"`
		ExamName            string     `db:"exam_name"`
		PassingScore        int        `db:"passing_score"`
		IDProfile           *int       `db:"id_profile"`
		IDExamType          int        `db:"id_exam_type"`
		RoomName            string     `db:"room_name"`
		KodePeserta         string     `db:"kodepeserta"`
		Nama                string     `db:"nama"`
		Email               *string    `db:"email"`
		StartedAt           *time.Time `db:"started_at"`
		FinishedAt          *time.Time `db:"finished_at"`
		SessionStatus       string     `db:"session_status"`
		IsLocked            bool       `db:"is_locked"`
		LockReason          *string    `db:"lock_reason"`
		ProctorWarning      *string    `db:"proctor_warning"`
		RiskScore           int        `db:"risk_score"`
		RiskLevel           string     `db:"risk_level"`
		TabSwitchCount      int        `db:"tab_switch_count"`
		FullscreenExitCount int        `db:"fullscreen_exit_count"`
		TotalViolations     int        `db:"total_violations"`
		UnlockCount         int        `db:"unlock_count"`
	}

	headerQ := `
		SELECT 
			sp.id_participant,
			sp.id_ep_schedule as schedule_id,
			e.id_ep_exam as exam_id,
			e.exam_name,
			e.passing_score,
			e.id_profile,
			e.id_exam_type,
			COALESCE(r.namaruang, ar.koderuang, 'Ruang Ujian ' || s.id_ruang) as room_name,
			sp.kodepeserta,
			COALESCE(p.nama, 'Peserta ' || sp.kodepeserta) as nama,
			p.email,
			sp.started_at,
			sp.finished_at,
			sp.session_status,
			sp.is_locked,
			sp.lock_reason,
			sp.proctor_warning,
			sp.risk_score,
			sp.risk_level,
			sp.tab_switch_count,
			sp.fullscreen_exit_count,
			sp.total_violations,
			sp.unlock_count
		FROM cat.ep_schedule_participants sp
		JOIN cat.ep_schedules s ON s.id_ep_schedule = sp.id_ep_schedule
		JOIN cat.ep_exams e ON e.id_ep_exam = s.id_ep_exam
		LEFT JOIN cat.at_ruangujian ar ON ar.idruangujian = s.id_ruang
		LEFT JOIN cat.at_ruang r ON r.koderuang = ar.koderuang OR r.koderuang = s.id_ruang::text
		LEFT JOIN cat.at_peserta p ON p.kodepeserta = sp.kodepeserta
		WHERE sp.id_ep_schedule = $1 AND sp.kodepeserta = $2
	`

	if err := r.db.GetContext(ctx, &h, headerQ, scheduleID, participantCode); err != nil {
		return nil, fmt.Errorf("data detail laporan untuk sesi ini tidak ditemukan atau belum diselesaikan: %w", err)
	}

	result := &dto.ParticipantExamResultDetailDTO{}
	result.Participant.ParticipantCode = h.KodePeserta
	result.Participant.Name = h.Nama
	result.Participant.Email = h.Email
	result.Participant.ScheduleID = h.ScheduleID
	result.Participant.ExamID = h.ExamID
	result.Participant.ExamName = h.ExamName
	result.Participant.SessionNumber = 1
	result.Participant.RoomName = h.RoomName
	result.Participant.IsVerified = 1
	result.Participant.ExamMode = "DYNAMIC"

	if h.StartedAt != nil {
		s := h.StartedAt.Format("2006-01-02 15:04:05")
		result.Participant.StartedAt = &s
	}
	if h.FinishedAt != nil {
		f := h.FinishedAt.Format("2006-01-02 15:04:05")
		result.Participant.FinishedAt = &f
		if h.StartedAt != nil {
			dur := int(h.FinishedAt.Sub(*h.StartedAt).Minutes())
			if dur < 0 {
				dur = 0
			}
			result.Participant.DurationMinutes = dur
		}
	}

	// Record Pelanggaran Peserta
	result.Violations.RiskScore = h.RiskScore
	result.Violations.RiskLevel = h.RiskLevel
	if h.IsLocked {
		result.Violations.IsLocked = 1
	}
	result.Violations.LockReason = h.LockReason
	result.Violations.ProctorWarning = h.ProctorWarning
	result.Violations.TabSwitchCount = h.TabSwitchCount
	result.Violations.FullscreenExitCount = h.FullscreenExitCount
	result.Violations.ActiveTabSwitchCount = h.TabSwitchCount
	result.Violations.ActiveFullscreenExitCount = h.FullscreenExitCount
	result.Violations.TotalViolations = h.TotalViolations
	result.Violations.UnlockCount = h.UnlockCount

	// Fetch security events
	type secEvtRow struct {
		EventType string    `db:"event_type"`
		Severity  string    `db:"severity"`
		RiskScore int       `db:"risk_score"`
		Metadata  *string   `db:"metadata"`
		IPAddress *string   `db:"ip_address"`
		DeviceID  *string   `db:"device_id"`
		CreatedAt time.Time `db:"created_at"`
	}
	var secEvts []secEvtRow
	secEvtQ := `
		SELECT event_type, severity, risk_score, metadata, ip_address, device_id, created_at
		FROM cat.ep_security_events
		WHERE id_ep_schedule = $1 AND kodepeserta = $2
		ORDER BY created_at ASC
	`
	_ = r.db.SelectContext(ctx, &secEvts, secEvtQ, scheduleID, participantCode)
	for _, e := range secEvts {
		var metaMap map[string]interface{}
		if e.Metadata != nil && *e.Metadata != "" {
			_ = json.Unmarshal([]byte(*e.Metadata), &metaMap)
		}
		ipStr := ""
		if e.IPAddress != nil {
			ipStr = *e.IPAddress
		}
		result.Violations.Events = append(result.Violations.Events, dto.SecurityEventItemDTO{
			ScheduleID:      scheduleID,
			ParticipantCode: participantCode,
			ParticipantName: h.Nama,
			EventType:       e.EventType,
			Severity:        e.Severity,
			RiskScore:       e.RiskScore,
			Metadata:        metaMap,
			IPAddress:       ipStr,
			CreatedAt:       e.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// Fetch all exam sections
	type secRow struct {
		IDSection    int    `db:"id_section"`
		KodeSoal     string `db:"kodesoal"`
		SectionOrder int    `db:"section_order"`
	}
	var sections []secRow
	_ = r.db.SelectContext(ctx, &sections, `
		SELECT id_section, kodesoal, section_order
		FROM cat.ep_exam_sections
		WHERE id_ep_exam = $1
		ORDER BY section_order ASC
	`, h.ExamID)

	// Fetch all participant answers in this session
	type ansRow struct {
		IDItem          int64  `db:"id_item"`
		SelectedOptions string `db:"selected_options"`
		IsDoubtful      bool   `db:"is_doubtful"`
	}
	var answers []ansRow
	_ = r.db.SelectContext(ctx, &answers, `
		SELECT id_item, selected_options, is_doubtful
		FROM cat.cat_participant_answers_multi
		WHERE id_session = $1
	`, h.IDParticipant)
	ansMap := make(map[int64]ansRow)
	for _, a := range answers {
		ansMap[a.IDItem] = a
	}

	totalQuestions := 0
	correctCount := 0
	wrongCount := 0
	emptyCount := 0
	totalWeight := 0.0
	earnedWeight := 0.0

	// Count per section for TOEFL score calculation fallback
	rawCountBySection := make(map[int]int)

	qNumber := 1

	for _, sec := range sections {
		// Stimuli for this section
		type stimRow struct {
			IDStimulus    int64   `db:"id_stimulus"`
			StimulusCode  string  `db:"stimulus_code"`
			Title         string  `db:"title"`
			NarrativeText string  `db:"narrative_text"`
			MediaType     string  `db:"media_type"`
			MediaURL      *string `db:"media_url"`
		}
		var stimuli []stimRow
		_ = r.db.SelectContext(ctx, &stimuli, `
			SELECT id_stimulus, stimulus_code, title, narrative_text, media_type, media_url
			FROM cat.cat_bank_stimulus
			WHERE kodesoal = $1 AND (softdelete = '0' OR softdelete IS NULL)
			ORDER BY sort_order ASC, id_stimulus ASC
		`, sec.KodeSoal)

		for _, s := range stimuli {
			type itemRow struct {
				IDItem            int64   `db:"id_item"`
				ItemOrder         int     `db:"item_order"`
				QuestionText      string  `db:"question_text"`
				QuestionMediaType string  `db:"question_media_type"`
				QuestionMediaURL  *string `db:"question_media_url"`
				CorrectAnswer     string  `db:"correct_answer"`
				WeightCorrect     float64 `db:"weight_correct"`
				WeightWrong       float64 `db:"weight_wrong"`
			}
			var items []itemRow
			_ = r.db.SelectContext(ctx, &items, `
				SELECT id_item, item_order, question_text, question_media_type, question_media_url, correct_answer, weight_correct, weight_wrong
				FROM cat.cat_stimulus_items
				WHERE id_stimulus = $1 AND (softdelete = '0' OR softdelete IS NULL)
				ORDER BY item_order ASC, id_item ASC
			`, s.IDStimulus)

			for _, it := range items {
				totalQuestions++
				totalWeight += it.WeightCorrect

				// Options
				type optRow struct {
					IDOption    int64   `db:"id_option"`
					OptionLabel string  `db:"option_label"`
					OptionText  string  `db:"option_text"`
					MediaType   string  `db:"media_type"`
					MediaURL    *string `db:"media_url"`
				}
				var opts []optRow
				_ = r.db.SelectContext(ctx, &opts, `
					SELECT id_option, option_label, option_text, media_type, media_url
					FROM cat.cat_item_options
					WHERE id_item = $1
					ORDER BY sort_order ASC, option_label ASC
				`, it.IDItem)

				userChoice := ""
				isDoubtful := false
				if a, ok := ansMap[it.IDItem]; ok {
					userChoice = strings.TrimSpace(a.SelectedOptions)
					isDoubtful = a.IsDoubtful
				}

				status := "EMPTY"
				scoreObtained := 0.0
				corr := strings.TrimSpace(it.CorrectAnswer)

				if userChoice == "" {
					emptyCount++
				} else if strings.EqualFold(userChoice, corr) {
					status = "CORRECT"
					correctCount++
					scoreObtained = it.WeightCorrect
					earnedWeight += it.WeightCorrect
					rawCountBySection[sec.SectionOrder]++
				} else {
					status = "WRONG"
					wrongCount++
					scoreObtained = -it.WeightWrong
					earnedWeight -= it.WeightWrong
				}

				var optDTOs []dto.QuestionOptionReviewDTO
				for _, o := range opts {
					isUser := strings.EqualFold(strings.TrimSpace(o.OptionLabel), userChoice)
					isCorr := strings.EqualFold(strings.TrimSpace(o.OptionLabel), corr)
					var mURL *string
					if o.MediaURL != nil && *o.MediaURL != "" {
						mURL = o.MediaURL
					}
					optDTOs = append(optDTOs, dto.QuestionOptionReviewDTO{
						OptionKey:    o.OptionLabel,
						OptionText:   o.OptionText,
						MediaURL:     mURL,
						IsUserChoice: isUser,
						IsCorrect:    isCorr,
					})
				}

				var sTitle *string
				if s.Title != "" {
					sTitle = &s.Title
				}
				var sText *string
				if s.NarrativeText != "" {
					sText = &s.NarrativeText
				}
				var sMediaType *string
				if s.MediaType != "" {
					sMediaType = &s.MediaType
				}
				var sMediaURL *string
				if s.MediaURL != nil && *s.MediaURL != "" {
					sMediaURL = s.MediaURL
				}

				var qMediaType *string
				if it.QuestionMediaType != "" {
					qMediaType = &it.QuestionMediaType
				}
				var qMediaURL *string
				if it.QuestionMediaURL != nil && *it.QuestionMediaURL != "" {
					qMediaURL = it.QuestionMediaURL
				}

				result.Questions = append(result.Questions, dto.QuestionAnswerReviewItemDTO{
					QuestionNumber:    qNumber,
					StimulusTitle:     sTitle,
					StimulusText:      sText,
					StimulusMediaType: sMediaType,
					StimulusMediaURL:  sMediaURL,
					QuestionText:      it.QuestionText,
					QuestionMediaType: qMediaType,
					QuestionMediaURL:  qMediaURL,
					Options:           optDTOs,
					UserSelectedKey:   userChoice,
					CorrectKey:        corr,
					Status:            status,
					IsDoubtful:        isDoubtful,
					WeightCorrect:     it.WeightCorrect,
					WeightWrong:       it.WeightWrong,
					ScoreObtained:     scoreObtained,
				})

				qNumber++
			}
		}
	}

	result.Score.TotalQuestions = totalQuestions
	result.Score.CorrectCount = correctCount
	result.Score.WrongCount = wrongCount
	result.Score.EmptyCount = emptyCount
	result.Score.TotalWeight = totalWeight
	result.Score.EarnedWeight = earnedWeight
	if totalQuestions > 0 {
		result.Score.AccuracyPercent = math.Round((float64(correctCount)/float64(totalQuestions))*1000) / 10
	}
	result.Score.SubmitReason = "COMPLETED"

	// Fetch score from ep_scores
	var epScore struct {
		ListeningRaw     int    `db:"listening_raw"`
		ListeningScaled  int    `db:"listening_scaled"`
		StructureRaw     int    `db:"structure_raw"`
		StructureScaled  int    `db:"structure_scaled"`
		ReadingRaw       int    `db:"reading_raw"`
		ReadingScaled    int    `db:"reading_scaled"`
		TotalScaledScore int    `db:"total_scaled_score"`
		CEFRLevel        string `db:"cefr_level"`
		PassingStatus    string `db:"passing_status"`
	}
	scoreQ := `
		SELECT listening_raw, listening_scaled, structure_raw, structure_scaled, reading_raw, reading_scaled, total_scaled_score, cefr_level, passing_status
		FROM cat.ep_scores
		WHERE id_participant = $1
	`
	if err := r.db.GetContext(ctx, &epScore, scoreQ, h.IDParticipant); err == nil {
		result.Score.FinalScore = float64(epScore.TotalScaledScore)
		result.Score.PassingGrade = float64(h.PassingScore)
		result.Score.PassingStatus = epScore.PassingStatus
	} else {
		// Calculate dynamically from conversion profiles and save to ep_scores
		profileID := 1
		if h.IDProfile != nil && *h.IDProfile > 0 {
			profileID = *h.IDProfile
		} else {
			_ = r.db.GetContext(ctx, &profileID, `SELECT id_profile FROM cat.ep_conversion_profiles WHERE id_exam_type = $1 AND is_default = TRUE LIMIT 1`, h.IDExamType)
		}

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

		rawL := rawCountBySection[1]
		rawS := rawCountBySection[2]
		rawR := rawCountBySection[3]

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
		scaledR := 24
		if m, ok := convMap["READING"]; ok {
			if v, found := m[rawR]; found {
				scaledR = v
			}
		}

		totalScaled := ((scaledL + scaledS + scaledR) * 10) / 3
		passStatus := "TIDAK LULUS"
		if totalScaled >= h.PassingScore {
			passStatus = "LULUS"
		}

		// CEFR mapping
		type cefrRow struct {
			CEFRLevel string `db:"cefr_level"`
			ScoreMin  int    `db:"score_min"`
			ScoreMax  int    `db:"score_max"`
		}
		var cefrList []cefrRow
		_ = r.db.SelectContext(ctx, &cefrList, `SELECT cefr_level, score_min, score_max FROM cat.ep_cefr_mapping WHERE id_exam_type = $1 ORDER BY score_min ASC`, h.IDExamType)
		cefr := "B1"
		for _, cf := range cefrList {
			if totalScaled >= cf.ScoreMin && totalScaled <= cf.ScoreMax {
				cefr = cf.CEFRLevel
				break
			}
		}

		// Upsert into cat.ep_scores
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
		`
		_, _ = r.db.ExecContext(ctx, upsertScoreQ, h.IDParticipant, rawL, scaledL, rawS, scaledS, rawR, scaledR, totalScaled, cefr, passStatus)

		result.Score.FinalScore = float64(totalScaled)
		result.Score.PassingGrade = float64(h.PassingScore)
		result.Score.PassingStatus = passStatus
	}

	return result, nil
}
