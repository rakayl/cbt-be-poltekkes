package repository

import (
	"context"
	"fmt"
	"poltekkes-cat-backend/internal/modules/bank_soal/dto"
	"poltekkes-cat-backend/internal/modules/bank_soal/entity"

	"github.com/jmoiron/sqlx"
)

type BankSoalRepository interface {
	GetQuestionBanks(ctx context.Context, page, perPage int, search string) ([]*entity.SoalHeader, int64, error)
	GetQuestionBankByCode(ctx context.Context, code string) (*entity.SoalHeader, error)
	GetQuestionsByCode(ctx context.Context, code string) ([]*entity.Pertanyaan, error)
	GetQuestionByNoUrut(ctx context.Context, code string, noUrut int) (*entity.Pertanyaan, error)
	CreateQuestionBank(ctx context.Context, code, name, desc string) error
	UpdateQuestionBank(ctx context.Context, code, name, desc string) error
	DeleteQuestionBank(ctx context.Context, code string) error
	CopyQuestionBank(ctx context.Context, sourceCode, newCode, newName string) error
	CreateQuestionItem(ctx context.Context, code string, req *dto.SaveQuestionItemDTO) (*entity.Pertanyaan, error)
	UpdateQuestionItem(ctx context.Context, code string, noUrut int, req *dto.SaveQuestionItemDTO) error
	DeleteQuestionItem(ctx context.Context, code string, noUrut int) error
}

type bankSoalRepository struct {
	db *sqlx.DB
}

func NewBankSoalRepository(db *sqlx.DB) BankSoalRepository {
	return &bankSoalRepository{db: db}
}

func (r *bankSoalRepository) GetQuestionBanks(ctx context.Context, page, perPage int, search string) ([]*entity.SoalHeader, int64, error) {
	offset := (page - 1) * perPage
	searchPattern := "%" + search + "%"

	var total int64
	countQuery := `
		SELECT COUNT(*) FROM cat.at_soal 
		WHERE (softdelete = '0' OR softdelete IS NULL)
		  AND (kodesoal ILIKE $1 OR namasoal ILIKE $1 OR COALESCE(keterangan, '') ILIKE $1)
	`
	err := r.db.GetContext(ctx, &total, countQuery, searchPattern)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT s.kodesoal, s.namasoal, COALESCE(s.keterangan, '') as keterangan, 
		       COALESCE(s.t_inserttime, NOW()) as t_inserttime,
		       (SELECT COUNT(*) FROM cat.at_pertanyaan p WHERE p.kodesoal = s.kodesoal AND (p.softdelete = '0' OR p.softdelete IS NULL)) as total_questions_static,
		       (SELECT COUNT(*) FROM cat.cat_bank_stimulus bs WHERE bs.kodesoal = s.kodesoal AND (bs.softdelete = '0' OR bs.softdelete IS NULL)) as total_stimuli,
		       (
		         (SELECT COUNT(*) FROM cat.at_pertanyaan p WHERE p.kodesoal = s.kodesoal AND (p.softdelete = '0' OR p.softdelete IS NULL))
		         +
		         (SELECT COUNT(*) FROM cat.cat_stimulus_items si JOIN cat.cat_bank_stimulus bs ON bs.id_stimulus = si.id_stimulus WHERE bs.kodesoal = s.kodesoal AND (bs.softdelete = '0' OR bs.softdelete IS NULL) AND (si.softdelete = '0' OR si.softdelete IS NULL))
		       ) as total_questions
		FROM cat.at_soal s
		WHERE (s.softdelete = '0' OR s.softdelete IS NULL)
		  AND (s.kodesoal ILIKE $1 OR s.namasoal ILIKE $1 OR COALESCE(s.keterangan, '') ILIKE $1)
		ORDER BY s.kodesoal DESC LIMIT $2 OFFSET $3
	`
	var banks []*entity.SoalHeader
	err = r.db.SelectContext(ctx, &banks, query, searchPattern, perPage, offset)
	return banks, total, err
}

func (r *bankSoalRepository) GetQuestionBankByCode(ctx context.Context, code string) (*entity.SoalHeader, error) {
	query := `
		SELECT s.kodesoal, s.namasoal, COALESCE(s.keterangan, '') as keterangan, 
		       COALESCE(s.t_inserttime, NOW()) as t_inserttime,
		       (SELECT COUNT(*) FROM cat.at_pertanyaan p WHERE p.kodesoal = s.kodesoal AND (p.softdelete = '0' OR p.softdelete IS NULL)) as total_questions_static,
		       (SELECT COUNT(*) FROM cat.cat_bank_stimulus bs WHERE bs.kodesoal = s.kodesoal AND (bs.softdelete = '0' OR bs.softdelete IS NULL)) as total_stimuli,
		       (
		         (SELECT COUNT(*) FROM cat.at_pertanyaan p WHERE p.kodesoal = s.kodesoal AND (p.softdelete = '0' OR p.softdelete IS NULL))
		         +
		         (SELECT COUNT(*) FROM cat.cat_stimulus_items si JOIN cat.cat_bank_stimulus bs ON bs.id_stimulus = si.id_stimulus WHERE bs.kodesoal = s.kodesoal AND (bs.softdelete = '0' OR bs.softdelete IS NULL) AND (si.softdelete = '0' OR si.softdelete IS NULL))
		       ) as total_questions
		FROM cat.at_soal s
		WHERE s.kodesoal = $1 AND (s.softdelete = '0' OR s.softdelete IS NULL)
	`
	var bank entity.SoalHeader
	err := r.db.GetContext(ctx, &bank, query, code)
	return &bank, err
}

func (r *bankSoalRepository) GetQuestionsByCode(ctx context.Context, code string) ([]*entity.Pertanyaan, error) {
	query := `
		SELECT kodesoal, nourut, jenisjawaban, pertanyaan,
		       pertanyaanimage, pertanyaanaudio, pertanyaanvideo,
		       jawaban1, jawaban1image, jawaban1audio, jawaban1video,
		       jawaban2, jawaban2image, jawaban2audio, jawaban2video,
		       jawaban3, jawaban3image, jawaban3audio, jawaban3video,
		       jawaban4, jawaban4image, jawaban4audio, jawaban4video,
		       jawaban5, jawaban5image, jawaban5audio, jawaban5video,
		       COALESCE(jawabanbenar, 0) as jawabanbenar,
		       COALESCE(bobot, 1.0) as bobot,
		       COALESCE(bobot_benar, COALESCE(bobot, 1.0)) as bobot_benar,
		       COALESCE(bobot_salah, 0.0) as bobot_salah,
		       katakunci
		FROM cat.at_pertanyaan
		WHERE kodesoal = $1 AND (softdelete = '0' OR softdelete IS NULL)
		ORDER BY nourut ASC
	`
	var list []*entity.Pertanyaan
	err := r.db.SelectContext(ctx, &list, query, code)
	return list, err
}

func (r *bankSoalRepository) GetQuestionByNoUrut(ctx context.Context, code string, noUrut int) (*entity.Pertanyaan, error) {
	query := `
		SELECT kodesoal, nourut, jenisjawaban, pertanyaan,
		       pertanyaanimage, pertanyaanaudio, pertanyaanvideo,
		       jawaban1, jawaban1image, jawaban1audio, jawaban1video,
		       jawaban2, jawaban2image, jawaban2audio, jawaban2video,
		       jawaban3, jawaban3image, jawaban3audio, jawaban3video,
		       jawaban4, jawaban4image, jawaban4audio, jawaban4video,
		       jawaban5, jawaban5image, jawaban5audio, jawaban5video,
		       COALESCE(jawabanbenar, 0) as jawabanbenar,
		       COALESCE(bobot, 1.0) as bobot,
		       COALESCE(bobot_benar, COALESCE(bobot, 1.0)) as bobot_benar,
		       COALESCE(bobot_salah, 0.0) as bobot_salah,
		       katakunci
		FROM cat.at_pertanyaan
		WHERE kodesoal = $1 AND nourut = $2 AND (softdelete = '0' OR softdelete IS NULL)
	`
	var item entity.Pertanyaan
	err := r.db.GetContext(ctx, &item, query, code, noUrut)
	return &item, err
}

func (r *bankSoalRepository) CreateQuestionBank(ctx context.Context, code, name, desc string) error {
	query := `
		INSERT INTO cat.at_soal (kodesoal, namasoal, keterangan, softdelete, t_inserttime)
		VALUES ($1, $2, $3, '0', NOW())
		ON CONFLICT (kodesoal) DO UPDATE
		SET namasoal = EXCLUDED.namasoal, keterangan = EXCLUDED.keterangan, softdelete = '0'
	`
	_, err := r.db.ExecContext(ctx, query, code, name, desc)
	return err
}

func (r *bankSoalRepository) UpdateQuestionBank(ctx context.Context, code, name, desc string) error {
	query := `
		UPDATE cat.at_soal
		SET namasoal = $2, keterangan = $3, t_updatetime = NOW()
		WHERE kodesoal = $1
	`
	res, err := r.db.ExecContext(ctx, query, code, name, desc)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("bank soal dengan kode %s tidak ditemukan", code)
	}
	return nil
}

func (r *bankSoalRepository) DeleteQuestionBank(ctx context.Context, code string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Soft delete question bank
	_, err = tx.ExecContext(ctx, `UPDATE cat.at_soal SET softdelete = '1', t_updatetime = NOW() WHERE kodesoal = $1`, code)
	if err != nil {
		return err
	}

	// Soft delete questions inside bank
	_, err = tx.ExecContext(ctx, `UPDATE cat.at_pertanyaan SET softdelete = '1', t_updatetime = NOW() WHERE kodesoal = $1`, code)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *bankSoalRepository) CopyQuestionBank(ctx context.Context, sourceCode, newCode, newName string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Create header
	insHeader := `
		INSERT INTO cat.at_soal (kodesoal, namasoal, keterangan, softdelete, t_inserttime)
		VALUES ($1, $2, 'Salinan dari ' || $3, '0', NOW())
	`
	_, err = tx.ExecContext(ctx, insHeader, newCode, newName, sourceCode)
	if err != nil {
		return fmt.Errorf("gagal membuat header bank soal baru: %w", err)
	}

	// 2. Duplicate questions
	insQuestions := `
		INSERT INTO cat.at_pertanyaan (
			kodesoal, nourut, jenisjawaban, pertanyaan, 
			pertanyaanimage, pertanyaanaudio, pertanyaanvideo,
			jawaban1, jawaban1image, jawaban1audio, jawaban1video,
			jawaban2, jawaban2image, jawaban2audio, jawaban2video,
			jawaban3, jawaban3image, jawaban3audio, jawaban3video,
			jawaban4, jawaban4image, jawaban4audio, jawaban4video,
			jawaban5, jawaban5image, jawaban5audio, jawaban5video,
			jawabanbenar, bobot, bobot_benar, bobot_salah, katakunci, isaktif, ispakai, softdelete, t_inserttime
		)
		SELECT $1, nourut, jenisjawaban, pertanyaan,
		       pertanyaanimage, pertanyaanaudio, pertanyaanvideo,
		       jawaban1, jawaban1image, jawaban1audio, jawaban1video,
		       jawaban2, jawaban2image, jawaban2audio, jawaban2video,
		       jawaban3, jawaban3image, jawaban3audio, jawaban3video,
		       jawaban4, jawaban4image, jawaban4audio, jawaban4video,
		       jawaban5, jawaban5image, jawaban5audio, jawaban5video,
		       jawabanbenar, COALESCE(bobot, 1.0), COALESCE(bobot_benar, COALESCE(bobot, 1.0)), COALESCE(bobot_salah, 0.0), katakunci, 1, 0, '0', NOW()
		FROM cat.at_pertanyaan
		WHERE kodesoal = $2 AND (softdelete = '0' OR softdelete IS NULL)
		ORDER BY nourut ASC
	`
	_, err = tx.ExecContext(ctx, insQuestions, newCode, sourceCode)
	if err != nil {
		return fmt.Errorf("gagal menyalin butir pertanyaan: %w", err)
	}

	return tx.Commit()
}

func getMediaAssignment(mediaType string, mediaURL *string) (*string, *string, *string) {
	var img, aud, vid *string
	if mediaURL == nil || *mediaURL == "" {
		return nil, nil, nil
	}
	switch mediaType {
	case "IMAGE":
		img = mediaURL
	case "AUDIO":
		aud = mediaURL
	case "VIDEO":
		vid = mediaURL
	}
	return img, aud, vid
}

func (r *bankSoalRepository) CreateQuestionItem(ctx context.Context, code string, req *dto.SaveQuestionItemDTO) (*entity.Pertanyaan, error) {
	var maxNo int
	_ = r.db.GetContext(ctx, &maxNo, `SELECT COALESCE(MAX(nourut), 0) FROM cat.at_pertanyaan WHERE kodesoal = $1`, code)
	newNo := maxNo + 1

	catTag := "Umum"
	if req.CategoryTag != "" {
		catTag = req.CategoryTag
	}

	bobotBenar := req.BobotBenar
	if bobotBenar == 0 && req.Bobot > 0 {
		bobotBenar = req.Bobot
	}
	if bobotBenar == 0 {
		bobotBenar = 1.0
	}
	bobotSalah := req.BobotSalah

	qImg, qAud, qVid := getMediaAssignment(req.QuestionMediaType, req.QuestionMediaURL)
	optAImg, optAAud, optAVid := getMediaAssignment(req.OptionAMediaType, req.OptionAMediaURL)
	optBImg, optBAud, optBVid := getMediaAssignment(req.OptionBMediaType, req.OptionBMediaURL)
	optCImg, optCAud, optCVid := getMediaAssignment(req.OptionCMediaType, req.OptionCMediaURL)
	optDImg, optDAud, optDVid := getMediaAssignment(req.OptionDMediaType, req.OptionDMediaURL)
	optEImg, optEAud, optEVid := getMediaAssignment(req.OptionEMediaType, req.OptionEMediaURL)

	query := `
		INSERT INTO cat.at_pertanyaan (
			kodesoal, nourut, jenisjawaban, pertanyaan,
			pertanyaanimage, pertanyaanaudio, pertanyaanvideo,
			jawaban1, jawaban1image, jawaban1audio, jawaban1video,
			jawaban2, jawaban2image, jawaban2audio, jawaban2video,
			jawaban3, jawaban3image, jawaban3audio, jawaban3video,
			jawaban4, jawaban4image, jawaban4audio, jawaban4video,
			jawaban5, jawaban5image, jawaban5audio, jawaban5video,
			jawabanbenar, bobot, bobot_benar, bobot_salah, katakunci, isaktif, ispakai, softdelete, t_inserttime
		) VALUES (
			$1, $2, 'M', $3,
			$4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18,
			$19, $20, $21, $22,
			$23, $24, $25, $26,
			$27, $28, $29, $30, $31, 1, 0, '0', NOW()
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		code, newNo, req.QuestionText,
		qImg, qAud, qVid,
		req.OptionA, optAImg, optAAud, optAVid,
		req.OptionB, optBImg, optBAud, optBVid,
		req.OptionC, optCImg, optCAud, optCVid,
		req.OptionD, optDImg, optDAud, optDVid,
		req.OptionE, optEImg, optEAud, optEVid,
		req.CorrectOptionKey, bobotBenar, bobotBenar, bobotSalah, catTag,
	)
	if err != nil {
		return nil, err
	}

	return &entity.Pertanyaan{
		KodeSoal:        code,
		NoUrut:          newNo,
		JenisJawaban:    "M",
		Pertanyaan:      req.QuestionText,
		PertanyaanImage: qImg,
		PertanyaanAudio: qAud,
		PertanyaanVideo: qVid,
		Jawaban1:        &req.OptionA,
		Jawaban1Image:   optAImg,
		Jawaban1Audio:   optAAud,
		Jawaban1Video:   optAVid,
		Jawaban2:        &req.OptionB,
		Jawaban2Image:   optBImg,
		Jawaban2Audio:   optBAud,
		Jawaban2Video:   optBVid,
		Jawaban3:        &req.OptionC,
		Jawaban3Image:   optCImg,
		Jawaban3Audio:   optCAud,
		Jawaban3Video:   optCVid,
		Jawaban4:        &req.OptionD,
		Jawaban4Image:   optDImg,
		Jawaban4Audio:   optDAud,
		Jawaban4Video:   optDVid,
		Jawaban5:        &req.OptionE,
		Jawaban5Image:   optEImg,
		Jawaban5Audio:   optEAud,
		Jawaban5Video:   optEVid,
		JawabanBenar:    req.CorrectOptionKey,
		Bobot:           bobotBenar,
		BobotBenar:      bobotBenar,
		BobotSalah:      bobotSalah,
		KataKunci:       &catTag,
	}, nil
}

func (r *bankSoalRepository) UpdateQuestionItem(ctx context.Context, code string, noUrut int, req *dto.SaveQuestionItemDTO) error {
	catTag := "Umum"
	if req.CategoryTag != "" {
		catTag = req.CategoryTag
	}

	bobotBenar := req.BobotBenar
	if bobotBenar == 0 && req.Bobot > 0 {
		bobotBenar = req.Bobot
	}
	if bobotBenar == 0 {
		bobotBenar = 1.0
	}
	bobotSalah := req.BobotSalah

	qImg, qAud, qVid := getMediaAssignment(req.QuestionMediaType, req.QuestionMediaURL)
	optAImg, optAAud, optAVid := getMediaAssignment(req.OptionAMediaType, req.OptionAMediaURL)
	optBImg, optBAud, optBVid := getMediaAssignment(req.OptionBMediaType, req.OptionBMediaURL)
	optCImg, optCAud, optCVid := getMediaAssignment(req.OptionCMediaType, req.OptionCMediaURL)
	optDImg, optDAud, optDVid := getMediaAssignment(req.OptionDMediaType, req.OptionDMediaURL)
	optEImg, optEAud, optEVid := getMediaAssignment(req.OptionEMediaType, req.OptionEMediaURL)

	query := `
		UPDATE cat.at_pertanyaan
		SET pertanyaan = $3,
		    pertanyaanimage = $4,
		    pertanyaanaudio = $5,
		    pertanyaanvideo = $6,
		    jawaban1 = $7,
		    jawaban1image = $8,
		    jawaban1audio = $9,
		    jawaban1video = $10,
		    jawaban2 = $11,
		    jawaban2image = $12,
		    jawaban2audio = $13,
		    jawaban2video = $14,
		    jawaban3 = $15,
		    jawaban3image = $16,
		    jawaban3audio = $17,
		    jawaban3video = $18,
		    jawaban4 = $19,
		    jawaban4image = $20,
		    jawaban4audio = $21,
		    jawaban4video = $22,
		    jawaban5 = $23,
		    jawaban5image = $24,
		    jawaban5audio = $25,
		    jawaban5video = $26,
		    jawabanbenar = $27,
		    bobot = $28,
		    bobot_benar = $29,
		    bobot_salah = $30,
		    katakunci = $31,
		    t_updatetime = NOW()
		WHERE kodesoal = $1 AND nourut = $2 AND (softdelete = '0' OR softdelete IS NULL)
	`
	res, err := r.db.ExecContext(ctx, query,
		code, noUrut, req.QuestionText,
		qImg, qAud, qVid,
		req.OptionA, optAImg, optAAud, optAVid,
		req.OptionB, optBImg, optBAud, optBVid,
		req.OptionC, optCImg, optCAud, optCVid,
		req.OptionD, optDImg, optDAud, optDVid,
		req.OptionE, optEImg, optEAud, optEVid,
		req.CorrectOptionKey, bobotBenar, bobotBenar, bobotSalah, catTag,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("butir soal #%d pada bank %s tidak ditemukan", noUrut, code)
	}
	return nil
}

func (r *bankSoalRepository) DeleteQuestionItem(ctx context.Context, code string, noUrut int) error {
	query := `
		UPDATE cat.at_pertanyaan
		SET softdelete = '1', t_updatetime = NOW()
		WHERE kodesoal = $1 AND nourut = $2
	`
	res, err := r.db.ExecContext(ctx, query, code, noUrut)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("butir soal #%d pada bank %s tidak ditemukan", noUrut, code)
	}
	return nil
}
