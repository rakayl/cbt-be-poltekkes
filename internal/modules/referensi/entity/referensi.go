package entity

type Ruang struct {
	KodeRuang  string  `db:"koderuang"`
	NamaRuang  string  `db:"namaruang"`
	Kapasitas  int     `db:"kapasitas"`
	Keterangan *string `db:"keterangan"`
	SoftDelete string  `db:"softdelete"`
}

type Skor struct {
	KodeSkor   string  `db:"kodeskor"`
	SkorBenar  float64 `db:"skorbenar"`
	SkorSalah  float64 `db:"skorsalah"`
	SoftDelete string  `db:"softdelete"`
}

type JenisUjian struct {
	KodeJenis  string `db:"kodejenis"`
	NamaJenis  string `db:"namajenis"`
	SoftDelete string `db:"softdelete"`
}

type Wilayah struct {
	IDWilayah     string  `db:"idwilayah"`
	ParentWilayah *string `db:"parentwilayah"`
	NamaWilayah   string  `db:"namawilayah"`
	Level         int     `db:"level"`
	SoftDelete    string  `db:"softdelete"`
}

type Panitia struct {
	NIP         string `db:"nip"`
	NamaPanitia string `db:"namapanitia"`
	SoftDelete  string `db:"softdelete"`
}

type JenisPeriode struct {
	JenisPeriode     string `db:"jenisperiode"`
	NamaJenisPeriode string `db:"namajenisperiode"`
	SoftDelete       string `db:"softdelete"`
}

type Unit struct {
	IDSatker    string `db:"idsatker"`
	NamaSatker  string `db:"namasatker"`
	NamaSingkat string `db:"namasingkat"`
}
