package dto

type RuangDTO struct {
	KodeRuang  string  `json:"kode_ruang"`
	NamaRuang  string  `json:"nama_ruang"`
	Kapasitas  int     `json:"kapasitas"`
	Keterangan *string `json:"keterangan,omitempty"`
}

type SkorDTO struct {
	KodeSkor  string  `json:"kode_skor"`
	SkorBenar float64 `json:"skor_benar"`
	SkorSalah float64 `json:"skor_salah"`
}

type JenisUjianDTO struct {
	KodeJenis string `json:"kode_jenis"`
	NamaJenis string `json:"nama_jenis"`
}

type WilayahDTO struct {
	IDWilayah     string  `json:"id_wilayah"`
	ParentWilayah *string `json:"parent_wilayah,omitempty"`
	NamaWilayah   string  `json:"nama_wilayah"`
	Level         int     `json:"level"`
}

type PanitiaDTO struct {
	NIP         string `json:"nip"`
	NamaPanitia string `json:"nama_panitia"`
}

type JenisPeriodeDTO struct {
	JenisPeriode     string `json:"jenis_periode"`
	KodeJenisPeriode string `json:"kode_jenis_periode"`
	NamaJenisPeriode string `json:"nama_jenis_periode"`
}

type SalinSoalRequestDTO struct {
	SumberKodeSoal string `json:"sumber_kode_soal" binding:"required"`
	TargetKodeSoal string `json:"target_kode_soal" binding:"required"`
	TargetNamaSoal string `json:"target_nama_soal" binding:"required"`
}

type SalinSoalResponseDTO struct {
	SumberKodeSoal string `json:"sumber_kode_soal"`
	TargetKodeSoal string `json:"target_kode_soal"`
	TotalDisalin   int    `json:"total_pertanyaan_disalin"`
}

type UnitDTO struct {
	IDSatker    string `json:"id_satker"`
	NamaSatker  string `json:"nama_satker"`
	NamaSingkat string `json:"nama_singkat"`
}

type CreateRuangRequestDTO struct {
	KodeRuang  string `json:"kode_ruang"`
	NamaRuang  string `json:"nama_ruang" binding:"required"`
	Kapasitas  int    `json:"kapasitas" binding:"required"`
	Keterangan string `json:"keterangan"`
}

type UpdateRuangRequestDTO struct {
	NamaRuang  string `json:"nama_ruang" binding:"required"`
	Kapasitas  int    `json:"kapasitas" binding:"required"`
	Keterangan string `json:"keterangan"`
}

type CreateJenisUjianRequestDTO struct {
	KodeJenis string `json:"kode_jenis"`
	NamaJenis string `json:"nama_jenis" binding:"required"`
}

type UpdateJenisUjianRequestDTO struct {
	NamaJenis string `json:"nama_jenis" binding:"required"`
}

type CreatePanitiaRequestDTO struct {
	NIP         string `json:"nip" binding:"required"`
	NamaPanitia string `json:"nama_panitia" binding:"required"`
}

type UpdatePanitiaRequestDTO struct {
	NamaPanitia string `json:"nama_panitia" binding:"required"`
}

type CreateJenisPeriodeRequestDTO struct {
	JenisPeriode     string `json:"jenis_periode" binding:"required"`
	NamaJenisPeriode string `json:"nama_jenis_periode" binding:"required"`
}

type UpdateJenisPeriodeRequestDTO struct {
	NamaJenisPeriode string `json:"nama_jenis_periode" binding:"required"`
}
