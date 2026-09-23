package dto

import "time"

// Exam Type & Sections
type ExamTypeDTO struct {
	IDExamType     int          `json:"id_exam_type"`
	TypeCode       string       `json:"type_code"`
	TypeName       string       `json:"type_name"`
	TotalSections  int          `json:"total_sections"`
	ScoreMin       int          `json:"score_min"`
	ScoreMax       int          `json:"score_max"`
	ScoreFormula   string       `json:"score_formula"`
	ValidityMonths int          `json:"validity_months"`
	IsActive       bool         `json:"is_active"`
	Sections       []SectionDTO `json:"sections,omitempty"`
}

type SectionDTO struct {
	IDSection        int              `json:"id_section"`
	IDExamType       int              `json:"id_exam_type"`
	SectionCode      string           `json:"section_code"`
	SectionName      string           `json:"section_name"`
	SectionOrder     int              `json:"section_order"`
	DefaultQuestions int              `json:"default_questions"`
	DefaultDuration  int              `json:"default_duration"`
	HasAudio         bool             `json:"has_audio"`
	AllowReplay      bool             `json:"allow_replay"`
	MaxReplayCount   int              `json:"max_replay_count"`
	CanGoBack        bool             `json:"can_go_back"`
	ScoreScaleMin    int              `json:"score_scale_min"`
	ScoreScaleMax    int              `json:"score_scale_max"`
	Parts            []SectionPartDTO `json:"parts,omitempty"`
}

type SectionPartDTO struct {
	IDPart          int     `json:"id_part"`
	PartCode        string  `json:"part_code"`
	PartName        string  `json:"part_name"`
	PartOrder       int     `json:"part_order"`
	QuestionStart   int     `json:"question_start"`
	QuestionEnd     int     `json:"question_end"`
	InstructionText *string `json:"instruction_text,omitempty"`
	HasPassage      bool    `json:"has_passage"`
	HasAudio        bool    `json:"has_audio"`
}

// Conversion Profiles
type ConversionProfileDTO struct {
	IDProfile   int                  `json:"id_profile"`
	IDExamType  int                  `json:"id_exam_type"`
	ProfileCode string               `json:"profile_code"`
	ProfileName string               `json:"profile_name"`
	IsDefault   bool                 `json:"is_default"`
	Description *string              `json:"description,omitempty"`
	Tables      []ScoreConversionDTO `json:"tables,omitempty"`
}

type ScoreConversionDTO struct {
	SectionCode string `json:"section_code"`
	RawScore    int    `json:"raw_score"`
	ScaledScore int    `json:"scaled_score"`
}

type SaveConversionProfileRequestDTO struct {
	IDExamType  int                  `json:"id_exam_type" binding:"required"`
	ProfileCode string               `json:"profile_code" binding:"required"`
	ProfileName string               `json:"profile_name" binding:"required"`
	IsDefault   bool                 `json:"is_default"`
	Description *string              `json:"description"`
	Tables      []ScoreConversionDTO `json:"tables" binding:"required"`
}

// Exam & Dynamic Section Config
type ExamSectionConfigDTO struct {
	IDSection       int    `json:"id_section" binding:"required"`
	KodeSoal        string `json:"kodesoal" binding:"required"`
	SectionOrder    int    `json:"section_order"`
	DurationMinutes int    `json:"duration_minutes" binding:"required"`
	AllowReplay     bool   `json:"allow_replay"`
	MaxReplayCount  int    `json:"max_replay_count"`
	ShuffleOptions  bool   `json:"shuffle_options"`
}

type CreateEPExamRequestDTO struct {
	IDExamType             int                    `json:"id_exam_type" binding:"required"`
	IDProfile              *int                   `json:"id_profile"`
	IDTemplate             *int                   `json:"id_template"`
	ExamName               string                 `json:"exam_name" binding:"required"`
	PassingScore           int                    `json:"passing_score"`
	ValidityMonthsOverride *int                   `json:"validity_months_override"`
	Description            *string                `json:"description"`
	Sections               []ExamSectionConfigDTO `json:"sections" binding:"required"`
}

type EPExamResponseDTO struct {
	IDEPExam               int                    `json:"id_ep_exam"`
	IDExamType             int                    `json:"id_exam_type"`
	ExamTypeName           string                 `json:"exam_type_name"`
	IDProfile              *int                   `json:"id_profile,omitempty"`
	ProfileName            *string                `json:"profile_name,omitempty"`
	IDTemplate             *int                   `json:"id_template,omitempty"`
	ExamName               string                 `json:"exam_name"`
	PassingScore           int                    `json:"passing_score"`
	ValidityMonthsOverride *int                   `json:"validity_months_override,omitempty"`
	Description            *string                `json:"description,omitempty"`
	IsActive               bool                   `json:"is_active"`
	CreatedAt              time.Time              `json:"created_at"`
	Sections               []EPExamSectionItemDTO `json:"sections"`
}

type EPExamSectionItemDTO struct {
	IDEPExamSection int    `json:"id_ep_exam_section"`
	IDSection       int    `json:"id_section"`
	SectionCode     string `json:"section_code"`
	SectionName     string `json:"section_name"`
	KodeSoal        string `json:"kodesoal"`
	SectionOrder    int    `json:"section_order"`
	DurationMinutes int    `json:"duration_minutes"`
	AllowReplay     bool   `json:"allow_replay"`
	MaxReplayCount  int    `json:"max_replay_count"`
	ShuffleOptions  bool   `json:"shuffle_options"`
	HasAudio        bool   `json:"has_audio"`
}

// Schedules
type CreateEPScheduleRequestDTO struct {
	IDEPExam     int    `json:"id_ep_exam" binding:"required"`
	IDRuang      int    `json:"id_ruang" binding:"required"`
	ExamDate     string `json:"exam_date" binding:"required"` // YYYY-MM-DD
	StartTime    string `json:"start_time" binding:"required"`
	EndTime      string `json:"end_time" binding:"required"`
	SessionToken string `json:"session_token" binding:"required"`
	Capacity     int    `json:"capacity"`
	ProctorName  string `json:"proctor_name"`
}

type UpdateEPScheduleRequestDTO struct {
	IDRuang      int    `json:"id_ruang" binding:"required"`
	ExamDate     string `json:"exam_date" binding:"required"` // YYYY-MM-DD
	StartTime    string `json:"start_time" binding:"required"`
	EndTime      string `json:"end_time" binding:"required"`
	SessionToken string `json:"session_token" binding:"required"`
	Capacity     int    `json:"capacity"`
	ProctorName  string `json:"proctor_name"`
	IsActive     *bool  `json:"is_active"`
}

type EPScheduleResponseDTO struct {
	IDEPSchedule int       `json:"id_ep_schedule"`
	IDEPExam     int       `json:"id_ep_exam"`
	ExamName     string    `json:"exam_name"`
	IDRuang      int       `json:"id_ruang"`
	RoomName     string    `json:"room_name"`
	ExamDate     time.Time `json:"exam_date"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	SessionToken string    `json:"session_token"`
	Capacity     int       `json:"capacity"`
	Registered   int       `json:"registered"`
	ProctorName  *string   `json:"proctor_name,omitempty"`
	IsActive     bool      `json:"is_active"`
}

type RegisterParticipantsRequestDTO struct {
	ParticipantCodes []string                     `json:"participant_codes"`
	Participants     []RegisterParticipantItemDTO `json:"participants"`
}

type RegisterParticipantItemDTO struct {
	KodePeserta string  `json:"kodepeserta" binding:"required"`
	Nama        string  `json:"nama"`
	Email       *string `json:"email"`
	HP          *string `json:"hp"`
	Kategori    string  `json:"kategori"` // MAHASISWA, PEGAWAI, UMUM
	Institusi   string  `json:"institusi"`
}

// Test Taker Session (Dynamic Examination)
type StartEPExamRequestDTO struct {
	ScheduleID   int    `json:"schedule_id" binding:"required"`
	SessionToken string `json:"session_token"`
}

type DynamicOptionDTO struct {
	IDOption    int64   `json:"id_option"`
	OptionLabel string  `json:"option_label"`
	OptionText  string  `json:"option_text"`
	MediaType   string  `json:"media_type"`
	MediaURL    *string `json:"media_url,omitempty"`
	OriginalKey string  `json:"original_key,omitempty"`
}

type DynamicItemDTO struct {
	IDItem             int64              `json:"id_item"`
	ItemOrder          int                `json:"item_order"`
	QuestionText       string             `json:"question_text"`
	QuestionMediaType  string             `json:"question_media_type"`
	QuestionMediaURL   *string            `json:"question_media_url,omitempty"`
	CorrectAnswer      string             `json:"correct_answer,omitempty"`
	Explanation        *string            `json:"explanation,omitempty"`
	Options            []DynamicOptionDTO `json:"options"`
	UserSelectedOption *string            `json:"user_selected_option,omitempty"`
	IsDoubtful         bool               `json:"is_doubtful"`
}

type CreateSubItemRequestDTO struct {
	ItemOrder         int                           `json:"item_order"`
	QuestionText      string                        `json:"question_text" binding:"required"`
	QuestionMediaType string                        `json:"question_media_type"`
	QuestionMediaURL  *string                       `json:"question_media_url,omitempty"`
	CorrectAnswer     string                        `json:"correct_answer" binding:"required"`
	Explanation       *string                       `json:"explanation,omitempty"`
	Options           []CreateStimulusItemOptionDTO `json:"options" binding:"required"`
}

type UpdateSubItemRequestDTO struct {
	QuestionText      string                        `json:"question_text" binding:"required"`
	QuestionMediaType string                        `json:"question_media_type"`
	QuestionMediaURL  *string                       `json:"question_media_url,omitempty"`
	CorrectAnswer     string                        `json:"correct_answer" binding:"required"`
	Explanation       *string                       `json:"explanation,omitempty"`
	Options           []CreateStimulusItemOptionDTO `json:"options" binding:"required"`
}

type DynamicStimulusDTO struct {
	IDStimulus     int64            `json:"id_stimulus"`
	StimulusNumber int              `json:"stimulus_number"`
	StimulusCode   string           `json:"stimulus_code"`
	Title          string           `json:"title"`
	NarrativeText  string           `json:"narrative_text"`
	MediaType      string           `json:"media_type"`
	MediaURL       *string          `json:"media_url,omitempty"`
	MediaMetadata  any              `json:"media_metadata,omitempty"`
	CategoryTag    string           `json:"category_tag"`
	SubItems       []DynamicItemDTO `json:"sub_items"`
}

type EPExamSessionResponseDTO struct {
	ScheduleID          int                  `json:"schedule_id"`
	ExamName            string               `json:"exam_name"`
	ParticipantCode     string               `json:"participant_code"`
	ParticipantName     string               `json:"participant_name"`
	TotalSections       int                  `json:"total_sections"`
	CurrentSectionOrder int                  `json:"current_section_order"`
	CurrentSectionCode  string               `json:"current_section_code"`
	CurrentSectionName  string               `json:"current_section_name"`
	DurationMinutes     int                  `json:"duration_minutes"`
	RemainingSeconds    int                  `json:"remaining_seconds"`
	AllowReplay         bool                 `json:"allow_replay"`
	MaxReplayCount      int                  `json:"max_replay_count"`
	ShuffleOptions      bool                 `json:"shuffle_options"`
	HasAudio            bool                 `json:"has_audio"`
	CanGoBack           bool                 `json:"can_go_back"`
	SessionStatus       string               `json:"session_status"`
	ServerTimeWIB       string               `json:"server_time_wib"`
	TotalQuestions      int                  `json:"total_questions"`
	TotalAnswered       int                  `json:"total_answered"`
	Stimuli             []DynamicStimulusDTO `json:"stimuli"`
}

type ImportBankQuestionsResponseDTO struct {
	StimulusCount int `json:"stimulus_count"`
	QuestionCount int `json:"question_count"`
}

type NextSectionRequestDTO struct {
	ScheduleID int `json:"schedule_id" binding:"required"`
}

type SaveAnswerDTO struct {
	ScheduleID     int    `json:"schedule_id" binding:"required"`
	IDStimulus     int64  `json:"id_stimulus" binding:"required"`
	IDItem         int64  `json:"id_item" binding:"required"`
	SelectedOption string `json:"selected_option"`
	IsDoubtful     bool   `json:"is_doubtful"`
}

type FinishEPExamDTO struct {
	ScheduleID   int    `json:"schedule_id" binding:"required"`
	SubmitReason string `json:"submit_reason"`
}

// Scores & Certificates
type EPScoreDetailDTO struct {
	IDScore          int     `json:"id_score"`
	ParticipantCode  string  `json:"participant_code"`
	ParticipantName  string  `json:"participant_name"`
	ListeningRaw     int     `json:"listening_raw"`
	ListeningScaled  int     `json:"listening_scaled"`
	StructureRaw     int     `json:"structure_raw"`
	StructureScaled  int     `json:"structure_scaled"`
	ReadingRaw       int     `json:"reading_raw"`
	ReadingScaled    int     `json:"reading_scaled"`
	TotalScaledScore int     `json:"total_scaled_score"`
	CEFRLevel        *string `json:"cefr_level,omitempty"`
	PassingStatus    string  `json:"passing_status"`
	CalculatedAt     string  `json:"calculated_at"`
}

type ParticipantScoreResultDTO struct {
	IDScore           int     `json:"id_score"`
	ScheduleID        int     `json:"schedule_id"`
	ExamName          string  `json:"exam_name"`
	RoomName          string  `json:"room_name"`
	ExamDate          string  `json:"exam_date"`
	StartTime         string  `json:"start_time"`
	ParticipantCode   string  `json:"participant_code"`
	ParticipantName   string  `json:"participant_name"`
	ListeningRaw      int     `json:"listening_raw"`
	ListeningScaled   int     `json:"listening_scaled"`
	StructureRaw      int     `json:"structure_raw"`
	StructureScaled   int     `json:"structure_scaled"`
	ReadingRaw        int     `json:"reading_raw"`
	ReadingScaled     int     `json:"reading_scaled"`
	TotalScaledScore  int     `json:"total_scaled_score"`
	CEFRLevel         *string `json:"cefr_level,omitempty"`
	PassingStatus     string  `json:"passing_status"`
	CalculatedAt      string  `json:"calculated_at"`
	HasCertificate    bool    `json:"has_certificate"`
	CertificateNumber *string `json:"certificate_number,omitempty"`
	VerificationCode  *string `json:"verification_code,omitempty"`
	IssuedDate        *string `json:"issued_date,omitempty"`
	ExpiryDate        *string `json:"expiry_date,omitempty"`
	ExamLocation      *string `json:"exam_location,omitempty"`
}

type CertificateVerificationDTO struct {
	CertificateNumber string  `json:"certificate_number"`
	ParticipantName   string  `json:"participant_name"`
	ParticipantCode   string  `json:"participant_code"`
	ExamTypeName      string  `json:"exam_type_name"`
	ExamDate          string  `json:"exam_date"`
	ExamLocation      string  `json:"exam_location"`
	ListeningScore    int     `json:"listening_score"`
	StructureScore    int     `json:"structure_score"`
	ReadingScore      int     `json:"reading_score"`
	TotalScore        int     `json:"total_score"`
	CEFRLevel         *string `json:"cefr_level,omitempty"`
	IssuedDate        string  `json:"issued_date"`
	ExpiryDate        string  `json:"expiry_date"`
	IsValid           bool    `json:"is_valid"`
	StatusMessage     string  `json:"status_message"`
}

// Question Bank Stimulus DTOs
type CreateStimulusItemOptionDTO struct {
	OptionLabel string  `json:"option_label" binding:"required"`
	OptionText  string  `json:"option_text" binding:"required"`
	MediaType   string  `json:"media_type"`
	MediaURL    *string `json:"media_url,omitempty"`
}

type CreateStimulusSubItemDTO struct {
	ItemOrder         int                           `json:"item_order"`
	QuestionText      string                        `json:"question_text" binding:"required"`
	QuestionMediaType string                        `json:"question_media_type"`
	QuestionMediaURL  *string                       `json:"question_media_url,omitempty"`
	CorrectAnswer     string                        `json:"correct_answer" binding:"required"`
	Explanation       *string                       `json:"explanation,omitempty"`
	Options           []CreateStimulusItemOptionDTO `json:"options" binding:"required"`
}

type CreateStimulusRequestDTO struct {
	KodeSoal        string                     `json:"kodesoal" binding:"required"`
	StimulusCode    string                     `json:"stimulus_code"`
	Title           string                     `json:"title"`
	NarrativeText   string                     `json:"narrative_text" binding:"required"`
	MediaType       string                     `json:"media_type"`
	MediaURL        *string                    `json:"media_url,omitempty"`
	CategoryTag     string                     `json:"category_tag"`
	DifficultyLevel int                        `json:"difficulty_level"`
	Items           []CreateStimulusSubItemDTO `json:"items" binding:"required"`
}

type UpdateStimulusRequestDTO struct {
	Title           string  `json:"title"`
	NarrativeText   string  `json:"narrative_text" binding:"required"`
	MediaType       string  `json:"media_type"`
	MediaURL        *string `json:"media_url,omitempty"`
	CategoryTag     string  `json:"category_tag"`
	DifficultyLevel int     `json:"difficulty_level"`
}

type CreateEPBankDTO struct {
	KodeSoal    string  `json:"kodesoal"`
	NamaSoal    string  `json:"namasoal" binding:"required"`
	Description *string `json:"description,omitempty"`
}

type UpdateEPBankDTO struct {
	NamaSoal    string  `json:"namasoal" binding:"required"`
	Description *string `json:"description,omitempty"`
}

type CopyEPBankDTO struct {
	NewKodeSoal string `json:"new_kodesoal" binding:"required"`
	NewNamaSoal string `json:"new_namasoal" binding:"required"`
}

// Anti-Cheat & Proctor Telemetry DTOs
type EPSecurityEventRequestDTO struct {
	ScheduleID int                    `json:"schedule_id" binding:"required"`
	EventType  string                 `json:"event_type" binding:"required"`
	Metadata   map[string]interface{} `json:"metadata"`
	DeviceID   string                 `json:"device_id"`
}

type EPSecurityEventResponseDTO struct {
	CurrentRiskScore    int    `json:"current_risk_score"`
	RiskLevel           string `json:"risk_level"`
	TabSwitchCount      int    `json:"tab_switch_count"`
	FullscreenExitCount int    `json:"fullscreen_exit_count"`
	TotalViolations     int    `json:"total_violations"`
	IsLocked            int    `json:"is_locked"`
	AutoLocked          bool   `json:"auto_locked"`
	MaxViolations       int    `json:"max_violations"`
}

type EPHeartbeatRequestDTO struct {
	ScheduleID      int  `json:"schedule_id" binding:"required"`
	IsFullscreen    bool `json:"is_fullscreen"`
	CurrentStimulus int  `json:"current_stimulus"`
	CurrentItem     int  `json:"current_item"`
}

type EPHeartbeatResponseDTO struct {
	ServerTime       string  `json:"server_time"`
	IsLocked         int     `json:"is_locked"`
	LockReason       *string `json:"lock_reason,omitempty"`
	ProctorWarning   *string `json:"proctor_warning,omitempty"`
	RemainingSeconds int     `json:"remaining_seconds"`
	SessionStatus    string  `json:"session_status"`
	CurrentSection   int     `json:"current_section"`
}

type EPProctorActionRequestDTO struct {
	ScheduleID      int    `json:"schedule_id" binding:"required"`
	ParticipantCode string `json:"participant_code"`
	KodePeserta     string `json:"kodepeserta"`
	Action          string `json:"action" binding:"required"` // SEND_WARNING, LOCK_SESSION, UNLOCK_SESSION, FORCE_SUBMIT
	Reason          string `json:"reason"`
	WarningMessage  string `json:"warning_message"`
}

type EPSecurityAuditItemDTO struct {
	ID           int64                  `json:"id"`
	IDEPSchedule int                    `json:"id_ep_schedule"`
	KodePeserta  string                 `json:"kodepeserta"`
	EventType    string                 `json:"event_type"`
	Severity     string                 `json:"severity"`
	RiskScore    int                    `json:"risk_score"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    string                 `json:"created_at"`
}

type CreateEPExternalParticipantRequestDTO struct {
	KodePeserta  string  `json:"kodepeserta"`
	NIK          *string `json:"nik"`
	Nama         string  `json:"nama" binding:"required"`
	JenisKelamin string  `json:"jenis_kelamin"`
	Email        string  `json:"email" binding:"required"`
	HP           string  `json:"hp" binding:"required"`
	Instansi     *string `json:"instansi"`
	TempatLahir  *string `json:"tempat_lahir"`
	TanggalLahir *string `json:"tanggal_lahir"`
	Alamat       *string `json:"alamat"`
	Password     string  `json:"password"`
}

type UpdateEPExternalParticipantRequestDTO struct {
	NIK          *string `json:"nik"`
	Nama         string  `json:"nama" binding:"required"`
	JenisKelamin string  `json:"jenis_kelamin"`
	Email        string  `json:"email" binding:"required"`
	HP           string  `json:"hp" binding:"required"`
	Instansi     *string `json:"instansi"`
	TempatLahir  *string `json:"tempat_lahir"`
	TanggalLahir *string `json:"tanggal_lahir"`
	Alamat       *string `json:"alamat"`
	Password     *string `json:"password"`
	IsActive     *bool   `json:"is_active"`
}

type EPExternalParticipantItemDTO struct {
	IDPesertaUmum int     `json:"id_peserta_umum"`
	KodePeserta   string  `json:"kodepeserta"`
	NIK           *string `json:"nik"`
	Nama          string  `json:"nama"`
	JenisKelamin  string  `json:"jenis_kelamin"`
	Email         string  `json:"email"`
	HP            string  `json:"hp"`
	Instansi      *string `json:"instansi"`
	TempatLahir   *string `json:"tempat_lahir"`
	TanggalLahir  *string `json:"tanggal_lahir"`
	Alamat        *string `json:"alamat"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// Proctor Snapshot Telemetry & Live Surveillance DTOs
type UploadProctorSnapshotRequestDTO struct {
	ScheduleID   int    `json:"schedule_id" binding:"required"`
	SnapshotType string `json:"snapshot_type"` // 'EXAM_START', 'PERIODIC', 'VIOLATION'
	TriggerEvent string `json:"trigger_event"` // 'EXAM_START', 'INTERVAL_TIMER', 'TAB_SWITCH', etc.
	ImageData    string `json:"image_data" binding:"required"`
	RiskScore    int    `json:"risk_score"`
}

type ProctorSnapshotItemDTO struct {
	ID           int64  `json:"id"`
	IDEPSchedule int    `json:"id_ep_schedule"`
	KodePeserta  string `json:"kodepeserta"`
	NamaPeserta  string `json:"nama_peserta"`
	SnapshotType string `json:"snapshot_type"`
	TriggerEvent string `json:"trigger_event"`
	ImageData    string `json:"image_data"`
	RiskScore    int    `json:"risk_score"`
	CreatedAt    string `json:"created_at"`
}

type ProctorLatestGridItemDTO struct {
	KodePeserta     string  `json:"kodepeserta"`
	NamaPeserta     string  `json:"nama_peserta"`
	LatestSnapshot  *string `json:"latest_snapshot,omitempty"`
	SnapshotType    *string `json:"snapshot_type,omitempty"`
	TriggerEvent    *string `json:"trigger_event,omitempty"`
	LastCapturedAt  *string `json:"last_captured_at,omitempty"`
	SessionStatus   string  `json:"session_status"`
	IsLocked        int     `json:"is_locked"`
	TotalViolations int     `json:"total_violations"`
	RiskLevel       string  `json:"risk_level"`
}
