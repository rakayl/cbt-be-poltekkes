package dto

import (
	"fmt"
	"strings"
	"time"
)

type ExamResponseDTO struct {
	ExamID        int     `json:"exam_id"`
	ExamName      string  `json:"exam_name"`
	PeriodID      int     `json:"period_id"`
	PeriodName    string  `json:"period_name"`
	PassingGrade  float64 `json:"passing_grade"`
	KodeJenis     string  `json:"kode_jenis"`
	NamaJenis     string  `json:"nama_jenis"`
	IDSatker      string  `json:"id_satker"`
	NamaSatker    string  `json:"nama_satker"`
	IsPercobaan   int     `json:"is_percobaan"`
	Description   *string `json:"description,omitempty"`
	MaxViolations int     `json:"max_violations"`
}

type CreateExamRequestDTO struct {
	PeriodID      int     `json:"period_id" binding:"required"`
	ExamName      string  `json:"exam_name" binding:"required"`
	KodeJenis     string  `json:"kode_jenis"`
	IDSatker      string  `json:"id_satker"`
	PassingGrade  float64 `json:"passing_grade"`
	IsPercobaan   int     `json:"is_percobaan"`
	Description   string  `json:"description"`
	MaxViolations int     `json:"max_violations"`
}

type UpdateExamRequestDTO struct {
	PeriodID      int     `json:"period_id"`
	ExamName      string  `json:"exam_name"`
	KodeJenis     string  `json:"kode_jenis"`
	IDSatker      string  `json:"id_satker"`
	PassingGrade  float64 `json:"passing_grade"`
	IsPercobaan   int     `json:"is_percobaan"`
	Description   string  `json:"description"`
	MaxViolations int     `json:"max_violations"`
}

type CreateScheduleRequestDTO struct {
	ExamID             int     `json:"exam_id"`
	NoJadwal           int     `json:"no_jadwal"`
	WaktuPengerjaan    int     `json:"waktu_pengerjaan" binding:"required"`
	Bobot              float64 `json:"bobot"`
	TampilkanNilai     int     `json:"tampilkan_nilai"`
	ExamDeadlineAt     *string `json:"exam_deadline_at,omitempty"`
	GracePeriodSeconds int     `json:"grace_period_seconds"`
	TglMulai           string  `json:"tgl_mulai"`
	TglSelesai         string  `json:"tgl_selesai"`
	WaktuMulai         string  `json:"waktu_mulai"`
	WaktuSelesai       string  `json:"waktu_selesai"`
	MaxViolations      int     `json:"max_violations"`
}

type UpdateSessionRequestDTO struct {
	NoJadwal           int     `json:"no_jadwal"`
	WaktuPengerjaan    int     `json:"waktu_pengerjaan"`
	Bobot              float64 `json:"bobot"`
	TampilkanNilai     int     `json:"tampilkan_nilai"`
	ExamDeadlineAt     *string `json:"exam_deadline_at,omitempty"`
	GracePeriodSeconds int     `json:"grace_period_seconds"`
	TglMulai           string  `json:"tgl_mulai"`
	TglSelesai         string  `json:"tgl_selesai"`
	WaktuMulai         string  `json:"waktu_mulai"`
	WaktuSelesai       string  `json:"waktu_selesai"`
	MaxViolations      int     `json:"max_violations"`
}

type CreateRoomSessionRequestDTO struct {
	ScheduleID    int    `json:"schedule_id"`
	KodeRuang     string `json:"kode_ruang" binding:"required"`
	TglMulai      string `json:"tgl_mulai" binding:"required"`
	TglSelesai    string `json:"tgl_selesai"`
	WaktuMulai    string `json:"waktu_mulai" binding:"required"`
	WaktuSelesai  string `json:"waktu_selesai"`
	JumlahPeserta int    `json:"jumlah_peserta"`
	Prioritas     int    `json:"prioritas"`
}

type UpdateRoomSessionRequestDTO struct {
	KodeRuang     string `json:"kode_ruang"`
	TglMulai      string `json:"tgl_mulai"`
	TglSelesai    string `json:"tgl_selesai"`
	WaktuMulai    string `json:"waktu_mulai"`
	WaktuSelesai  string `json:"waktu_selesai"`
	JumlahPeserta int    `json:"jumlah_peserta"`
	Prioritas     int    `json:"prioritas"`
}

type RoomResponseDTO struct {
	KodeRuang        string `json:"kode_ruang"`
	NamaRuang        string `json:"nama_ruang"`
	Kapasitas        int    `json:"kapasitas"`
	TotalKapasitas   int    `json:"total_kapasitas"`
	RoomCount        int    `json:"room_count"`
	ParticipantCount int    `json:"participant_count"`
}

type CreateExamParticipantRequestDTO struct {
	ScheduleID      int    `json:"schedule_id" binding:"required"`
	ParticipantCode string `json:"participant_code" binding:"required"`
	DeskNumber      int    `json:"desk_number"`
}

type AddExamParticipantRequestDTO struct {
	ParticipantCode string `json:"participant_code" binding:"required"`
}

type ExamImportSipenmaruRequestDTO struct {
	CBTPeriodID          int         `json:"cbt_period_id"`
	PMBPeriodID          interface{} `json:"pmb_period_id"`
	IDSistemKuliah       int         `json:"id_sistem_kuliah"`
	IDJalurPendaftaran   int         `json:"id_jalur_pendaftaran"`
	IDGelombang          int         `json:"id_gelombang"`
	OnlyAdministrasi     bool        `json:"only_administrasi"`
	SelectedPendaftarIDs []string    `json:"selected_pendaftar_ids"`
	AutoPlotting         bool        `json:"auto_plotting"`
}

func (r *ExamImportSipenmaruRequestDTO) GetPMBPeriodString() string {
	if r.PMBPeriodID == nil {
		return ""
	}
	switch val := r.PMBPeriodID.(type) {
	case string:
		return strings.TrimSpace(val)
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

type ExamImportSipenmaruResponseDTO struct {
	TotalProcessed      int      `json:"total_processed"`
	ImportedCount       int      `json:"imported_count"`
	SkippedCount        int      `json:"skipped_count"`
	ExamRegisteredCount int      `json:"exam_registered_count"`
	SessionPlottedCount int      `json:"session_plotted_count"`
	ErrorCount          int      `json:"error_count"`
	Errors              []string `json:"errors"`
}

type ScheduleDetailDTO struct {
	ScheduleID       int    `json:"schedule_id"`
	ExamID           int    `json:"exam_id"`
	ExamName         string `json:"exam_name"`
	NoJadwal         int    `json:"no_jadwal"`
	WaktuPengerjaan  int    `json:"waktu_pengerjaan"`
	RoomCount        int    `json:"room_count"`
	ParticipantCount int    `json:"participant_count"`
}

type SessionResponseDTO struct {
	SessionID          int     `json:"session_id"`
	ExamID             int     `json:"exam_id"`
	NoJadwal           int     `json:"no_jadwal"`
	WaktuPengerjaan    int     `json:"waktu_pengerjaan"`
	Bobot              float64 `json:"bobot"`
	TampilkanNilai     int     `json:"tampilkan_nilai"`
	ExamDeadlineAt     *string `json:"exam_deadline_at,omitempty"`
	GracePeriodSeconds int     `json:"grace_period_seconds"`
	RoomCount          int     `json:"room_count"`
	WaktuMulai         string  `json:"waktu_mulai,omitempty"`
	WaktuSelesai       string  `json:"waktu_selesai,omitempty"`
	TglMulai           string  `json:"tgl_mulai,omitempty"`
	TglSelesai         string  `json:"tgl_selesai,omitempty"`
	TokenUjian         string  `json:"token_ujian"`
	MaxViolations      int     `json:"max_violations"`
}

type RefreshTokenResponseDTO struct {
	ScheduleID int    `json:"schedule_id"`
	TokenUjian string `json:"token_ujian"`
	Message    string `json:"message"`
}

type SessionRoomResponseDTO struct {
	RoomSessionID int    `json:"room_session_id"`
	SessionID     int    `json:"session_id"`
	KodeRuang     string `json:"kode_ruang"`
	NamaRuang     string `json:"nama_ruang"`
	TglMulai      string `json:"tgl_mulai"`
	TglSelesai    string `json:"tgl_selesai"`
	WaktuMulai    string `json:"waktu_mulai"`
	WaktuSelesai  string `json:"waktu_selesai"`
	JumlahPeserta int    `json:"jumlah_peserta"`
	Prioritas     int    `json:"prioritas"`
}

type AvailableParticipantDTO struct {
	ParticipantCode string  `json:"participant_code"`
	Name            string  `json:"name"`
	Email           *string `json:"email,omitempty"`
	HP              *string `json:"hp,omitempty"`
}

type AutoDistributeRequestDTO struct {
	OverwriteExisting bool `json:"overwrite_existing"`
}

type AutoDistributeResponseDTO struct {
	TotalAssigned   int    `json:"total_assigned"`
	SessionsUsed    int    `json:"sessions_used"`
	RoomsUsed       int    `json:"rooms_used"`
	UnassignedCount int    `json:"unassigned_count"`
	Message         string `json:"message"`
}

type RoomParticipantDTO struct {
	ParticipantCode  string  `json:"participant_code"`
	Name             string  `json:"name"`
	Email            *string `json:"email,omitempty"`
	CityName         *string `json:"city_name,omitempty"`
	SessionID        int     `json:"session_id"`
	RoomSessionID    int     `json:"room_session_id"`
	RoomName         string  `json:"room_name"`
	IsLoggedIn       int     `json:"is_logged_in"`
	IsVerified       int     `json:"is_verified"`
	BarcodeScannedAt *string `json:"barcode_scanned_at,omitempty"`
}

type AddRoomParticipantRequestDTO struct {
	ParticipantCodes []string `json:"participant_codes" binding:"required"`
}

type ExamParticipantDTO struct {
	ParticipantCode  string   `json:"participant_code"`
	Name             string   `json:"name"`
	Email            *string  `json:"email,omitempty"`
	CityName         *string  `json:"city_name,omitempty"`
	Score            *float64 `json:"score,omitempty"`
	IsLoggedIn       int      `json:"is_logged_in"`
	IsPlotted        bool     `json:"is_plotted"`
	SessionNumber    *int     `json:"session_number,omitempty"`
	RoomName         *string  `json:"room_name,omitempty"`
	TglMulai         *string  `json:"tgl_mulai,omitempty"`
	Hari             *string  `json:"hari,omitempty"`
	WaktuMulai       *string  `json:"waktu_mulai,omitempty"`
	WaktuSelesai     *string  `json:"waktu_selesai,omitempty"`
	IsVerified       int      `json:"is_verified"`
	BarcodeScannedAt *string  `json:"barcode_scanned_at,omitempty"`
	ScannedBy        *string  `json:"scanned_by,omitempty"`
}

type ParticipantScheduleDTO struct {
	ScheduleID       int     `json:"schedule_id"`
	ExamID           int     `json:"exam_id"`
	ExamName         string  `json:"exam_name"`
	QuestionBankCode string  `json:"question_bank_code"`
	RoomID           int     `json:"room_id"`
	RoomName         string  `json:"room_name"`
	ExamDate         string  `json:"exam_date"`
	EndDate          string  `json:"end_date"`
	StartTime        string  `json:"start_time"`
	EndTime          string  `json:"end_time"`
	DurationMinutes  int     `json:"duration_minutes"`
	SessionToken     string  `json:"session_token"`
	Capacity         int     `json:"capacity"`
	ScheduleStatus   string  `json:"schedule_status"`
	IsVerified       int     `json:"is_verified"`
	BarcodeScannedAt *string `json:"barcode_scanned_at,omitempty"`
	ExamBarcode      string  `json:"exam_barcode"`
	HasStarted       bool    `json:"has_started"`
	IsFinished       bool    `json:"is_finished"`
	RemainingSeconds int     `json:"remaining_seconds"`
	MaxViolations    int     `json:"max_violations"`
	IsEP             bool    `json:"is_ep"`
	ExamType         string  `json:"exam_type"`
}

type ParticipantScheduleListResponseDTO struct {
	ActiveSchedules []*ParticipantScheduleDTO `json:"active_schedules"`
}

type ScheduleResponseDTO struct {
	ScheduleID      int       `json:"schedule_id"`
	ExamID          int       `json:"exam_id"`
	ExamName        string    `json:"exam_name"`
	PeriodID        int       `json:"period_id"`
	PeriodName      string    `json:"period_name"`
	QuestionCode    string    `json:"question_code"`
	RoomName        string    `json:"room_name"`
	ExamDate        time.Time `json:"exam_date"`
	StartTime       string    `json:"start_time"`
	EndTime         string    `json:"end_time"`
	DurationMinutes int       `json:"duration_minutes"`
	SessionToken    string    `json:"session_token"`
	Capacity        int       `json:"capacity"`
	MaxViolations   int       `json:"max_violations"`
}

type StartExamRequestDTO struct {
	ScheduleID        int    `json:"schedule_id" binding:"required"`
	SessionToken      string `json:"session_token"`
	DeviceID          string `json:"device_id,omitempty"`
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
}

type QuestionOptionDTO struct {
	OptionKey  int     `json:"option_key"`
	OptionText string  `json:"option_text"`
	MediaType  string  `json:"media_type,omitempty"` // NONE, AUDIO, IMAGE, VIDEO
	MediaURL   *string `json:"media_url,omitempty"`
}

type SessionQuestionItemDTO struct {
	QuestionNumber    int                 `json:"question_number"`
	QuestionText      string              `json:"question_text"`
	QuestionMediaType string              `json:"question_media_type,omitempty"` // NONE, AUDIO, IMAGE, VIDEO
	QuestionMediaURL  *string             `json:"question_media_url,omitempty"`
	Options           []QuestionOptionDTO `json:"options"`
	CategoryTag       string              `json:"category_tag,omitempty"`
	QuestionType      string              `json:"question_type,omitempty"` // STATIC, DYNAMIC_STIMULUS
	WeightCorrect     float64             `json:"weight_correct"`
	WeightWrong       float64             `json:"weight_wrong"`
	StimulusID        *int64              `json:"stimulus_id,omitempty"`
	StimulusItemID    *int64              `json:"stimulus_item_id,omitempty"`
	StimulusTitle     *string             `json:"stimulus_title,omitempty"`
	StimulusText      *string             `json:"stimulus_text,omitempty"`
	StimulusMediaType string              `json:"stimulus_media_type,omitempty"`
	StimulusMediaURL  *string             `json:"stimulus_media_url,omitempty"`
}

type SavedUserAnswerDTO struct {
	SelectedOptionKey int  `json:"selected_option_key"`
	IsDoubtful        bool `json:"is_doubtful"`
}

type SessionQuestionsResponseDTO struct {
	ScheduleID       int                        `json:"schedule_id"`
	QuestionBankCode string                     `json:"question_bank_code"`
	TotalQuestions   int                        `json:"total_questions"`
	SnapshotVersion  int                        `json:"snapshot_version"`
	RemainingSeconds int                        `json:"remaining_seconds"`
	MaxViolations    int                        `json:"max_violations"`
	UserAnswers      map[int]SavedUserAnswerDTO `json:"user_answers"`
	Questions        []SessionQuestionItemDTO   `json:"questions"`
}

type StartExamResponseDTO struct {
	ScheduleID       int    `json:"schedule_id"`
	ExamID           int    `json:"exam_id"`
	QuestionBankCode string `json:"question_bank_code"`
	DurationMinutes  int    `json:"duration_minutes"`
	RemainingSeconds int    `json:"remaining_seconds"`
	MaxViolations    int    `json:"max_violations"`
	StartedAt        string `json:"started_at"`
	ServerTime       string `json:"server_time"`
	IsLocked         int    `json:"is_locked"`
}

type SaveAnswerRequestDTO struct {
	ScheduleID     int    `json:"schedule_id" binding:"required"`
	QuestionCode   string `json:"question_code" binding:"required"`
	QuestionNumber int    `json:"question_number" binding:"required"`
	SelectedOption int    `json:"selected_option_key"`
	IsDoubtful     bool   `json:"is_doubtful"`
}

type SaveAnswerResponseDTO struct {
	QuestionNumber int    `json:"question_number"`
	SelectedOption int    `json:"selected_option"`
	IsDoubtful     bool   `json:"is_doubtful"`
	SavedAt        string `json:"saved_at"`
}

type HeartbeatRequestDTO struct {
	ScheduleID             int    `json:"schedule_id"`
	DeviceID               string `json:"device_id,omitempty"`
	IsFullscreen           bool   `json:"is_fullscreen"`
	CurrentQuestion        int    `json:"current_question,omitempty"`
	IsPaused               bool   `json:"is_paused"`
	IsCameraActive         bool   `json:"is_camera_active"`
	ClientRemainingSeconds int    `json:"client_remaining_seconds,omitempty"`
}

type HeartbeatResponseDTO struct {
	ServerTime       string  `json:"server_time"`
	RemainingSeconds int     `json:"remaining_seconds"`
	IsLocked         int     `json:"is_locked"`
	LockReason       *string `json:"lock_reason,omitempty"`
	ProctorWarning   *string `json:"proctor_warning,omitempty"`
	ShouldSubmit     bool    `json:"should_submit"`
	SubmitReason     string  `json:"submit_reason,omitempty"`
	TabSwitchCount   int     `json:"tab_switch_count"`
	RiskScore        int     `json:"risk_score"`
	MaxViolations    int     `json:"max_violations"`
}

type SecurityEventRequestDTO struct {
	ScheduleID int                    `json:"schedule_id" binding:"required"`
	EventType  string                 `json:"event_type" binding:"required"`
	Severity   string                 `json:"severity,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	DeviceID   string                 `json:"device_id,omitempty"`
}

type SecurityEventResponseDTO struct {
	CurrentRiskScore    int    `json:"current_risk_score"`
	RiskLevel           string `json:"risk_level"`
	TabSwitchCount      int    `json:"tab_switch_count"`
	FullscreenExitCount int    `json:"fullscreen_exit_count"`
	IsLocked            int    `json:"is_locked"`
	AutoLocked          bool   `json:"auto_locked"`
	MaxViolations       int    `json:"max_violations"`
}

type SecurityEventItemDTO struct {
	ID              int64                  `json:"id"`
	ScheduleID      int                    `json:"schedule_id"`
	ParticipantCode string                 `json:"participant_code"`
	ParticipantName string                 `json:"participant_name,omitempty"`
	EventType       string                 `json:"event_type"`
	Severity        string                 `json:"severity"`
	RiskScore       int                    `json:"risk_score"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	IPAddress       string                 `json:"ip_address,omitempty"`
	CreatedAt       string                 `json:"created_at"`
}

type ProctorActionRequestDTO struct {
	ScheduleID      int    `json:"schedule_id" binding:"required"`
	ParticipantCode string `json:"participant_code" binding:"required"`
	Action          string `json:"action" binding:"required"` // LOCK_SESSION, UNLOCK_SESSION, FORCE_SUBMIT, SEND_WARNING, RESET_SESSION
	Reason          string `json:"reason,omitempty"`
	WarningMessage  string `json:"warning_message,omitempty"`
}

type LiveMonitoringParticipantDTO struct {
	ParticipantCode           string  `json:"participant_code"`
	Name                      string  `json:"name"`
	DeskNumber                int     `json:"desk_number"`
	ExamStatus                string  `json:"exam_status"`
	TotalQuestions            int     `json:"total_questions"`
	AnsweredCount             int     `json:"answered_count"`
	RemainingSeconds          int     `json:"remaining_seconds"`
	ConnectionStatus          string  `json:"connection_status"`
	LastActivity              string  `json:"last_activity"`
	RiskScore                 int     `json:"risk_score"`
	RiskLevel                 string  `json:"risk_level"`
	TabSwitchCount            int     `json:"tab_switch_count"`
	FullscreenExitCount       int     `json:"fullscreen_exit_count"`
	ActiveTabSwitchCount      int     `json:"active_tab_switch_count"`
	ActiveFullscreenExitCount int     `json:"active_fullscreen_exit_count"`
	TotalViolations           int     `json:"total_violations"`
	UnlockCount               int     `json:"unlock_count"`
	DisconnectCount           int     `json:"disconnect_count"`
	ReconnectCount            int     `json:"reconnect_count"`
	IsLocked                  int     `json:"is_locked"`
	LockReason                *string `json:"lock_reason,omitempty"`
	ProctorWarning            *string `json:"proctor_warning,omitempty"`
	IsVerified                int     `json:"is_verified"`
	BarcodeScannedAt          *string `json:"barcode_scanned_at,omitempty"`
}

type AttendanceStatsDTO struct {
	TotalParticipants int     `json:"total_participants"`
	VerifiedCount     int     `json:"verified_count"`
	UnverifiedCount   int     `json:"unverified_count"`
	AttendancePct     float64 `json:"attendance_pct"`
}

type AttendanceParticipantItemDTO struct {
	ParticipantCode  string  `json:"participant_code"`
	Name             string  `json:"name"`
	RoomName         string  `json:"room_name"`
	DeskNumber       int     `json:"desk_number"`
	SessionNumber    int     `json:"session_number"`
	IsVerified       int     `json:"is_verified"`
	BarcodeScannedAt *string `json:"barcode_scanned_at"`
	ScannedBy        *string `json:"scanned_by"`
}

type AttendanceSummaryResponseDTO struct {
	ScheduleID   int                            `json:"schedule_id"`
	SessionNo    int                            `json:"session_no"`
	ExamName     string                         `json:"exam_name"`
	Stats        AttendanceStatsDTO             `json:"stats"`
	Participants []AttendanceParticipantItemDTO `json:"participants"`
}

type VerifyBarcodeRequestDTO struct {
	ScheduleID int    `json:"schedule_id" binding:"required"`
	Barcode    string `json:"barcode" binding:"required"`
}

type VerifyBarcodeResponseDTO struct {
	ParticipantCode string              `json:"participant_code"`
	Name            string              `json:"name"`
	IsVerified      int                 `json:"is_verified"`
	AlreadyVerified bool                `json:"already_verified"`
	ScannedAt       string              `json:"scanned_at"`
	ScannedBy       string              `json:"scanned_by"`
	Message         string              `json:"message"`
	DeskNumber      int                 `json:"desk_number,omitempty"`
	RoomName        string              `json:"room_name,omitempty"`
	SessionNumber   int                 `json:"session_number,omitempty"`
	Stats           *AttendanceStatsDTO `json:"stats,omitempty"`
}

type ProctorManualVerifyDTO struct {
	ScheduleID int `json:"schedule_id" binding:"required"`
	IsVerified int `json:"is_verified"`
}

type LiveMonitoringResponseDTO struct {
	ScheduleID   int                            `json:"schedule_id"`
	Stats        map[string]interface{}         `json:"stats"`
	Participants []LiveMonitoringParticipantDTO `json:"participants"`
}

type FinishExamRequestDTO struct {
	SubmitReason string `json:"submit_reason,omitempty"` // MANUAL_SUBMIT, PARTICIPANT_TIMEOUT, EXAM_DEADLINE, FORCE_SUBMIT
}

type FinishExamResponseDTO struct {
	ParticipantCode string  `json:"participant_code"`
	ScheduleID      int     `json:"schedule_id"`
	TotalQuestions  int     `json:"total_questions"`
	CorrectCount    int     `json:"correct_count"`
	WrongCount      int     `json:"wrong_count"`
	EmptyCount      int     `json:"empty_count"`
	FinalScore      float64 `json:"final_score"`
	TotalWeight     float64 `json:"total_weight"`
	EarnedWeight    float64 `json:"earned_weight"`
	PassingStatus   string  `json:"passing_status"`
	SubmittedAt     string  `json:"submitted_at"`
	SubmitReason    string  `json:"submit_reason"`
}

type CollusionPairDTO struct {
	ParticipantCode1         string  `json:"participant_code_1"`
	Name1                    string  `json:"name_1"`
	DeskNumber1              int     `json:"desk_number_1"`
	ParticipantCode2         string  `json:"participant_code_2"`
	Name2                    string  `json:"name_2"`
	DeskNumber2              int     `json:"desk_number_2"`
	OverallSimilarityPercent float64 `json:"overall_similarity_percent"`
	WrongSimilarityPercent   float64 `json:"wrong_similarity_percent"`
	IdenticalCount           int     `json:"identical_count"`
	TotalCompared            int     `json:"total_compared"`
	IdenticalWrongCount      int     `json:"identical_wrong_count"`
	RiskStatus               string  `json:"risk_status"` // HIGH_COLLUSION, SUSPICIOUS, NORMAL
}

type CollusionReportResponseDTO struct {
	ScheduleID           int                `json:"schedule_id"`
	TotalParticipants    int                `json:"total_participants"`
	TotalPairsAnalyzed   int                `json:"total_pairs_analyzed"`
	SuspiciousPairsCount int                `json:"suspicious_pairs_count"`
	Pairs                []CollusionPairDTO `json:"pairs"`
}

// Dynamic Multimedia & Multi-Item DTOs
type DynamicItemOptionDTO struct {
	IDOption    int64   `json:"id_option"`
	OptionLabel string  `json:"option_label"` // A, B, C, D, E
	OptionText  string  `json:"option_text"`
	MediaType   string  `json:"media_type"` // NONE, AUDIO, IMAGE
	MediaURL    *string `json:"media_url,omitempty"`
}

type DynamicSubItemDTO struct {
	IDItem             int64                  `json:"id_item"`
	ItemOrder          int                    `json:"item_order"`
	QuestionText       string                 `json:"question_text"`
	QuestionMediaType  string                 `json:"question_media_type"` // NONE, AUDIO, IMAGE, VIDEO
	QuestionMediaURL   *string                `json:"question_media_url,omitempty"`
	ItemType           string                 `json:"item_type"` // SINGLE_CHOICE, MULTI_CHOICE
	WeightCorrect      float64                `json:"weight_correct"`
	WeightWrong        float64                `json:"weight_wrong"`
	Options            []DynamicItemOptionDTO `json:"options"`
	UserSelectedOption *string                `json:"user_selected_option,omitempty"`
	IsDoubtful         bool                   `json:"is_doubtful"`
}

type DynamicStimulusUnitDTO struct {
	IDStimulus     int64               `json:"id_stimulus"`
	StimulusNumber int                 `json:"stimulus_number"`
	StimulusCode   string              `json:"stimulus_code"`
	Title          string              `json:"title"`
	NarrativeText  string              `json:"narrative_text"`
	MediaType      string              `json:"media_type"` // NONE, AUDIO, IMAGE, VIDEO, MULTI
	MediaURL       *string             `json:"media_url,omitempty"`
	MediaMetadata  any                 `json:"media_metadata,omitempty"`
	CategoryTag    string              `json:"category_tag"`
	SubItems       []DynamicSubItemDTO `json:"sub_items"`
}

type DynamicExamSessionResponseDTO struct {
	ScheduleID       int                      `json:"schedule_id"`
	ExamName         string                   `json:"exam_name"`
	IsDynamic        bool                     `json:"is_dynamic"`
	TimeWindowStart  string                   `json:"time_window_start"`
	TimeWindowEnd    string                   `json:"time_window_end"`
	DurationMinutes  int                      `json:"duration_minutes"`
	RemainingSeconds int                      `json:"remaining_seconds"`
	SessionStatus    string                   `json:"session_status"` // NOT_STARTED, IN_PROGRESS, SUBMITTED, EXPIRED_AUTO_SUBMIT
	ScoringRule      string                   `json:"scoring_rule"`   // STANDARD, PENALTY_MINUS_ONE
	ServerTimeWIB    string                   `json:"server_time_wib"`
	TotalStimuli     int                      `json:"total_stimuli"`
	TotalSubItems    int                      `json:"total_sub_items"`
	Stimuli          []DynamicStimulusUnitDTO `json:"stimuli"`
}

type SaveMultiAnswerDTO struct {
	ScheduleID      int    `json:"schedule_id" binding:"required"`
	ParticipantCode string `json:"participant_code"`
	IDStimulus      int64  `json:"id_stimulus" binding:"required"`
	IDItem          int64  `json:"id_item" binding:"required"`
	SelectedOption  string `json:"selected_option"` // e.g. "A"
	IsDoubtful      bool   `json:"is_doubtful"`
}

// === Bank Soal per Session ===

type SessionBankSoalDTO struct {
	KodeSoal               string  `json:"kode_soal"`
	NamaSoal               string  `json:"nama_soal"`
	NoUrut                 int     `json:"no_urut"`
	JumlahPertanyaan       int     `json:"jumlah_pertanyaan"`
	JumlahSoalStatis       int     `json:"jumlah_soal_statis"`
	JumlahKasusDinamis     int     `json:"jumlah_kasus_dinamis"`
	Bobot                  float64 `json:"bobot"`
	KodeSkor               string  `json:"kode_skor"`
	TotalAvailable         int     `json:"total_available"`
	TotalStatisAvailable   int     `json:"total_statis_available"`
	TotalStimulusAvailable int     `json:"total_stimulus_available"`
}

type AddBankSoalToSessionDTO struct {
	KodeSoal           string  `json:"kode_soal" binding:"required"`
	JumlahPertanyaan   int     `json:"jumlah_pertanyaan"`
	JumlahSoalStatis   int     `json:"jumlah_soal_statis"`
	JumlahKasusDinamis int     `json:"jumlah_kasus_dinamis"`
	Bobot              float64 `json:"bobot"`
	KodeSkor           string  `json:"kode_skor"`
}

// === Completed Participants & Detailed Review DTOs ===

type CompletedParticipantDTO struct {
	ParticipantCode     string  `json:"participant_code"`
	Name                string  `json:"name"`
	Email               *string `json:"email"`
	ScheduleID          int     `json:"schedule_id"`
	SessionNumber       int     `json:"session_number"`
	RoomName            string  `json:"room_name"`
	StartedAt           *string `json:"started_at"`
	FinishedAt          *string `json:"finished_at"`
	DurationMinutes     int     `json:"duration_minutes"`
	FinalScore          float64 `json:"final_score"`
	PassingStatus       string  `json:"passing_status"` // "LULUS" / "TIDAK LULUS"
	RiskScore           int     `json:"risk_score"`
	RiskLevel           string  `json:"risk_level"`
	TabSwitchCount      int     `json:"tab_switch_count"`
	FullscreenExitCount int     `json:"fullscreen_exit_count"`
	TotalViolations     int     `json:"total_violations"`
	UnlockCount         int     `json:"unlock_count"`
	DisconnectCount     int     `json:"disconnect_count"`
	IsLocked            int     `json:"is_locked"`
	LockReason          *string `json:"lock_reason"`
	TotalAnswered       int     `json:"total_answered"`
	CorrectCount        int     `json:"correct_count"`
	WrongCount          int     `json:"wrong_count"`
	EmptyCount          int     `json:"empty_count"`
	IsVerified          int     `json:"is_verified"`
}

type QuestionOptionReviewDTO struct {
	OptionKey    string  `json:"option_key"` // "A", "B", "C", "D", "E"
	OptionText   string  `json:"option_text"`
	MediaURL     *string `json:"media_url,omitempty"`
	IsUserChoice bool    `json:"is_user_choice"`
	IsCorrect    bool    `json:"is_correct"`
}

type QuestionAnswerReviewItemDTO struct {
	QuestionNumber    int                       `json:"question_number"`
	StimulusTitle     *string                   `json:"stimulus_title,omitempty"`
	StimulusText      *string                   `json:"stimulus_text,omitempty"`
	StimulusMediaType *string                   `json:"stimulus_media_type,omitempty"`
	StimulusMediaURL  *string                   `json:"stimulus_media_url,omitempty"`
	QuestionText      string                    `json:"question_text"`
	QuestionMediaType *string                   `json:"question_media_type,omitempty"`
	QuestionMediaURL  *string                   `json:"question_media_url,omitempty"`
	Options           []QuestionOptionReviewDTO `json:"options"`
	UserSelectedKey   string                    `json:"user_selected_key"` // e.g. "A", "B", "C", "D", "E" or ""
	CorrectKey        string                    `json:"correct_key"`       // e.g. "A", "B", "C", "D", "E"
	Status            string                    `json:"status"`            // "CORRECT", "WRONG", "EMPTY"
	IsDoubtful        bool                      `json:"is_doubtful"`
	WeightCorrect     float64                   `json:"weight_correct"`
	WeightWrong       float64                   `json:"weight_wrong"`
	ScoreObtained     float64                   `json:"score_obtained"`
}

type ParticipantExamResultDetailDTO struct {
	// 1. Data Peserta
	Participant struct {
		ParticipantCode  string  `json:"participant_code"`
		Name             string  `json:"name"`
		Email            *string `json:"email"`
		ScheduleID       int     `json:"schedule_id"`
		ExamID           int     `json:"exam_id"`
		ExamName         string  `json:"exam_name"`
		SessionNumber    int     `json:"session_number"`
		RoomName         string  `json:"room_name"`
		StartedAt        *string `json:"started_at"`
		FinishedAt       *string `json:"finished_at"`
		DurationMinutes  int     `json:"duration_minutes"`
		IsVerified       int     `json:"is_verified"`
		BarcodeScannedAt *string `json:"barcode_scanned_at"`
		ClientDeviceID   *string `json:"client_device_id"`
		ExamMode         string  `json:"exam_mode"` // "STANDARD" | "DYNAMIC"
	} `json:"participant"`

	// 2. Nilai Peserta
	Score struct {
		FinalScore      float64 `json:"final_score"`
		PassingGrade    float64 `json:"passing_grade"`
		PassingStatus   string  `json:"passing_status"`
		TotalQuestions  int     `json:"total_questions"`
		CorrectCount    int     `json:"correct_count"`
		WrongCount      int     `json:"wrong_count"`
		EmptyCount      int     `json:"empty_count"`
		AccuracyPercent float64 `json:"accuracy_percent"`
		TotalWeight     float64 `json:"total_weight"`
		EarnedWeight    float64 `json:"earned_weight"`
		SubmitReason    string  `json:"submit_reason"`
	} `json:"score"`

	// 3. Record Pelanggaran Peserta
	Violations struct {
		RiskScore                 int                    `json:"risk_score"`
		RiskLevel                 string                 `json:"risk_level"`
		IsLocked                  int                    `json:"is_locked"`
		LockReason                *string                `json:"lock_reason"`
		ProctorWarning            *string                `json:"proctor_warning"`
		TabSwitchCount            int                    `json:"tab_switch_count"`
		FullscreenExitCount       int                    `json:"fullscreen_exit_count"`
		ActiveTabSwitchCount      int                    `json:"active_tab_switch_count"`
		ActiveFullscreenExitCount int                    `json:"active_fullscreen_exit_count"`
		TotalViolations           int                    `json:"total_violations"`
		UnlockCount               int                    `json:"unlock_count"`
		DisconnectCount           int                    `json:"disconnect_count"`
		ReconnectCount            int                    `json:"reconnect_count"`
		Events                    []SecurityEventItemDTO `json:"events"`
	} `json:"violations"`

	// 4. Soal dan Jawaban Peserta
	Questions []QuestionAnswerReviewItemDTO `json:"questions"`
}

type ParticipantExamSummaryDTO struct {
	ScheduleID      int     `json:"schedule_id"`
	ExamID          int     `json:"exam_id"`
	ExamName        string  `json:"exam_name"`
	PeriodName      string  `json:"period_name"`
	SessionNumber   int     `json:"session_number"`
	RoomName        string  `json:"room_name"`
	ExamDate        string  `json:"exam_date"`
	StartedAt       *string `json:"started_at,omitempty"`
	FinishedAt      *string `json:"finished_at,omitempty"`
	DurationMinutes int     `json:"duration_minutes"`
	FinalScore      float64 `json:"final_score"`
	PassingGrade    float64 `json:"passing_grade"`
	PassingStatus   string  `json:"passing_status"`
	TampilkanNilai  int     `json:"tampilkan_nilai"`
	RiskScore       int     `json:"risk_score"`
	TabSwitchCount  int     `json:"tab_switch_count"`
	IsLocked        int     `json:"is_locked"`
	TotalAnswered   int     `json:"total_answered"`
	TotalQuestions  int     `json:"total_questions"`
}
