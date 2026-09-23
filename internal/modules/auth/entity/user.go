package entity

import "time"

type User struct {
	UserID   int     `db:"userid"`
	Username string  `db:"username"`
	RealName string  `db:"userdesc"`
	Email    string  `db:"email"`
	Password string  `db:"password"`
	Salt     *string `db:"salt"`
	IsActive int     `db:"isactive"`
}

type UserRole struct {
	UserID   int    `db:"userid"`
	RoleID   string `db:"idrole"`
	RoleName string `db:"namarole"`
	UnitID   string `db:"idsatker"`
	UnitName string `db:"namaunit"`
}

type ModuleMenu struct {
	IDMenu     int     `db:"idmenu"`
	ParentMenu *int    `db:"parentmenu"`
	IDTarget   *string `db:"idtarget"`
	IDModul    string  `db:"idmodul"`
	NamaMenu   string  `db:"namamenu"`
	LevelMenu  int     `db:"levelmenu"`
	FAIcon     *string `db:"faicon"`
	NamaFile   *string `db:"namafile"`
}

type Peserta struct {
	KodePeserta string  `db:"kodepeserta"`
	Nama        string  `db:"nama"`
	Password    *string `db:"password"`
	Salt        *string `db:"salt"`
	Email       *string `db:"email"`
	HP          *string `db:"hp"`
	Alamat      *string `db:"alamat"`
	IsAktif     int     `db:"isaktif"`
	TokenLogin  *string `db:"tokenlogin"`
}

type SesiPeserta struct {
	IDJadwalUjian    int        `db:"idjadwalujian"`
	IDUjian          int        `db:"idujian"`
	NamaUjian        string     `db:"namaujian"`
	KodeSoal         string     `db:"kodesoal"`
	IDRuang          int        `db:"idruang"`
	NamaRuang        string     `db:"namaruang"`
	TglUjian         time.Time  `db:"tglujian"`
	TglSelesai       time.Time  `db:"tglselesai"`
	JamMulai         string     `db:"jammulai"`
	JamSelesai       string     `db:"jamselesai"`
	WaktuPengerjaan  int        `db:"waktupengerjaan"`
	TokenUjian       string     `db:"tokenujian"`
	Kapasitas        int        `db:"kapasitas"`
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
