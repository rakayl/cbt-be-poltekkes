package entity

type Periode struct {
	IDPeriode        int     `db:"idperiode"`
	NamaPeriode      string  `db:"namaperiode"`
	JenisPeriode     string  `db:"jenisperiode"`
	IsOnline         int     `db:"isonline"`
	NIP              *string `db:"nip"`
	DigitKodePeserta *string `db:"digitkodepeserta"`
	SoftDelete       string  `db:"softdelete"`
}
