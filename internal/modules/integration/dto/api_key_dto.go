package dto

import "time"

// =========================================================================
// API Key Management DTOs
// =========================================================================

type CreateAPIKeyRequestDTO struct {
	Name        string `json:"name" binding:"required"`
	APIKey      string `json:"api_key"`      // Optional, if empty will be auto-generated
	IPWhitelist string `json:"ip_whitelist"` // Optional, comma-separated IPs or CIDR
	Description string `json:"description"`
}

type UpdateAPIKeyRequestDTO struct {
	Name        string `json:"name" binding:"required"`
	IPWhitelist string `json:"ip_whitelist"`
	IsActive    *bool  `json:"is_active"`
	Description string `json:"description"`
}

type APIKeyResponseDTO struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	APIKey      string     `json:"api_key"`
	IPWhitelist string     `json:"ip_whitelist"`
	IsActive    bool       `json:"is_active"`
	Description string     `json:"description"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// =========================================================================
// SPMB Integration DTOs
// =========================================================================

type SPMBActiveExamDTO struct {
	ExamID          int     `json:"exam_id"`
	ExamName        string  `json:"exam_name"`
	PeriodID        int     `json:"period_id"`
	PeriodName      string  `json:"period_name"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	PassingGrade    float64 `json:"passing_grade"`
	Description     string  `json:"description"`
	TotalSessions   int     `json:"total_sessions"`
	TotalCapacity   int     `json:"total_capacity"`
	RegisteredCount int     `json:"registered_count"`
	RemainingQuota  int     `json:"remaining_quota"`
	IsOpen          bool    `json:"is_open"`
}

type SPMBRegisterRequestDTO struct {
	ExamID          int    `json:"exam_id" binding:"required"`
	IDPendaftar     string `json:"idpendaftar" binding:"required"`
	NomorUjian      string `json:"nomor_ujian"` // Optional, if empty can be generated or equal to idpendaftar
	Nama            string `json:"nama" binding:"required"`
	NIK             string `json:"nik"`
	JK              string `json:"jk"` // L / P
	Email           string `json:"email"`
	HP              string `json:"hp"`
	Alamat          string `json:"alamat"`
	IDKota          *int   `json:"idkota"`
	Password        string `json:"password"`          // Optional, if empty defaults to participant code
	AutoPlotSession bool   `json:"auto_plot_session"` // Default true
}

type SPMBCredentialsDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SPMBScheduleInfoDTO struct {
	IsPlotted       bool   `json:"is_plotted"`
	SessionID       int    `json:"session_id,omitempty"`
	RoomName        string `json:"room_name,omitempty"`
	ExamDate        string `json:"exam_date,omitempty"`
	StartTime       string `json:"start_time,omitempty"`
	EndTime         string `json:"end_time,omitempty"`
	DurationMinutes int    `json:"duration_minutes,omitempty"`
}

type SPMBRegisterResponseDTO struct {
	ParticipantCode string              `json:"participant_code"`
	IDPendaftar     string              `json:"idpendaftar"`
	ExamID          int                 `json:"exam_id"`
	ExamName        string              `json:"exam_name"`
	PeriodID        int                 `json:"period_id"`
	PeriodName      string              `json:"period_name"`
	CBTCredentials  SPMBCredentialsDTO  `json:"cbt_credentials"`
	Schedule        SPMBScheduleInfoDTO `json:"schedule"`
}

// =========================================================================
// API Access Audit Log DTOs
// =========================================================================

type APIAccessLogResponseDTO struct {
	ID             int64     `json:"id"`
	APIKeyID       *int      `json:"api_key_id"`
	ClientName     string    `json:"client_name"`
	IPAddress      string    `json:"ip_address"`
	Method         string    `json:"method"`
	Endpoint       string    `json:"endpoint"`
	StatusCode     int       `json:"status_code"`
	ResponseTimeMS int       `json:"response_time_ms"`
	UserAgent      string    `json:"user_agent"`
	ErrorMessage   string    `json:"error_message"`
	CreatedAt      time.Time `json:"created_at"`
}
