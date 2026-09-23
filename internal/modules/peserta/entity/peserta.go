package entity

type Peserta struct {
	KodePeserta    string  `db:"kodepeserta" json:"kodepeserta"`
	IDPeriode      int     `db:"idperiode" json:"idperiode"`
	NamaPeriode    string  `db:"namaperiode" json:"namaperiode"`
	Nama           string  `db:"nama" json:"nama"`
	JK             string  `db:"jk" json:"jk"`
	HP             *string `db:"hp" json:"hp,omitempty"`
	Email          *string `db:"email" json:"email,omitempty"`
	ExtImage       *string `db:"extimage" json:"extimage,omitempty"`
	Password       *string `db:"password" json:"-"`
	Salt           *string `db:"salt" json:"-"`
	TokenLogin     *string `db:"tokenlogin" json:"tokenlogin,omitempty"`
	SumberData     *string `db:"sumberdata" json:"sumberdata,omitempty"`
	Alamat         *string `db:"alamat" json:"alamat,omitempty"`
	IDKota         *int    `db:"idkota" json:"idkota,omitempty"`
	IDKotaLama     *string `db:"idkota_lama" json:"idkota_lama,omitempty"`
	NamaWilayah    string  `db:"namawilayah" json:"namawilayah"`
	Profesi        *string `db:"profesi" json:"profesi,omitempty"`
	Status         *string `db:"status" json:"status,omitempty"`
	FilePembayaran *string `db:"filepembayaran" json:"filepembayaran,omitempty"`
	Hint           *string `db:"hint" json:"hint,omitempty"`
	IsValid        int     `db:"isvalid" json:"isvalid"`
	IsAktif        int     `db:"isaktif" json:"isaktif"`
	IsLogin        int     `db:"islogin" json:"islogin"`
	Keterangan     *string `db:"keterangan" json:"keterangan,omitempty"`
	KodeReferensi  *string `db:"kodereferensi" json:"kodereferensi,omitempty"`
	SoftDelete     string  `db:"softdelete" json:"softdelete"`
}
