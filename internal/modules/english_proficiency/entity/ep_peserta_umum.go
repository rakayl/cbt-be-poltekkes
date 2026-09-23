package entity

import "time"

// EPPesertaUmum represents master data of external/general test-takers for TOEFL/EPT
type EPPesertaUmum struct {
	IDPesertaUmum int        `db:"id_peserta_umum" json:"id_peserta_umum"`
	KodePeserta   string     `db:"kodepeserta" json:"kodepeserta"`
	NIK           *string    `db:"nik" json:"nik"`
	Nama          string     `db:"nama" json:"nama"`
	JenisKelamin  string     `db:"jenis_kelamin" json:"jenis_kelamin"`
	Email         string     `db:"email" json:"email"`
	HP            string     `db:"hp" json:"hp"`
	Instansi      *string    `db:"instansi" json:"instansi"`
	TempatLahir   *string    `db:"tempat_lahir" json:"tempat_lahir"`
	TanggalLahir  *time.Time `db:"tanggal_lahir" json:"tanggal_lahir"`
	Alamat        *string    `db:"alamat" json:"alamat"`
	Password      string     `db:"password" json:"-"`
	IsActive      bool       `db:"is_active" json:"is_active"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
