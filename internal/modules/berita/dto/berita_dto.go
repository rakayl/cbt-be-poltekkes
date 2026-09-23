package dto

import "time"

type CreateBeritaRequestDTO struct {
	JudulBerita string `json:"judul_berita" binding:"required,min=5"`
	IsiBerita   string `json:"isi_berita" binding:"required"`
	Khusus      int    `json:"khusus"`
}

type BeritaResponseDTO struct {
	IDBerita    int        `json:"id_berita"`
	JudulBerita string     `json:"judul_berita"`
	IsiBerita   string     `json:"isi_berita"`
	WaktuRilis  *time.Time `json:"waktu_rilis,omitempty"`
}
