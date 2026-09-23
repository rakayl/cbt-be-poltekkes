package entity

import "time"

// ExamType represents TOEFL_ITP or TOEIC_LR exam definition
type ExamType struct {
	IDExamType     int       `db:"id_exam_type" json:"id_exam_type"`
	TypeCode       string    `db:"type_code" json:"type_code"`
	TypeName       string    `db:"type_name" json:"type_name"`
	TotalSections  int       `db:"total_sections" json:"total_sections"`
	ScoreMin       int       `db:"score_min" json:"score_min"`
	ScoreMax       int       `db:"score_max" json:"score_max"`
	ScoreFormula   string    `db:"score_formula" json:"score_formula"`
	ValidityMonths int       `db:"validity_months" json:"validity_months"`
	IsActive       bool      `db:"is_active" json:"is_active"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// Section represents a standard section within an exam type
type Section struct {
	IDSection        int       `db:"id_section" json:"id_section"`
	IDExamType       int       `db:"id_exam_type" json:"id_exam_type"`
	SectionCode      string    `db:"section_code" json:"section_code"`
	SectionName      string    `db:"section_name" json:"section_name"`
	SectionOrder     int       `db:"section_order" json:"section_order"`
	DefaultQuestions int       `db:"default_questions" json:"default_questions"`
	DefaultDuration  int       `db:"default_duration" json:"default_duration"`
	HasAudio         bool      `db:"has_audio" json:"has_audio"`
	AllowReplay      bool      `db:"allow_replay" json:"allow_replay"`
	MaxReplayCount   int       `db:"max_replay_count" json:"max_replay_count"`
	CanGoBack        bool      `db:"can_go_back" json:"can_go_back"`
	ScoreScaleMin    int       `db:"score_scale_min" json:"score_scale_min"`
	ScoreScaleMax    int       `db:"score_scale_max" json:"score_scale_max"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

// SectionPart represents a sub-part (Part A, B, C) within a section
type SectionPart struct {
	IDPart          int       `db:"id_part" json:"id_part"`
	IDSection       int       `db:"id_section" json:"id_section"`
	PartCode        string    `db:"part_code" json:"part_code"`
	PartName        string    `db:"part_name" json:"part_name"`
	PartOrder       int       `db:"part_order" json:"part_order"`
	QuestionStart   int       `db:"question_start" json:"question_start"`
	QuestionEnd     int       `db:"question_end" json:"question_end"`
	InstructionText *string   `db:"instruction_text" json:"instruction_text"`
	HasPassage      bool      `db:"has_passage" json:"has_passage"`
	HasAudio        bool      `db:"has_audio" json:"has_audio"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

// ConversionProfile represents a score conversion table profile (ETS vs Custom)
type ConversionProfile struct {
	IDProfile   int       `db:"id_profile" json:"id_profile"`
	IDExamType  int       `db:"id_exam_type" json:"id_exam_type"`
	ProfileCode string    `db:"profile_code" json:"profile_code"`
	ProfileName string    `db:"profile_name" json:"profile_name"`
	IsDefault   bool      `db:"is_default" json:"is_default"`
	Description *string   `db:"description" json:"description"`
	CreatedBy   *string   `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// ScoreConversion maps raw_score -> scaled_score
type ScoreConversion struct {
	IDConversion int    `db:"id_conversion" json:"id_conversion"`
	IDProfile    int    `db:"id_profile" json:"id_profile"`
	SectionCode  string `db:"section_code" json:"section_code"`
	RawScore     int    `db:"raw_score" json:"raw_score"`
	ScaledScore  int    `db:"scaled_score" json:"scaled_score"`
}

// CEFRMapping maps total scaled score range to CEFR level
type CEFRMapping struct {
	IDMapping   int     `db:"id_mapping" json:"id_mapping"`
	IDExamType  int     `db:"id_exam_type" json:"id_exam_type"`
	CEFRLevel   string  `db:"cefr_level" json:"cefr_level"`
	ScoreMin    int     `db:"score_min" json:"score_min"`
	ScoreMax    int     `db:"score_max" json:"score_max"`
	Description *string `db:"description" json:"description"`
}

// CertificateTemplate represents official certificate layout
type CertificateTemplate struct {
	IDTemplate      int       `db:"id_template" json:"id_template"`
	IDExamType      int       `db:"id_exam_type" json:"id_exam_type"`
	TemplateName    string    `db:"template_name" json:"template_name"`
	LogoURL         *string   `db:"logo_url" json:"logo_url"`
	HeaderText      *string   `db:"header_text" json:"header_text"`
	InstitutionName *string   `db:"institution_name" json:"institution_name"`
	CertTitle       *string   `db:"cert_title" json:"cert_title"`
	SignatoryName   *string   `db:"signatory_name" json:"signatory_name"`
	SignatoryTitle  *string   `db:"signatory_title" json:"signatory_title"`
	SignatureURL    *string   `db:"signature_url" json:"signature_url"`
	StampURL        *string   `db:"stamp_url" json:"stamp_url"`
	IsActive        bool      `db:"is_active" json:"is_active"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

// EPExam represents an admin created TOEFL / TOEIC exam instance
type EPExam struct {
	IDEPExam               int       `db:"id_ep_exam" json:"id_ep_exam"`
	IDExamType             int       `db:"id_exam_type" json:"id_exam_type"`
	IDProfile              *int      `db:"id_profile" json:"id_profile"`
	IDTemplate             *int      `db:"id_template" json:"id_template"`
	ExamName               string    `db:"exam_name" json:"exam_name"`
	PassingScore           int       `db:"passing_score" json:"passing_score"`
	ValidityMonthsOverride *int      `db:"validity_months_override" json:"validity_months_override"`
	Description            *string   `db:"description" json:"description"`
	IsActive               bool      `db:"is_active" json:"is_active"`
	CreatedBy              *string   `db:"created_by" json:"created_by"`
	CreatedAt              time.Time `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time `db:"updated_at" json:"updated_at"`

	// Derived
	ExamTypeName string           `db:"exam_type_name" json:"exam_type_name,omitempty"`
	ProfileName  *string          `db:"profile_name" json:"profile_name,omitempty"`
	Sections     []*EPExamSection `json:"sections,omitempty"`
}

// EPExamSection represents dynamic section configuration and dynamic timer per exam
type EPExamSection struct {
	IDEPExamSection int       `db:"id_ep_exam_section" json:"id_ep_exam_section"`
	IDEPExam        int       `db:"id_ep_exam" json:"id_ep_exam"`
	IDSection       int       `db:"id_section" json:"id_section"`
	KodeSoal        string    `db:"kodesoal" json:"kodesoal"`
	SectionOrder    int       `db:"section_order" json:"section_order"`
	DurationMinutes int       `db:"duration_minutes" json:"duration_minutes"`
	AllowReplay     bool      `db:"allow_replay" json:"allow_replay"`
	MaxReplayCount  int       `db:"max_replay_count" json:"max_replay_count"`
	ShuffleOptions  bool      `db:"shuffle_options" json:"shuffle_options"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`

	// Derived
	SectionCode string `db:"section_code" json:"section_code,omitempty"`
	SectionName string `db:"section_name" json:"section_name,omitempty"`
	HasAudio    bool   `db:"has_audio" json:"has_audio,omitempty"`
}

// EPSchedule represents a scheduled session
type EPSchedule struct {
	IDEPSchedule int       `db:"id_ep_schedule" json:"id_ep_schedule"`
	IDEPExam     int       `db:"id_ep_exam" json:"id_ep_exam"`
	IDRuang      int       `db:"id_ruang" json:"id_ruang"`
	ExamDate     time.Time `db:"exam_date" json:"exam_date"`
	StartTime    string    `db:"start_time" json:"start_time"`
	EndTime      string    `db:"end_time" json:"end_time"`
	SessionToken string    `db:"session_token" json:"session_token"`
	Capacity     int       `db:"capacity" json:"capacity"`
	ProctorName  *string   `db:"proctor_name" json:"proctor_name"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`

	// Derived
	ExamName   string `db:"exam_name" json:"exam_name,omitempty"`
	RoomName   string `db:"room_name" json:"room_name,omitempty"`
	Registered int    `db:"registered" json:"registered,omitempty"`
}

// ScheduleParticipant represents a registered test taker with dynamic section state
type ScheduleParticipant struct {
	IDParticipant     int        `db:"id_participant" json:"id_participant"`
	IDEPSchedule      int        `db:"id_ep_schedule" json:"id_ep_schedule"`
	KodePeserta       string     `db:"kodepeserta" json:"kodepeserta"`
	SeatNumber        *int       `db:"seat_number" json:"seat_number"`
	StartedAt         *time.Time `db:"started_at" json:"started_at"`
	FinishedAt        *time.Time `db:"finished_at" json:"finished_at"`
	CurrentSection    int        `db:"current_section" json:"current_section"`
	SectionStartedAt  *time.Time `db:"section_started_at" json:"section_started_at"`
	SectionDeadlineAt *time.Time `db:"section_deadline_at" json:"section_deadline_at"`
	SessionStatus       string     `db:"session_status" json:"session_status"`
	IsLocked            bool       `db:"is_locked" json:"is_locked"`
	LockReason          *string    `db:"lock_reason" json:"lock_reason"`
	RiskScore           int        `db:"risk_score" json:"risk_score"`
	TabSwitchCount      int        `db:"tab_switch_count" json:"tab_switch_count"`
	FullscreenExitCount int        `db:"fullscreen_exit_count" json:"fullscreen_exit_count"`
	RiskLevel           string     `db:"risk_level" json:"risk_level"`
	TotalViolations     int        `db:"total_violations" json:"total_violations"`
	UnlockCount         int        `db:"unlock_count" json:"unlock_count"`
	ProctorWarning      *string    `db:"proctor_warning" json:"proctor_warning"`
	LastSeenAt          *time.Time `db:"last_seen_at" json:"last_seen_at"`
	SoftDelete          string     `db:"softdelete" json:"softdelete"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`

	// Derived
	Nama   string  `db:"nama" json:"nama,omitempty"`
	Email  *string `db:"email" json:"email,omitempty"`
	HP     *string `db:"hp" json:"hp,omitempty"`
	Alamat *string `db:"alamat" json:"alamat,omitempty"`
}

// EPSecurityEvent represents a recorded anti-cheat / proctor security event
type EPSecurityEvent struct {
	ID           int64     `db:"id" json:"id"`
	IDEPSchedule int       `db:"id_ep_schedule" json:"id_ep_schedule"`
	KodePeserta  string    `db:"kodepeserta" json:"kodepeserta"`
	EventType    string    `db:"event_type" json:"event_type"`
	Severity     string    `db:"severity" json:"severity"`
	RiskScore    int       `db:"risk_score" json:"risk_score"`
	Metadata     *string   `db:"metadata" json:"metadata"`
	IPAddress    *string   `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent    *string   `db:"user_agent" json:"user_agent,omitempty"`
	DeviceID     *string   `db:"device_id" json:"device_id,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	NamaPeserta  *string   `db:"nama_peserta" json:"nama_peserta,omitempty"`
}

// EPScore holds section breakdown and total scaled score
type EPScore struct {
	IDScore          int       `db:"id_score" json:"id_score"`
	IDParticipant    int       `db:"id_participant" json:"id_participant"`
	ListeningRaw     int       `db:"listening_raw" json:"listening_raw"`
	ListeningScaled  int       `db:"listening_scaled" json:"listening_scaled"`
	StructureRaw     int       `db:"structure_raw" json:"structure_raw"`
	StructureScaled  int       `db:"structure_scaled" json:"structure_scaled"`
	ReadingRaw       int       `db:"reading_raw" json:"reading_raw"`
	ReadingScaled    int       `db:"reading_scaled" json:"reading_scaled"`
	TotalScaledScore int       `db:"total_scaled_score" json:"total_scaled_score"`
	CEFRLevel        *string   `db:"cefr_level" json:"cefr_level"`
	PassingStatus    string    `db:"passing_status" json:"passing_status"`
	CalculatedAt     time.Time `db:"calculated_at" json:"calculated_at"`

	// Derived
	Nama        string `db:"nama" json:"nama,omitempty"`
	KodePeserta string `db:"kodepeserta" json:"kodepeserta,omitempty"`
	ExamName    string `db:"exam_name" json:"exam_name,omitempty"`
	RoomName    string `db:"room_name" json:"room_name,omitempty"`
	ExamDate    string `db:"exam_date" json:"exam_date,omitempty"`
	StartTime   string `db:"start_time" json:"start_time,omitempty"`
}

// EPCertificate represents official score report certificate
type EPCertificate struct {
	IDCertificate     int       `db:"id_certificate" json:"id_certificate"`
	IDScore           int       `db:"id_score" json:"id_score"`
	CertificateNumber string    `db:"certificate_number" json:"certificate_number"`
	ParticipantName   string    `db:"participant_name" json:"participant_name"`
	ParticipantCode   string    `db:"participant_code" json:"participant_code"`
	ExamTypeName      string    `db:"exam_type_name" json:"exam_type_name"`
	ExamDate          time.Time `db:"exam_date" json:"exam_date"`
	ExamLocation      string    `db:"exam_location" json:"exam_location"`
	ListeningScore    int       `db:"listening_score" json:"listening_score"`
	StructureScore    int       `db:"structure_score" json:"structure_score"`
	ReadingScore      int       `db:"reading_score" json:"reading_score"`
	TotalScore        int       `db:"total_score" json:"total_score"`
	CEFRLevel         *string   `db:"cefr_level" json:"cefr_level"`
	IssuedDate        time.Time `db:"issued_date" json:"issued_date"`
	ExpiryDate        time.Time `db:"expiry_date" json:"expiry_date"`
	IsRevoked         bool      `db:"is_revoked" json:"is_revoked"`
	RevokedReason     *string   `db:"revoked_reason" json:"revoked_reason"`
	VerificationCode  string    `db:"verification_code" json:"verification_code"`
	QRCodeURL         *string   `db:"qr_code_url" json:"qr_code_url"`
	PDFURL            *string   `db:"pdf_url" json:"pdf_url"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// EPQuestionBank represents question bank header for TOEFL / TOEIC
type EPQuestionBank struct {
	IDEPBank       int64     `db:"id_ep_bank" json:"id_ep_bank"`
	KodeSoal       string    `db:"kodesoal" json:"kodesoal"`
	NamaSoal       string    `db:"namasoal" json:"namasoal"`
	IDExamType     int       `db:"id_exam_type" json:"id_exam_type"`
	ExamTypeName   string    `db:"exam_type_name" json:"exam_type_name,omitempty"`
	IDSection      int       `db:"id_section" json:"id_section"`
	SectionName    string    `db:"section_name" json:"section_name,omitempty"`
	SectionCode    string    `db:"section_code" json:"section_code,omitempty"`
	Description    *string   `db:"description" json:"description"`
	TotalStimuli   int       `db:"total_stimuli" json:"total_stimuli"`
	TotalQuestions int       `db:"total_questions" json:"total_questions"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
