package repository

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/peserta/dto"
	"poltekkes-cat-backend/internal/modules/peserta/entity"

	"github.com/jmoiron/sqlx"
)

type PesertaRepository interface {
	GetParticipants(ctx context.Context, periodID, page, perPage int, search string) ([]*entity.Peserta, int64, error)
	GetParticipantByCode(ctx context.Context, code string) (*entity.Peserta, error)
	CreateParticipant(ctx context.Context, p *entity.Peserta) error
	UpdateParticipant(ctx context.Context, code string, p *entity.Peserta) error
	DeleteParticipant(ctx context.Context, code string) error
	ResetLogin(ctx context.Context, code string) error
	SearchParticipants(ctx context.Context, query string) ([]*entity.Peserta, error)
	ChangePassword(ctx context.Context, code string, hashedPassword string, hint string) error

	// Sipenmaru Integration
	GetSipenmaruFilterOptions(ctx context.Context) (*dto.SipenmaruFilterOptionsDTO, error)
	GetSipenmaruCandidates(ctx context.Context, req *dto.SipenmaruPreviewRequestDTO) ([]dto.SipenmaruCandidateDTO, error)
	ImportSipenmaruCandidates(ctx context.Context, req *dto.ImportSipenmaruRequestDTO) (*dto.ImportSipenmaruResponseDTO, error)
}

type pesertaRepository struct {
	db       *sqlx.DB // cbtDB
	siakadDB *sqlx.DB
}

func NewPesertaRepository(db *sqlx.DB, siakadDBs ...*sqlx.DB) PesertaRepository {
	sDB := db
	if len(siakadDBs) > 0 && siakadDBs[0] != nil {
		sDB = siakadDBs[0]
	}
	return &pesertaRepository{db: db, siakadDB: sDB}
}

func (r *pesertaRepository) GetParticipants(ctx context.Context, periodID, page, perPage int, search string) ([]*entity.Peserta, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	whereClause := " WHERE (p.softdelete = '0' OR p.softdelete IS NULL)"
	args := []interface{}{}
	argIdx := 1

	if periodID > 0 {
		whereClause += fmt.Sprintf(" AND p.idperiode = $%d", argIdx)
		args = append(args, periodID)
		argIdx++
	}

	if search != "" {
		whereClause += fmt.Sprintf(" AND (p.nama ILIKE $%d OR p.kodepeserta ILIKE $%d OR p.email ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM cat.at_peserta p" + whereClause
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT p.kodepeserta, p.idperiode, COALESCE(pr.namaperiode, '') AS namaperiode,
		       p.nama, p.jk,
		       p.hp, p.email, p.extimage, p.sumberdata, p.alamat, p.idkota,
		       COALESCE(w.namawilayah, p.idkota_lama, '') AS namawilayah,
		       p.profesi, p.status, p.filepembayaran, p.hint,
		       p.isvalid, p.isaktif, p.islogin, p.keterangan, p.kodereferensi, p.softdelete
		FROM cat.at_peserta p
		LEFT JOIN cat.at_periode pr ON pr.idperiode = p.idperiode
		LEFT JOIN cat.ms_wilayah w ON (p.idkota IS NOT NULL AND w.idwilayah = CAST(p.idkota AS TEXT))
		                           OR (p.idkota_lama IS NOT NULL AND p.idkota_lama != '' AND w.idwilayah = p.idkota_lama)
		%s
		ORDER BY p.kodepeserta ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	queryArgs := append(args, perPage, offset)
	var participants []*entity.Peserta
	err = r.db.SelectContext(ctx, &participants, query, queryArgs...)
	return participants, total, err
}

func (r *pesertaRepository) GetParticipantByCode(ctx context.Context, code string) (*entity.Peserta, error) {
	query := `
		SELECT p.kodepeserta, p.idperiode, COALESCE(pr.namaperiode, '') AS namaperiode,
		       p.nama, p.jk,
		       p.hp, p.email, p.extimage, p.sumberdata, p.alamat, p.idkota,
		       COALESCE(w.namawilayah, p.idkota_lama, '') AS namawilayah,
		       p.profesi, p.status, p.filepembayaran, p.hint,
		       p.isvalid, p.isaktif, p.islogin, p.keterangan, p.kodereferensi, p.softdelete
		FROM cat.at_peserta p
		LEFT JOIN cat.at_periode pr ON pr.idperiode = p.idperiode
		LEFT JOIN cat.ms_wilayah w ON (p.idkota IS NOT NULL AND w.idwilayah = CAST(p.idkota AS TEXT))
		                           OR (p.idkota_lama IS NOT NULL AND p.idkota_lama != '' AND w.idwilayah = p.idkota_lama)
		WHERE p.kodepeserta = $1 AND (p.softdelete = '0' OR p.softdelete IS NULL)
	`
	var p entity.Peserta
	err := r.db.GetContext(ctx, &p, query, code)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *pesertaRepository) CreateParticipant(ctx context.Context, p *entity.Peserta) error {
	query := `
		INSERT INTO cat.at_peserta (
			kodepeserta, idperiode, nama, jk, hp, email,
			sumberdata, alamat, idkota, idkota_lama, profesi, status,
			filepembayaran, hint, password, isvalid, isaktif,
			islogin, keterangan, kodereferensi, softdelete
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			0, $18, $19, '0'
		)
		ON CONFLICT (kodepeserta) DO UPDATE SET
			idperiode = EXCLUDED.idperiode,
			nama = EXCLUDED.nama,
			jk = EXCLUDED.jk,
			hp = EXCLUDED.hp,
			email = EXCLUDED.email,
			sumberdata = EXCLUDED.sumberdata,
			alamat = EXCLUDED.alamat,
			idkota = EXCLUDED.idkota,
			idkota_lama = EXCLUDED.idkota_lama,
			profesi = EXCLUDED.profesi,
			status = EXCLUDED.status,
			filepembayaran = COALESCE(EXCLUDED.filepembayaran, cat.at_peserta.filepembayaran),
			hint = EXCLUDED.hint,
			password = COALESCE(EXCLUDED.password, cat.at_peserta.password),
			isvalid = EXCLUDED.isvalid,
			isaktif = EXCLUDED.isaktif,
			keterangan = EXCLUDED.keterangan,
			kodereferensi = EXCLUDED.kodereferensi,
			softdelete = '0'
	`
	_, err := r.db.ExecContext(ctx, query,
		p.KodePeserta, p.IDPeriode, p.Nama, p.JK, p.HP, p.Email,
		p.SumberData, p.Alamat, p.IDKota, p.IDKotaLama, p.Profesi, p.Status,
		p.FilePembayaran, p.Hint, p.Password, p.IsValid, p.IsAktif,
		p.Keterangan, p.KodeReferensi,
	)
	return err
}

func (r *pesertaRepository) UpdateParticipant(ctx context.Context, code string, p *entity.Peserta) error {
	query := `
		UPDATE cat.at_peserta SET
			idperiode = $1,
			nama = $2,
			jk = $3,
			hp = $4,
			email = $5,
			sumberdata = $6,
			alamat = $7,
			idkota = $8,
			idkota_lama = $9,
			profesi = $10,
			status = $11,
			filepembayaran = COALESCE($12, filepembayaran),
			hint = COALESCE($13, hint),
			password = COALESCE($14, password),
			isvalid = $15,
			isaktif = $16,
			keterangan = $17,
			kodereferensi = $18,
			softdelete = '0'
		WHERE kodepeserta = $19
	`
	_, err := r.db.ExecContext(ctx, query,
		p.IDPeriode, p.Nama, p.JK, p.HP, p.Email,
		p.SumberData, p.Alamat, p.IDKota, p.IDKotaLama, p.Profesi, p.Status,
		p.FilePembayaran, p.Hint, p.Password, p.IsValid, p.IsAktif,
		p.Keterangan, p.KodeReferensi, code,
	)
	return err
}

func (r *pesertaRepository) DeleteParticipant(ctx context.Context, code string) error {
	query := `UPDATE cat.at_peserta SET softdelete = '1' WHERE kodepeserta = $1`
	_, err := r.db.ExecContext(ctx, query, code)
	return err
}

func (r *pesertaRepository) ResetLogin(ctx context.Context, code string) error {
	query := `UPDATE cat.at_peserta SET islogin = 0, tokenlogin = NULL WHERE kodepeserta = $1`
	_, err := r.db.ExecContext(ctx, query, code)
	return err
}

func (r *pesertaRepository) ChangePassword(ctx context.Context, code string, hashedPassword string, hint string) error {
	query := `UPDATE cat.at_peserta SET password = $1, hint = $2 WHERE kodepeserta = $3 AND (softdelete = '0' OR softdelete IS NULL)`
	result, err := r.db.ExecContext(ctx, query, hashedPassword, hint, code)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("peserta dengan kode %s tidak ditemukan", code)
	}
	return nil
}

func (r *pesertaRepository) SearchParticipants(ctx context.Context, query string) ([]*entity.Peserta, error) {
	sql := `
		SELECT p.kodepeserta, p.idperiode, COALESCE(pr.namaperiode, '') AS namaperiode,
		       p.nama, p.jk, p.hp, p.email, p.alamat, p.isvalid, p.isaktif
		FROM cat.at_peserta p
		LEFT JOIN cat.at_periode pr ON pr.idperiode = p.idperiode
		WHERE (p.softdelete = '0' OR p.softdelete IS NULL)
		  AND (p.nama ILIKE $1 OR p.kodepeserta ILIKE $1)
		ORDER BY p.nama ASC
		LIMIT 15
	`
	var list []*entity.Peserta
	err := r.db.SelectContext(ctx, &list, sql, "%"+query+"%")
	return list, err
}

func (r *pesertaRepository) GetSipenmaruFilterOptions(ctx context.Context) (*dto.SipenmaruFilterOptionsDTO, error) {
	opts := &dto.SipenmaruFilterOptionsDTO{
		CBTPeriods:    make([]dto.CBTPeriodOptionDTO, 0),
		PMBPeriods:    make([]string, 0),
		SistemKuliah:  make([]dto.SistemKuliahOptionDTO, 0),
		JalurDaftar:   make([]dto.JalurOptionDTO, 0),
		GelombangList: make([]dto.GelombangOptionDTO, 0),
	}

	// 1. CBT Periods
	cbtQuery := `
		SELECT idperiode, COALESCE(namaperiode, '') AS namaperiode 
		FROM cat.at_periode 
		WHERE softdelete = '0' OR softdelete IS NULL 
		ORDER BY idperiode DESC
	`
	rows, err := r.db.QueryContext(ctx, cbtQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p dto.CBTPeriodOptionDTO
			if err := rows.Scan(&p.IDPeriode, &p.NamaPeriode); err == nil {
				opts.CBTPeriods = append(opts.CBTPeriods, p)
			}
		}
	}

	// 2. PMB Periods from pendaftaran.pd_periodedaftar (or distinct from pd_pendaftar)
	pmbQuery := `
		SELECT DISTINCT idperiode::text 
		FROM (
			SELECT idperiode FROM pendaftaran.pd_periodedaftar WHERE softdelete = '0' OR softdelete IS NULL
			UNION
			SELECT idperiode FROM pendaftaran.pd_pendaftar WHERE softdelete = '0' OR softdelete IS NULL
		) sub
		WHERE idperiode IS NOT NULL
		ORDER BY idperiode::text DESC
	`
	pmbRows, err := r.siakadDB.QueryContext(ctx, pmbQuery)
	if err == nil {
		defer pmbRows.Close()
		for pmbRows.Next() {
			var per string
			if err := pmbRows.Scan(&per); err == nil && per != "" {
				opts.PMBPeriods = append(opts.PMBPeriods, per)
			}
		}
	}

	// 3. Sistem Kuliah
	skQuery := `
		SELECT idsistemkuliah, COALESCE(namasistemkuliah, '') 
		FROM ref.lv_sistemkuliah 
		WHERE softdelete = '0' OR softdelete IS NULL 
		ORDER BY idsistemkuliah ASC
	`
	skRows, err := r.siakadDB.QueryContext(ctx, skQuery)
	if err == nil {
		defer skRows.Close()
		for skRows.Next() {
			var sk dto.SistemKuliahOptionDTO
			if err := skRows.Scan(&sk.IDSistemKuliah, &sk.NamaSistemKuliah); err == nil {
				opts.SistemKuliah = append(opts.SistemKuliah, sk)
			}
		}
	}

	// 4. Jalur Pendaftaran
	jpQuery := `
		SELECT idjalurpendaftaran, COALESCE(namajalurpendaftaran, '') 
		FROM ref.lv_jalurpendaftaran 
		WHERE softdelete = '0' OR softdelete IS NULL 
		ORDER BY idjalurpendaftaran ASC
	`
	jpRows, err := r.siakadDB.QueryContext(ctx, jpQuery)
	if err == nil {
		defer jpRows.Close()
		for jpRows.Next() {
			var jp dto.JalurOptionDTO
			if err := jpRows.Scan(&jp.IDJalurPendaftaran, &jp.NamaJalurPendaftaran); err == nil {
				opts.JalurDaftar = append(opts.JalurDaftar, jp)
			}
		}
	}

	// 5. Gelombang
	glQuery := `
		SELECT idgelombang, COALESCE(namagelombang, '') 
		FROM ref.lv_gelombang 
		WHERE softdelete = '0' OR softdelete IS NULL 
		ORDER BY idgelombang ASC
	`
	glRows, err := r.siakadDB.QueryContext(ctx, glQuery)
	if err == nil {
		defer glRows.Close()
		for glRows.Next() {
			var gl dto.GelombangOptionDTO
			if err := glRows.Scan(&gl.IDGelombang, &gl.NamaGelombang); err == nil {
				opts.GelombangList = append(opts.GelombangList, gl)
			}
		}
	}

	return opts, nil
}

func (r *pesertaRepository) GetSipenmaruCandidates(ctx context.Context, req *dto.SipenmaruPreviewRequestDTO) ([]dto.SipenmaruCandidateDTO, error) {
	candidates := make([]dto.SipenmaruCandidateDTO, 0)

	baseQuery := `
		SELECT 
			p.idpendaftar,
			COALESCE(p.namapendaftar, '') AS nama,
			COALESCE(p.jk, 'L') AS jk,
			COALESCE(p.hp, '') AS hp,
			COALESCE(p.email, '') AS email,
			COALESCE(p.alamat, '') AS alamat,
			p.idkota,
			COALESCE(p.idperiode::text, '') AS id_periode,
			COALESCE(p.idsistemkuliah, 0) AS id_sistem_kuliah,
			COALESCE(sk.namasistemkuliah, '') AS nama_sistem_kuliah,
			COALESCE(p.idjalurpendaftaran, 0) AS id_jalur_pendaftaran,
			COALESCE(jp.namajalurpendaftaran, '') AS nama_jalur_pendaftaran,
			COALESCE(p.idgelombang, 0) AS id_gelombang,
			COALESCE(gl.namagelombang, '') AS nama_gelombang,
			p.idunit1::text AS id_unit1,
			COALESCE(u1.namaunit, '') AS nama_unit1,
			p.idunit2::text AS id_unit2,
			COALESCE(u2.namaunit, '') AS nama_unit2,
			COALESCE(p.isadministrasi, 0) AS is_administrasi,
			COALESCE(p.noujian, '') AS no_ujian_sipenmaru
		FROM pendaftaran.pd_pendaftar p
		LEFT JOIN ref.lv_sistemkuliah sk ON sk.idsistemkuliah = p.idsistemkuliah
		LEFT JOIN ref.lv_jalurpendaftaran jp ON jp.idjalurpendaftaran = p.idjalurpendaftaran
		LEFT JOIN ref.lv_gelombang gl ON gl.idgelombang = p.idgelombang
		LEFT JOIN ref.ms_unit u1 ON u1.idunit::text = p.idunit1::text
		LEFT JOIN ref.ms_unit u2 ON u2.idunit::text = p.idunit2::text
		WHERE (p.softdelete = '0' OR p.softdelete IS NULL)
	`

	var conditions []string
	var args []interface{}
	argIdx := 1

	pmbPeriodStr := req.GetPMBPeriodString()
	if pmbPeriodStr != "" {
		conditions = append(conditions, fmt.Sprintf("p.idperiode::text = $%d", argIdx))
		args = append(args, pmbPeriodStr)
		argIdx++
	}
	if req.IDSistemKuliah > 0 {
		conditions = append(conditions, fmt.Sprintf("p.idsistemkuliah = $%d", argIdx))
		args = append(args, req.IDSistemKuliah)
		argIdx++
	}
	if req.IDJalurPendaftaran > 0 {
		conditions = append(conditions, fmt.Sprintf("p.idjalurpendaftaran = $%d", argIdx))
		args = append(args, req.IDJalurPendaftaran)
		argIdx++
	}
	if req.IDGelombang > 0 {
		conditions = append(conditions, fmt.Sprintf("p.idgelombang = $%d", argIdx))
		args = append(args, req.IDGelombang)
		argIdx++
	}
	if req.OnlyAdministrasi {
		conditions = append(conditions, "p.isadministrasi = 1")
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY p.idpendaftar ASC"

	rows, err := r.siakadDB.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return candidates, fmt.Errorf("gagal query pendaftar: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c dto.SipenmaruCandidateDTO
		err := rows.Scan(
			&c.IDPendaftar,
			&c.Nama,
			&c.JK,
			&c.HP,
			&c.Email,
			&c.Alamat,
			&c.IDKota,
			&c.IDPeriode,
			&c.IDSistemKuliah,
			&c.NamaSistemKuliah,
			&c.IDJalurPendaftaran,
			&c.NamaJalurPendaftaran,
			&c.IDGelombang,
			&c.NamaGelombang,
			&c.IDUnit1,
			&c.NamaUnit1,
			&c.IDUnit2,
			&c.NamaUnit2,
			&c.IsAdministrasi,
			&c.NoUjianSipenmaru,
		)
		if err != nil {
			return candidates, fmt.Errorf("gagal scan candidate: %w", err)
		}
		candidates = append(candidates, c)
	}

	// Match against CBT database (cat.at_peserta) to mark already imported candidates
	if len(candidates) > 0 {
		var importedList []struct {
			IDPendaftar string `db:"idpendaftar"`
			KodePeserta string `db:"kodepeserta"`
		}
		_ = r.db.SelectContext(ctx, &importedList, `
			SELECT COALESCE(idpendaftar, '') as idpendaftar, kodepeserta 
			FROM cat.at_peserta 
			WHERE idpendaftar IS NOT NULL AND idpendaftar != '' AND (softdelete = '0' OR softdelete IS NULL)
		`)
		importedMap := make(map[string]string, len(importedList))
		for _, item := range importedList {
			importedMap[item.IDPendaftar] = item.KodePeserta
		}

		for i := range candidates {
			if kode, exists := importedMap[candidates[i].IDPendaftar]; exists {
				candidates[i].IsImported = true
				candidates[i].ExistingKodePeserta = kode
			}
		}
	}

	return candidates, nil
}

func (r *pesertaRepository) ImportSipenmaruCandidates(ctx context.Context, req *dto.ImportSipenmaruRequestDTO) (*dto.ImportSipenmaruResponseDTO, error) {
	resp := &dto.ImportSipenmaruResponseDTO{
		Errors:               make([]string, 0),
		ImportedParticipants: make([]dto.ImportedParticipantItem, 0),
	}

	// 1. Fetch eligible candidates
	previewReq := &dto.SipenmaruPreviewRequestDTO{
		CBTPeriodID:        req.CBTPeriodID,
		PMBPeriodID:        req.PMBPeriodID,
		IDSistemKuliah:     req.IDSistemKuliah,
		IDJalurPendaftaran: req.IDJalurPendaftaran,
		IDGelombang:        req.IDGelombang,
		OnlyAdministrasi:   req.OnlyAdministrasi,
	}

	candidates, err := r.GetSipenmaruCandidates(ctx, previewReq)
	if err != nil {
		return nil, err
	}

	selectedMap := make(map[string]bool)
	hasSelection := len(req.SelectedPendaftarIDs) > 0
	for _, id := range req.SelectedPendaftarIDs {
		selectedMap[id] = true
	}

	// 2. Year prefix for CBT kode peserta
	yearPrefix := fmt.Sprintf("%02d", time.Now().Year()%100)
	pmbPeriodStr := req.GetPMBPeriodString()
	if pmbPeriodStr != "" && len(pmbPeriodStr) >= 4 {
		yearPrefix = pmbPeriodStr[2:4]
	}

	// 3. Process each candidate
	for _, c := range candidates {
		if hasSelection && !selectedMap[c.IDPendaftar] {
			continue
		}

		resp.TotalProcessed++

		// Check if already imported
		if c.IsImported {
			resp.SkippedCount++
			continue
		}

		// Double check in database inside loop to avoid concurrency race
		var existingKode string
		checkErr := r.db.GetContext(ctx, &existingKode,
			"SELECT kodepeserta FROM cat.at_peserta WHERE idpendaftar = $1 AND (softdelete = '0' OR softdelete IS NULL)",
			c.IDPendaftar)
		if checkErr == nil && existingKode != "" {
			resp.SkippedCount++
			continue
		}

		// Determine or generate Exam Number (NoUjian)
		noujian := c.NoUjianSipenmaru
		kodepeserta := ""

		if noujian != "" {
			// If already has no ujian, check length
			if len(noujian) >= 12 {
				kodepeserta = noujian
			} else {
				kodepeserta = yearPrefix + noujian
			}
		} else {
			// Generate formal Poltekkes exam number: [prodi1:2][prodi2:2][ruang:3][jalur:2][sequence:4]
			p1 := "00"
			if c.IDUnit1 != nil && len(*c.IDUnit1) > 0 {
				var kd string
				_ = r.siakadDB.GetContext(ctx, &kd, "SELECT COALESCE(kodedaftar, '') FROM ref.ms_unit WHERE idunit::text = $1", *c.IDUnit1)
				if kd != "" {
					if len(kd) >= 2 {
						p1 = kd[:2]
					} else {
						p1 = fmt.Sprintf("%02s", kd)
					}
				} else {
					u := *c.IDUnit1
					if len(u) >= 2 {
						p1 = u[len(u)-2:]
					}
				}
			}

			p2 := "00"
			if c.IDUnit2 != nil && len(*c.IDUnit2) > 0 {
				var kd string
				_ = r.siakadDB.GetContext(ctx, &kd, "SELECT COALESCE(kodedaftar, '') FROM ref.ms_unit WHERE idunit::text = $1", *c.IDUnit2)
				if kd != "" {
					if len(kd) >= 2 {
						p2 = kd[:2]
					} else {
						p2 = fmt.Sprintf("%02s", kd)
					}
				} else {
					u := *c.IDUnit2
					if len(u) >= 2 {
						p2 = u[len(u)-2:]
					}
				}
			}

			// Sequence calculation
			var maxSeq int
			seqQuery := `
				SELECT COALESCE(MAX(
					CASE WHEN length(kodepeserta) >= 4 THEN SUBSTRING(kodepeserta FROM length(kodepeserta)-3 FOR 4)::int ELSE 0 END
				), 0)
				FROM cat.at_peserta 
				WHERE idperiode = $1
			`
			_ = r.db.GetContext(ctx, &maxSeq, seqQuery, req.CBTPeriodID)
			seq := maxSeq + resp.ImportedCount + 1

			noujian = fmt.Sprintf("%s%s001%02d%04d", p1, p2, c.IDJalurPendaftaran, seq)
			kodepeserta = fmt.Sprintf("%s%s", yearPrefix, noujian)
		}

		if len(kodepeserta) > 20 {
			kodepeserta = kodepeserta[:20]
		}

		// Ensure kodepeserta is unique in cat.at_peserta
		var dupCount int
		_ = r.db.GetContext(ctx, &dupCount, "SELECT COUNT(*) FROM cat.at_peserta WHERE kodepeserta = $1", kodepeserta)
		if dupCount > 0 {
			kodepeserta = fmt.Sprintf("%s%d", kodepeserta[:len(kodepeserta)-3], resp.ImportedCount+1)
		}

		// Password: MD5 of kodepeserta
		hasher := md5.New()
		hasher.Write([]byte(kodepeserta))
		hashedPassword := hex.EncodeToString(hasher.Sum(nil))
		hint := kodepeserta

		// Insert into cat.at_peserta
		insertQuery := `
			INSERT INTO cat.at_peserta (
				kodepeserta, idperiode, nama, jk, hp, email, alamat, idkota,
				password, hint, sumberdata, isvalid, isaktif, islogin,
				kodereferensi, idpendaftar, softdelete
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8,
				$9, $10, 'P', 1, 0, 1,
				$11, $12, '0'
			)
			ON CONFLICT (kodepeserta) DO UPDATE SET
				nama = EXCLUDED.nama,
				idpendaftar = EXCLUDED.idpendaftar,
				softdelete = '0'
		`
		_, insertErr := r.db.ExecContext(ctx, insertQuery,
			kodepeserta, req.CBTPeriodID, c.Nama, c.JK, c.HP, c.Email, c.Alamat, c.IDKota,
			hashedPassword, hint, c.IDPendaftar, c.IDPendaftar,
		)
		if insertErr != nil {
			resp.ErrorCount++
			resp.Errors = append(resp.Errors, fmt.Sprintf("%s (%s): %v", c.Nama, c.IDPendaftar, insertErr))
			continue
		}

		// Update noujian in pendaftaran.pd_pendaftar if empty
		if c.NoUjianSipenmaru == "" {
			_, _ = r.siakadDB.ExecContext(ctx,
				"UPDATE pendaftaran.pd_pendaftar SET noujian = $1 WHERE idpendaftar = $2 AND (noujian IS NULL OR noujian = '')",
				noujian, c.IDPendaftar)
		}

		resp.ImportedCount++
		resp.ImportedParticipants = append(resp.ImportedParticipants, dto.ImportedParticipantItem{
			KodePeserta: kodepeserta,
			IDPendaftar: c.IDPendaftar,
			Nama:        c.Nama,
		})
	}

	return resp, nil
}
