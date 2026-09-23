package entity

import "time"

type PesertaNilai struct {
	IDJadwalUjian int        `db:"idjadwalujian"`
	KodePeserta   string     `db:"kodepeserta"`
	Nama          string     `db:"nama"`
	NoUrutPeserta int        `db:"nourutpeserta"`
	TglMulai      *time.Time `db:"tglmulai"`
	TglSelesai    *time.Time `db:"tglselesai"`
	StatusUjian   string     `db:"statusujian"`
	Nilai         float64    `db:"nilai"`
	StatusLulus   string     `db:"statuslulus"`
	Ranking       int        `db:"ranking"`
}
