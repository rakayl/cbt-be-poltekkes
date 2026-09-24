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
	IDUjian        int     `json:"idujian"`
	NamaUjian      string  `json:"namaujian"`
	IDPeriode      int     `json:"idperiode"`
	NamaPeriode    string  `json:"namaperiode"`
	IsOnline       bool    `json:"is_online"`
	Metode         string  `json:"metode"` // "Online / Daring" atau "Offline / Di Kampus"
	TglMulai       string  `json:"tglmulai"`
	TglSelesai     string  `json:"tglselesai"`
	NilaiMinimal   float64 `json:"nilaiminimal"`
	Keterangan     string  `json:"keterangan"`
	TotalSesi      int     `json:"totalsesi"`
	TotalKapasitas int     `json:"totalkapasitas"`
	JumlahPeserta  int     `json:"jumlahpeserta"`
	SisaKuota      int     `json:"sisakuota"`
	IsOpen         bool    `json:"is_open"`
}

type SPMBRegisterRequestDTO struct {
	IDUjian         int    `json:"idujian"`
	ExamID          int    `json:"exam_id"` // Fallback alias
	IDPendaftar     string `json:"idpendaftar" binding:"required"`
	NomorUjian      string `json:"nomor_ujian"`
	NoUjian         string `json:"noujian"` // Alias
	Nama            string `json:"nama" binding:"required"`
	NIK             string `json:"nik"`
	JK              string `json:"jk"` // L / P
	Email           string `json:"email"`
	HP              string `json:"hp"`
	Alamat          string `json:"alamat"`
	IDKota          *int   `json:"idkota"`
	Password        string `json:"password"`          // Optional, if empty defaults to participant code
	AutoPlotSession *bool  `json:"auto_plot_session"` // Default true
}

type SPMBCredentialsDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SPMBScheduleInfoDTO struct {
	IsPlotted       bool   `json:"is_plotted"`
	IDJadwalUjian   int    `json:"idjadwalujian,omitempty"`
	NamaRuang       string `json:"namaruang,omitempty"`
	TglUjian        string `json:"tglujian,omitempty"`
	JamMulai        string `json:"jammulai,omitempty"`
	JamSelesai      string `json:"jamselesai,omitempty"`
	WaktuPengerjaan int    `json:"waktupengerjaan,omitempty"`
	Catatan         string `json:"catatan,omitempty"`
}

type SPMBRegisterResponseDTO struct {
	KodePeserta    string              `json:"kodepeserta"`
	IDPendaftar    string              `json:"idpendaftar"`
	IDUjian        int                 `json:"idujian"`
	NamaUjian      string              `json:"namaujian"`
	IDPeriode      int                 `json:"idperiode"`
	NamaPeriode    string              `json:"namaperiode"`
	CBTCredentials SPMBCredentialsDTO  `json:"cbt_credentials"`
	Schedule       SPMBScheduleInfoDTO `json:"jadwal"`
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
	RequestBody    string    `json:"request_body"`
	ResponseBody   string    `json:"response_body"`
	CreatedAt      time.Time `json:"created_at"`
}
