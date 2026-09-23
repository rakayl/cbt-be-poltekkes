package entity

import "time"

type Berita struct {
	IDBerita    int        `db:"idberita"`
	JudulBerita string     `db:"judulberita"`
	IsiBerita   string     `db:"isiberita"`
	ExtImage    *string    `db:"extimage"`
	Khusus      *int       `db:"khusus"`
	TInsertTime *time.Time `db:"t_inserttime"`
	SoftDelete  string     `db:"softdelete"`
}
