package entity

import "time"

type Ujian struct {
	IDUjian       int        `db:"idujian"`
	NamaUjian     string     `db:"namaujian"`
	IDPeriode     int        `db:"idperiode"`
	NamaPeriode   string     `db:"namaperiode"`
	NilaiMinimal  *float64   `db:"nilaiminimal"`
	IDSatker      string     `db:"idsatker"`
	NamaSatker    string     `db:"namasatker"`
	KodeJenis     string     `db:"kodejenis"`
	NamaJenis     string     `db:"namajenis"`
	IsPercobaan   int        `db:"ispercobaan"`
	Keterangan    *string    `db:"keterangan"`
	SoftDelete    string     `db:"softdelete"`
	MaxViolations int        `db:"max_violations"`
	TUpdateTime   *time.Time `db:"t_updatetime"`
	TUpdateUser   *string    `db:"t_updateuser"`
}

type ExamParticipant struct {
	KodePeserta        string     `db:"kodepeserta"`
	Nama               string     `db:"nama"`
	Email              *string    `db:"email"`
	NamaWilayah        *string    `db:"namawilayah"`
	Nilai              *float64   `db:"nilai"`
	IsLoggedIn         int        `db:"isloggedin"`
	IsPlotted          int        `db:"is_plotted"`
	SessionNumber      *int       `db:"session_number"`
	RoomName           *string    `db:"room_name"`
	TglMulai           *string    `db:"tgl_mulai"`
	WaktuMulai         *string    `db:"waktu_mulai"`
	WaktuSelesai       *string    `db:"waktu_selesai"`
	WaktuPengerjaanStr *string    `db:"waktupengerjaan_str"`
	IsVerified         int        `db:"is_verified"`
	BarcodeScannedAt   *time.Time `db:"barcode_scanned_at"`
	ScannedBy          *string    `db:"scanned_by"`
}

type JadwalUjian struct {
	IDJadwalUjian   int       `db:"idjadwalujian"`
	IDUjian         int       `db:"idujian"`
	NamaUjian       string    `db:"namaujian"`
	IDPeriode       int       `db:"idperiode"`
	NamaPeriode     string    `db:"namaperiode"`
	KodeSoal        string    `db:"kodesoal"`
	IDRuang         int       `db:"idruang"`
	NamaRuang       string    `db:"namaruang"`
	TglUjian        time.Time `db:"tglujian"`
	JamMulai        string    `db:"jammulai"`
	JamSelesai      string    `db:"jamselesai"`
	WaktuPengerjaan int       `db:"waktupengerjaan"`
	TokenUjian      string    `db:"tokenujian"`
	Kapasitas       int       `db:"kapasitas"`
	MaxViolations   int       `db:"max_violations"`
}

type LiveMonitorRow struct {
	KodePeserta               string     `db:"kodepeserta"`
	Nama                      string     `db:"nama"`
	NoUrutPeserta             int        `db:"nourutpeserta"`
	StatusUjian               string     `db:"statusujian"`
	TotalQuestions            int        `db:"total_questions"`
	JumlahTerjawab            int        `db:"jumlahterjawab"`
	RiskScore                 int        `db:"risk_score"`
	RiskLevel                 string     `db:"risk_level"`
	TabSwitchCount            int        `db:"tab_switch_count"`
	FullscreenExitCount       int        `db:"fullscreen_exit_count"`
	ActiveTabSwitchCount      int        `db:"active_tab_switch_count"`
	ActiveFullscreenExitCount int        `db:"active_fullscreen_exit_count"`
	TotalViolations           int        `db:"total_violations"`
	UnlockCount               int        `db:"unlock_count"`
	DisconnectCount           int        `db:"disconnect_count"`
	ReconnectCount            int        `db:"reconnect_count"`
	IsLocked                  int        `db:"is_locked"`
	LockReason                *string    `db:"lock_reason"`
	ProctorWarning            *string    `db:"proctor_warning"`
	LastHeartbeat             *time.Time `db:"last_heartbeat"`
	WaktuMulai                *time.Time `db:"waktumulai"`
	WaktuSelesai              *time.Time `db:"waktuselesai"`
	ClientDeviceID            *string    `db:"client_device_id"`
	IsVerified                int        `db:"is_verified"`
	BarcodeScannedAt          *time.Time `db:"barcode_scanned_at"`
}

type SecurityEvent struct {
	ID          int64     `db:"id"`
	IDJadwal    int       `db:"idjadwal"`
	KodePeserta string    `db:"kodepeserta"`
	EventType   string    `db:"event_type"`
	Severity    string    `db:"severity"`
	RiskScore   int       `db:"risk_score"`
	Metadata    *string   `db:"metadata"`
	IPAddress   *string   `db:"ip_address"`
	UserAgent   *string   `db:"user_agent"`
	DeviceID    *string   `db:"device_id"`
	CreatedAt   time.Time `db:"created_at"`
}

type ParticipantSessionExt struct {
	IDSession          int64      `db:"id_session" json:"id_session"`
	IDJadwalUjian      int        `db:"idjadwalujian" json:"idjadwalujian"`
	KodePeserta        string     `db:"kodepeserta" json:"kodepeserta"`
	SessionStatus      string     `db:"session_status" json:"session_status"`
	StartedAt          *time.Time `db:"started_at" json:"started_at"`
	CompletedAt        *time.Time `db:"completed_at" json:"completed_at"`
	CalculatedDeadline *time.Time `db:"calculated_deadline" json:"calculated_deadline"`
	SessionSnapshot    []byte     `db:"session_snapshot" json:"session_snapshot"`
	TotalScore         float64    `db:"total_score" json:"total_score"`
	TotalCorrect       int        `db:"total_correct" json:"total_correct"`
	TotalWrong         int        `db:"total_wrong" json:"total_wrong"`
	TotalUnanswered    int        `db:"total_unanswered" json:"total_unanswered"`
	PassStatus         string     `db:"pass_status" json:"pass_status"`
	SubmitReason       *string    `db:"submit_reason" json:"submit_reason"`
	ClientIP           *string    `db:"client_ip" json:"client_ip"`
	UserAgent          *string    `db:"user_agent" json:"user_agent"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

// JadwalSesi represents a row from cat.at_jadwalujian for admin CRUD
type JadwalSesi struct {
	IDJadwalUjian      int     `db:"idjadwalujian"`
	IDUjian            int     `db:"idujian"`
	NoJadwal           int     `db:"nojadwal"`
	WaktuPengerjaan    string  `db:"waktupengerjaan"` // stored as char(4), e.g. "0060"
	Bobot              float64 `db:"bobot"`
	TampilkanNilai     int     `db:"tampilkannilai"`
	ExamDeadlineAt     *string `db:"exam_deadline_at"`
	GracePeriodSeconds int     `db:"grace_period_seconds"`
	RoomCount          int     `db:"room_count"`
	WaktuMulai         *string `db:"waktu_mulai"`
	WaktuSelesai       *string `db:"waktu_selesai"`
	TglMulai           *string `db:"tgl_mulai"`
	TglSelesai         *string `db:"tgl_selesai"`
	TokenUjian         string  `db:"token_ujian"`
	MaxViolations      int     `db:"max_violations"`
}

// RuangUjian represents a row from cat.at_ruangujian for admin CRUD
type RuangUjian struct {
	IDRuangUjian       int    `db:"idruangujian"`
	IDJadwalUjian      int    `db:"idjadwalujian"`
	KodeRuang          string `db:"koderuang"`
	NamaRuang          string `db:"namaruang"`
	TglMulai           string `db:"tglmulai"`
	TglSelesai         string `db:"tglselesai"`
	WaktuMulai         string `db:"waktumulai_str"`
	WaktuSelesai       string `db:"waktuselesai_str"`
	WaktuPengerjaanStr string `db:"waktupengerjaan_str"`
	JumlahPeserta      int    `db:"jumlahpeserta"`
	Prioritas          int    `db:"prioritas"`
}

// AvailableParticipant represents a peserta that can be invited to an exam
type AvailableParticipant struct {
	KodePeserta string  `db:"kodepeserta"`
	Nama        string  `db:"nama"`
	Email       *string `db:"email"`
	HP          *string `db:"hp"`
}

// RoomParticipantRow represents a participant assigned to a specific room & session
type RoomParticipantRow struct {
	KodePeserta      string     `db:"kodepeserta"`
	Nama             string     `db:"nama"`
	Email            *string    `db:"email"`
	NamaWilayah      *string    `db:"namawilayah"`
	IDJadwalUjian    int        `db:"idjadwalujian"`
	IDRuangUjian     int        `db:"idruangujian"`
	NamaRuang        string     `db:"namaruang"`
	IsLogin          int        `db:"islogin"`
	IsVerified       int        `db:"is_verified"`
	BarcodeScannedAt *time.Time `db:"barcode_scanned_at"`
}

// SoalUjian represents a row from cat.at_soalujian linking session to bank soal
type SoalUjian struct {
	IDJadwalUjian          int     `db:"idjadwalujian"`
	KodeSoal               string  `db:"kodesoal"`
	NamaSoal               string  `db:"namasoal"`
	NoUrut                 int     `db:"nourut"`
	JumlahPertanyaan       int     `db:"jumlahpertanyaan"`
	JumlahSoalStatis       int     `db:"jumlah_soal_statis"`
	JumlahKasusDinamis     int     `db:"jumlah_kasus_dinamis"`
	Bobot                  float64 `db:"bobot"`
	KodeSkor               string  `db:"kodeskor"`
	TotalAvailable         int     `db:"total_available"`
	TotalStatisAvailable   int     `db:"total_statis_available"`
	TotalStimulusAvailable int     `db:"total_stimulus_available"`
}
