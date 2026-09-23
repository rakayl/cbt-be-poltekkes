package dto

type PesertaResponseDTO struct {
	KodePeserta     string  `json:"kodepeserta"`
	ParticipantCode string  `json:"participant_code"`
	IDPeriode       int     `json:"idperiode"`
	NamaPeriode     string  `json:"namaperiode"`
	Nama            string  `json:"nama"`
	Name            string  `json:"name"`
	JK              string  `json:"jk"`
	Email           *string `json:"email"`
	HP              *string `json:"hp"`
	Phone           *string `json:"phone"`
	Alamat          *string `json:"alamat,omitempty"`
	Address         *string `json:"address,omitempty"`
	IDKota          *int    `json:"idkota"`
	NamaWilayah     string  `json:"namawilayah"`
	Profesi         *string `json:"profesi,omitempty"`
	Status          *string `json:"status,omitempty"`
	SumberData      *string `json:"sumberdata,omitempty"`
	KodeReferensi   *string `json:"kodereferensi,omitempty"`
	Hint            *string `json:"hint,omitempty"`
	IsValid         int     `json:"isvalid"`
	IsAktif         int     `json:"isaktif"`
	IsActive        int     `json:"is_active"`
	IsLogin         int     `json:"islogin"`
	FilePembayaran  *string `json:"filepembayaran,omitempty"`
	Keterangan      *string `json:"keterangan,omitempty"`
}

type SavePesertaRequestDTO struct {
	IDPeriode      int     `json:"idperiode" binding:"required"`
	KodePeserta    string  `json:"kodepeserta" binding:"required"`
	Nama           string  `json:"nama" binding:"required"`
	JK             string  `json:"jk" binding:"required"`
	HP             *string `json:"hp"`
	Email          *string `json:"email"`
	Alamat         *string `json:"alamat"`
	IDKota         *int    `json:"idkota"`
	NamaWilayah    *string `json:"namawilayah"`
	Profesi        *string `json:"profesi"`
	Status         *string `json:"status"`
	SumberData     *string `json:"sumberdata"`
	KodeReferensi  *string `json:"kodereferensi"`
	Password       *string `json:"password"`
	Hint           *string `json:"hint"`
	IsValid        int     `json:"isvalid"`
	IsAktif        int     `json:"isaktif"`
	FilePembayaran *string `json:"filepembayaran"`
	Keterangan     *string `json:"keterangan"`
}

type ChangePasswordRequestDTO struct {
	Password string `json:"password" binding:"required,min=4"`
}

type ResetParticipantLoginDTO struct {
	KodePeserta string `json:"participant_code"`
	Code        string `json:"kodepeserta"`
}
