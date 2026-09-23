package repository

import (
	"context"
	"fmt"
	"poltekkes-cat-backend/internal/modules/referensi/dto"
	"poltekkes-cat-backend/internal/modules/referensi/entity"

	"github.com/jmoiron/sqlx"
)

type ReferensiRepository interface {
	GetRuangList(ctx context.Context) ([]*entity.Ruang, error)
	CreateRuang(ctx context.Context, req *dto.CreateRuangRequestDTO) (string, error)
	UpdateRuang(ctx context.Context, kodeRuang string, req *dto.UpdateRuangRequestDTO) error
	DeleteRuang(ctx context.Context, kodeRuang string) error

	GetSkorList(ctx context.Context) ([]*entity.Skor, error)

	GetJenisUjianList(ctx context.Context) ([]*entity.JenisUjian, error)
	CreateJenisUjian(ctx context.Context, req *dto.CreateJenisUjianRequestDTO) (string, error)
	UpdateJenisUjian(ctx context.Context, kodeJenis string, req *dto.UpdateJenisUjianRequestDTO) error
	DeleteJenisUjian(ctx context.Context, kodeJenis string) error

	GetWilayahList(ctx context.Context, parentID string) ([]*entity.Wilayah, error)

	GetPanitiaList(ctx context.Context) ([]*entity.Panitia, error)
	CreatePanitia(ctx context.Context, req *dto.CreatePanitiaRequestDTO) error
	UpdatePanitia(ctx context.Context, nip string, req *dto.UpdatePanitiaRequestDTO) error
	DeletePanitia(ctx context.Context, nip string) error

	GetJenisPeriodeList(ctx context.Context) ([]*entity.JenisPeriode, error)
	CreateJenisPeriode(ctx context.Context, req *dto.CreateJenisPeriodeRequestDTO) error
	UpdateJenisPeriode(ctx context.Context, jenisPeriode string, req *dto.UpdateJenisPeriodeRequestDTO) error
	DeleteJenisPeriode(ctx context.Context, jenisPeriode string) error

	GetUnitList(ctx context.Context) ([]*entity.Unit, error)
	SalinSoal(ctx context.Context, sumberKode, targetKode, targetNama string) (int, error)
}

type referensiRepository struct {
	db       *sqlx.DB // cbtDB
	siakadDB *sqlx.DB
}

func NewReferensiRepository(db *sqlx.DB, siakadDBs ...*sqlx.DB) ReferensiRepository {
	sDB := db
	if len(siakadDBs) > 0 && siakadDBs[0] != nil {
		sDB = siakadDBs[0]
	}
	return &referensiRepository{db: db, siakadDB: sDB}
}

func (r *referensiRepository) GetRuangList(ctx context.Context) ([]*entity.Ruang, error) {
	query := `
		SELECT koderuang, namaruang, COALESCE(kapasitas, 40) as kapasitas, keterangan
		FROM cat.at_ruang
		WHERE softdelete = '0' OR softdelete IS NULL
		ORDER BY koderuang ASC
	`
	var list []*entity.Ruang
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *referensiRepository) GetSkorList(ctx context.Context) ([]*entity.Skor, error) {
	query := `
		SELECT kodeskor, skorbenar, skorsalah
		FROM cat.at_skor
		WHERE softdelete = '0' OR softdelete IS NULL
		ORDER BY kodeskor ASC
	`
	var list []*entity.Skor
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *referensiRepository) GetJenisUjianList(ctx context.Context) ([]*entity.JenisUjian, error) {
	query := `
		SELECT kodejenis, namajenis
		FROM cat.at_jenisujian
		WHERE softdelete = '0' OR softdelete IS NULL
		ORDER BY kodejenis ASC
	`
	var list []*entity.JenisUjian
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *referensiRepository) GetWilayahList(ctx context.Context, parentID string) ([]*entity.Wilayah, error) {
	query := `
		SELECT idwilayah, parentwilayah, namawilayah, level
		FROM cat.ms_wilayah
		WHERE (softdelete = '0' OR softdelete IS NULL)
	`
	if parentID != "" {
		query += fmt.Sprintf(" AND parentwilayah = '%s'", parentID)
	}
	query += " ORDER BY idwilayah ASC LIMIT 100"

	var list []*entity.Wilayah
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *referensiRepository) GetPanitiaList(ctx context.Context) ([]*entity.Panitia, error) {
	query := `
		SELECT nip, namapanitia
		FROM cat.at_panitia
		WHERE softdelete = '0' OR softdelete IS NULL
		ORDER BY nip ASC
	`
	var list []*entity.Panitia
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *referensiRepository) GetJenisPeriodeList(ctx context.Context) ([]*entity.JenisPeriode, error) {
	query := `
		SELECT jenisperiode, namajenisperiode
		FROM cat.at_jenisperiode
		WHERE softdelete = '0' OR softdelete IS NULL
		ORDER BY jenisperiode ASC
	`
	var list []*entity.JenisPeriode
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *referensiRepository) GetUnitList(ctx context.Context) ([]*entity.Unit, error) {
	query := `
		SELECT idsatker, namasatker, COALESCE(namasingkat, '') as namasingkat
		FROM gate.sc_unit
		WHERE softdelete = '0' OR softdelete IS NULL
		ORDER BY namasatker ASC
	`
	var list []*entity.Unit
	err := r.siakadDB.SelectContext(ctx, &list, query)
	if err != nil {
		return make([]*entity.Unit, 0), nil
	}
	return list, nil
}

func (r *referensiRepository) SalinSoal(ctx context.Context, sumberKode, targetKode, targetNama string) (int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. Insert or update header soal
	insertHeader := `
		INSERT INTO cat.at_soal (kodesoal, namasoal, keterangan, t_inserttime, softdelete)
		VALUES ($1, $2, 'Disalin dari ' || $3, NOW(), '0')
		ON CONFLICT (kodesoal) DO UPDATE 
		SET namasoal = EXCLUDED.namasoal, keterangan = EXCLUDED.keterangan, softdelete = '0'
	`
	_, err = tx.ExecContext(ctx, insertHeader, targetKode, targetNama, sumberKode)
	if err != nil {
		return 0, fmt.Errorf("gagal membuat header bank soal tujuan: %w", err)
	}

	// 2. Salin seluruh pertanyaan dan opsi jawaban yang belum ada di target
	copyQuery := `
		INSERT INTO cat.at_pertanyaan (
			kodesoal, nourut, jenisjawaban, pertanyaan, 
			jawaban1, jawaban2, jawaban3, jawaban4, jawaban5, 
			jawabanbenar, isaktif, softdelete
		)
		SELECT $1::varchar, p.nourut, p.jenisjawaban, p.pertanyaan, 
		       p.jawaban1, p.jawaban2, p.jawaban3, p.jawaban4, p.jawaban5, 
		       p.jawabanbenar, 1, '0'
		FROM cat.at_pertanyaan p
		WHERE p.kodesoal = $2 
		  AND (p.softdelete = '0' OR p.softdelete IS NULL)
		  AND NOT EXISTS (
		      SELECT 1 FROM cat.at_pertanyaan t 
		      WHERE t.kodesoal = $1 AND t.nourut = p.nourut
		  )
	`
	result, err := tx.ExecContext(ctx, copyQuery, targetKode, sumberKode)
	if err != nil {
		return 0, fmt.Errorf("gagal menyalin butir pertanyaan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

func (r *referensiRepository) CreateRuang(ctx context.Context, req *dto.CreateRuangRequestDTO) (string, error) {
	kode := req.KodeRuang
	if kode == "" {
		var nextNum int
		err := r.db.GetContext(ctx, &nextNum, `SELECT COALESCE(MAX(NULLIF(regexp_replace(koderuang, '\D', '', 'g'), '')::integer), 0) + 1 FROM cat.at_ruang`)
		if err != nil {
			nextNum = 1
		}
		kode = fmt.Sprintf("%03d", nextNum)
	}

	query := `
		INSERT INTO cat.at_ruang (koderuang, namaruang, kapasitas, keterangan, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, $3, $4, '0', NOW(), 'i-ms_ruang')
		ON CONFLICT (koderuang) DO UPDATE
		SET namaruang = EXCLUDED.namaruang, kapasitas = EXCLUDED.kapasitas, keterangan = EXCLUDED.keterangan, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-ms_ruang'
	`
	_, err := r.db.ExecContext(ctx, query, kode, req.NamaRuang, req.Kapasitas, req.Keterangan)
	return kode, err
}

func (r *referensiRepository) UpdateRuang(ctx context.Context, kodeRuang string, req *dto.UpdateRuangRequestDTO) error {
	query := `
		UPDATE cat.at_ruang
		SET namaruang = $1, kapasitas = $2, keterangan = $3, t_updatetime = NOW(), t_updateact = 'u-ms_ruang'
		WHERE koderuang = $4
	`
	_, err := r.db.ExecContext(ctx, query, req.NamaRuang, req.Kapasitas, req.Keterangan, kodeRuang)
	return err
}

func (r *referensiRepository) DeleteRuang(ctx context.Context, kodeRuang string) error {
	query := `
		UPDATE cat.at_ruang
		SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-ms_ruang'
		WHERE koderuang = $1
	`
	_, err := r.db.ExecContext(ctx, query, kodeRuang)
	return err
}

func (r *referensiRepository) CreateJenisUjian(ctx context.Context, req *dto.CreateJenisUjianRequestDTO) (string, error) {
	kode := req.KodeJenis
	if kode == "" {
		var nextNum int
		err := r.db.GetContext(ctx, &nextNum, `SELECT COALESCE(MAX(NULLIF(regexp_replace(kodejenis, '\D', '', 'g'), '')::integer), 0) + 1 FROM cat.at_jenisujian`)
		if err != nil {
			nextNum = 1
		}
		kode = fmt.Sprintf("%d", nextNum)
	}

	query := `
		INSERT INTO cat.at_jenisujian (kodejenis, namajenis, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, '0', NOW(), 'i-ms_jenisujian')
		ON CONFLICT (kodejenis) DO UPDATE
		SET namajenis = EXCLUDED.namajenis, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-ms_jenisujian'
	`
	_, err := r.db.ExecContext(ctx, query, kode, req.NamaJenis)
	return kode, err
}

func (r *referensiRepository) UpdateJenisUjian(ctx context.Context, kodeJenis string, req *dto.UpdateJenisUjianRequestDTO) error {
	query := `
		UPDATE cat.at_jenisujian
		SET namajenis = $1, t_updatetime = NOW(), t_updateact = 'u-ms_jenisujian'
		WHERE kodejenis = $2
	`
	_, err := r.db.ExecContext(ctx, query, req.NamaJenis, kodeJenis)
	return err
}

func (r *referensiRepository) DeleteJenisUjian(ctx context.Context, kodeJenis string) error {
	query := `
		UPDATE cat.at_jenisujian
		SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-ms_jenisujian'
		WHERE kodejenis = $1
	`
	_, err := r.db.ExecContext(ctx, query, kodeJenis)
	return err
}

func (r *referensiRepository) CreatePanitia(ctx context.Context, req *dto.CreatePanitiaRequestDTO) error {
	query := `
		INSERT INTO cat.at_panitia (nip, namapanitia, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, '0', NOW(), 'i-ms_panitia')
		ON CONFLICT (nip) DO UPDATE
		SET namapanitia = EXCLUDED.namapanitia, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-ms_panitia'
	`
	_, err := r.db.ExecContext(ctx, query, req.NIP, req.NamaPanitia)
	return err
}

func (r *referensiRepository) UpdatePanitia(ctx context.Context, nip string, req *dto.UpdatePanitiaRequestDTO) error {
	query := `
		UPDATE cat.at_panitia
		SET namapanitia = $1, t_updatetime = NOW(), t_updateact = 'u-ms_panitia'
		WHERE nip = $2
	`
	_, err := r.db.ExecContext(ctx, query, req.NamaPanitia, nip)
	return err
}

func (r *referensiRepository) DeletePanitia(ctx context.Context, nip string) error {
	query := `
		UPDATE cat.at_panitia
		SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-ms_panitia'
		WHERE nip = $1
	`
	_, err := r.db.ExecContext(ctx, query, nip)
	return err
}

func (r *referensiRepository) CreateJenisPeriode(ctx context.Context, req *dto.CreateJenisPeriodeRequestDTO) error {
	query := `
		INSERT INTO cat.at_jenisperiode (jenisperiode, namajenisperiode, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, '0', NOW(), 'i-ms_jenisperiode')
		ON CONFLICT (jenisperiode) DO UPDATE
		SET namajenisperiode = EXCLUDED.namajenisperiode, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-ms_jenisperiode'
	`
	_, err := r.db.ExecContext(ctx, query, req.JenisPeriode, req.NamaJenisPeriode)
	return err
}

func (r *referensiRepository) UpdateJenisPeriode(ctx context.Context, jenisPeriode string, req *dto.UpdateJenisPeriodeRequestDTO) error {
	query := `
		UPDATE cat.at_jenisperiode
		SET namajenisperiode = $1, t_updatetime = NOW(), t_updateact = 'u-ms_jenisperiode'
		WHERE jenisperiode = $2
	`
	_, err := r.db.ExecContext(ctx, query, req.NamaJenisPeriode, jenisPeriode)
	return err
}

func (r *referensiRepository) DeleteJenisPeriode(ctx context.Context, jenisPeriode string) error {
	query := `
		UPDATE cat.at_jenisperiode
		SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-ms_jenisperiode'
		WHERE jenisperiode = $1
	`
	_, err := r.db.ExecContext(ctx, query, jenisPeriode)
	return err
}

