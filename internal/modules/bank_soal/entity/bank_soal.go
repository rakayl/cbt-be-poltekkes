package entity

import "time"

type SoalHeader struct {
	KodeSoal       string    `db:"kodesoal"`
	NamaSoal       string    `db:"namasoal"`
	Keterangan     *string   `db:"keterangan"`
	TotalQuestions       int       `db:"total_questions"`
	TotalQuestionsStatic int       `db:"total_questions_static"`
	TotalStimuli         int       `db:"total_stimuli"`
	TInsertTime          time.Time `db:"t_inserttime"`
	SoftDelete           string    `db:"softdelete"`
}

type Pertanyaan struct {
	KodeSoal        string  `db:"kodesoal"`
	NoUrut          int     `db:"nourut"`
	JenisJawaban    string  `db:"jenisjawaban"`
	Pertanyaan      string  `db:"pertanyaan"`
	PertanyaanImage *string `db:"pertanyaanimage"`
	PertanyaanAudio *string `db:"pertanyaanaudio"`
	PertanyaanVideo *string `db:"pertanyaanvideo"`
	Jawaban1        *string `db:"jawaban1"`
	Jawaban1Image   *string `db:"jawaban1image"`
	Jawaban1Audio   *string `db:"jawaban1audio"`
	Jawaban1Video   *string `db:"jawaban1video"`
	Jawaban2        *string `db:"jawaban2"`
	Jawaban2Image   *string `db:"jawaban2image"`
	Jawaban2Audio   *string `db:"jawaban2audio"`
	Jawaban2Video   *string `db:"jawaban2video"`
	Jawaban3        *string `db:"jawaban3"`
	Jawaban3Image   *string `db:"jawaban3image"`
	Jawaban3Audio   *string `db:"jawaban3audio"`
	Jawaban3Video   *string `db:"jawaban3video"`
	Jawaban4        *string `db:"jawaban4"`
	Jawaban4Image   *string `db:"jawaban4image"`
	Jawaban4Audio   *string `db:"jawaban4audio"`
	Jawaban4Video   *string `db:"jawaban4video"`
	Jawaban5        *string `db:"jawaban5"`
	Jawaban5Image   *string `db:"jawaban5image"`
	Jawaban5Audio   *string `db:"jawaban5audio"`
	Jawaban5Video   *string `db:"jawaban5video"`
	JawabanBenar    int     `db:"jawabanbenar"`
	Bobot           float64 `db:"bobot"`
	BobotBenar      float64 `db:"bobot_benar"`
	BobotSalah      float64 `db:"bobot_salah"`
	KataKunci       *string `db:"katakunci"`
	SoftDelete      string  `db:"softdelete"`
}
