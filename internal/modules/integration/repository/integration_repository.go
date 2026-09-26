package repository

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/integration/dto"
	"poltekkes-cat-backend/internal/modules/integration/entity"

	"github.com/jmoiron/sqlx"
)

type IntegrationRepository interface {
	// API Key CRUD
	GetAPIKeys(ctx context.Context) ([]*entity.APIKey, error)
	GetAPIKeyByID(ctx context.Context, id int) (*entity.APIKey, error)
	GetAPIKeyBySecret(ctx context.Context, secret string) (*entity.APIKey, error)
	CreateAPIKey(ctx context.Context, key *entity.APIKey) (int, error)
	UpdateAPIKey(ctx context.Context, key *entity.APIKey) error
	DeleteAPIKey(ctx context.Context, id int) error
	RegenerateAPIKey(ctx context.Context, id int, newSecret string) error
	UpdateLastUsed(ctx context.Context, id int) error

	// SPMB Integration
	GetSPMBActiveExams(ctx context.Context, refDate string) ([]*dto.SPMBActiveExamDTO, error)
	RegisterSPMBParticipant(ctx context.Context, req *dto.SPMBRegisterRequestDTO, plainPassword string) (*dto.SPMBRegisterResponseDTO, error)
	GetSPMBParticipant(ctx context.Context, idPendaftar string) (*dto.SPMBParticipantDetailDTO, error)
	UpdateSPMBParticipant(ctx context.Context, idPendaftar string, req *dto.SPMBUpdateParticipantDTO) (*dto.SPMBParticipantDetailDTO, error)
	DeleteSPMBParticipant(ctx context.Context, idPendaftar string) error
	GetUnplottedQueueSummary(ctx context.Context) (*dto.UnplottedQueueSummaryDTO, error)

	// API Access Logs
	CreateAccessLog(ctx context.Context, log *entity.APIAccessLog) error
	GetAccessLogs(ctx context.Context, apiKeyID int, page, perPage int, search string) ([]*entity.APIAccessLog, int64, error)
}

type integrationRepository struct {
	db *sqlx.DB
}

func NewIntegrationRepository(db *sqlx.DB) IntegrationRepository {
	return &integrationRepository{db: db}
}

func (r *integrationRepository) GetAPIKeys(ctx context.Context) ([]*entity.APIKey, error) {
	query := `
		SELECT id, name, api_key, ip_whitelist, is_active, description, last_used_at, created_at, updated_at
		FROM cat.at_api_keys
		WHERE (softdelete = '0' OR softdelete IS NULL)
		ORDER BY id DESC
	`
	var keys []*entity.APIKey
	err := r.db.SelectContext(ctx, &keys, query)
	if keys == nil {
		keys = []*entity.APIKey{}
	}
	return keys, err
}

func (r *integrationRepository) GetAPIKeyByID(ctx context.Context, id int) (*entity.APIKey, error) {
	query := `
		SELECT id, name, api_key, ip_whitelist, is_active, description, last_used_at, created_at, updated_at
		FROM cat.at_api_keys
		WHERE id = $1 AND (softdelete = '0' OR softdelete IS NULL)
	`
	var key entity.APIKey
	err := r.db.GetContext(ctx, &key, query, id)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *integrationRepository) GetAPIKeyBySecret(ctx context.Context, secret string) (*entity.APIKey, error) {
	query := `
		SELECT id, name, api_key, ip_whitelist, is_active, description, last_used_at, created_at, updated_at
		FROM cat.at_api_keys
		WHERE api_key = $1 AND is_active = TRUE AND (softdelete = '0' OR softdelete IS NULL)
	`
	var key entity.APIKey
	err := r.db.GetContext(ctx, &key, query, secret)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *integrationRepository) CreateAPIKey(ctx context.Context, key *entity.APIKey) (int, error) {
	query := `
		INSERT INTO cat.at_api_keys (name, api_key, ip_whitelist, is_active, description, created_at, updated_at, softdelete)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), '0')
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query, key.Name, key.APIKey, key.IPWhitelist, key.IsActive, key.Description).Scan(&id)
	return id, err
}

func (r *integrationRepository) UpdateAPIKey(ctx context.Context, key *entity.APIKey) error {
	query := `
		UPDATE cat.at_api_keys
		SET name = $1, ip_whitelist = $2, is_active = $3, description = $4, updated_at = NOW()
		WHERE id = $5 AND (softdelete = '0' OR softdelete IS NULL)
	`
	_, err := r.db.ExecContext(ctx, query, key.Name, key.IPWhitelist, key.IsActive, key.Description, key.ID)
	return err
}

func (r *integrationRepository) DeleteAPIKey(ctx context.Context, id int) error {
	query := `
		UPDATE cat.at_api_keys
		SET softdelete = '1', updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *integrationRepository) RegenerateAPIKey(ctx context.Context, id int, newSecret string) error {
	query := `
		UPDATE cat.at_api_keys
		SET api_key = $1, updated_at = NOW()
		WHERE id = $2 AND (softdelete = '0' OR softdelete IS NULL)
	`
	_, err := r.db.ExecContext(ctx, query, newSecret, id)
	return err
}

func (r *integrationRepository) UpdateLastUsed(ctx context.Context, id int) error {
	query := `UPDATE cat.at_api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// =========================================================================
// SPMB Integration
// =========================================================================

func (r *integrationRepository) GetSPMBActiveExams(ctx context.Context, refDate string) ([]*dto.SPMBActiveExamDTO, error) {
	if refDate == "" {
		refDate = time.Now().Format("2006-01-02")
	}

	query := `
		SELECT 
			u.idujian,
			u.namaujian,
			u.idperiode,
			COALESCE(p.namaperiode, '') as namaperiode,
			COALESCE(p.isonline, 0) as isonline,
			COALESCE(u.nilaiminimal, 0) as nilaiminimal,
			COALESCE(u.keterangan, '') as keterangan,
			COALESCE(MIN(COALESCE(j.tglmulai, r.tglmulai))::text, '') as start_date,
			COALESCE(MAX(COALESCE(j.tglselesai, r.tglselesai, j.tglmulai, r.tglmulai))::text, '') as end_date,
			COUNT(DISTINCT j.idjadwalujian) as total_sessions,
			COALESCE(SUM(r.jumlahpeserta), 0) as total_capacity,
			(
				SELECT COUNT(DISTINCT pu.kodepeserta)
				FROM cat.at_pesertaujian pu
				WHERE pu.idujian = u.idujian AND (pu.softdelete = '0' OR pu.softdelete IS NULL)
			) as registered_count
		FROM cat.at_ujian u
		LEFT JOIN cat.at_periode p ON p.idperiode = u.idperiode
		LEFT JOIN cat.at_jadwalujian j ON j.idujian = u.idujian AND (j.softdelete = '0' OR j.softdelete IS NULL)
		LEFT JOIN cat.at_ruangujian r ON r.idjadwalujian = j.idjadwalujian AND (r.softdelete = '0' OR r.softdelete IS NULL)
		WHERE (u.softdelete = '0' OR u.softdelete IS NULL)
		GROUP BY u.idujian, u.namaujian, u.idperiode, p.namaperiode, p.isonline, u.nilaiminimal, u.keterangan
		HAVING MAX(COALESCE(j.tglselesai, r.tglselesai, j.tglmulai, r.tglmulai)) >= $1::date
		ORDER BY start_date ASC, u.idujian DESC
	`

	rows, err := r.db.QueryContext(ctx, query, refDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query active exams: %w", err)
	}
	defer rows.Close()

	results := make([]*dto.SPMBActiveExamDTO, 0)
	for rows.Next() {
		var item dto.SPMBActiveExamDTO
		var rawCapacity, regCount, isOnline int
		if scanErr := rows.Scan(
			&item.IDUjian,
			&item.NamaUjian,
			&item.IDPeriode,
			&item.NamaPeriode,
			&isOnline,
			&item.NilaiMinimal,
			&item.Keterangan,
			&item.TglMulai,
			&item.TglSelesai,
			&item.TotalSesi,
			&rawCapacity,
			&regCount,
		); scanErr != nil {
			return nil, scanErr
		}

		item.IsOnline = (isOnline == 1)
		item.JumlahPeserta = regCount

		if item.IsOnline {
			item.Metode = "Online / Daring"
			item.TotalKapasitas = 0 // Bebas / tidak dibatasi PC
			item.SisaKuota = 999999
			item.IsOpen = true
		} else {
			item.Metode = "Offline / Di Kampus"
			item.TotalKapasitas = rawCapacity
			rem := rawCapacity - regCount
			if rem < 0 {
				rem = 0
			}
			item.SisaKuota = rem
			// Pendaftaran SPMB tetap buka agar pendaftar tidak terblokir
			item.IsOpen = true
		}

		results = append(results, &item)
	}

	return results, nil
}

func (r *integrationRepository) RegisterSPMBParticipant(ctx context.Context, req *dto.SPMBRegisterRequestDTO, plainPassword string) (*dto.SPMBRegisterResponseDTO, error) {
	// 1. Validate exam exists and get period & isonline
	var exam struct {
		IDUjian     int    `db:"idujian"`
		NamaUjian   string `db:"namaujian"`
		IDPeriode   int    `db:"idperiode"`
		NamaPeriode string `db:"namaperiode"`
		IsOnline    int    `db:"isonline"`
	}

	examID := req.IDUjian
	if examID <= 0 {
		examID = req.ExamID
	}

	err := r.db.GetContext(ctx, &exam, `
		SELECT u.idujian, u.namaujian, u.idperiode, COALESCE(p.namaperiode, '') as namaperiode,
		       COALESCE(p.isonline, 0) as isonline
		FROM cat.at_ujian u
		LEFT JOIN cat.at_periode p ON p.idperiode = u.idperiode
		WHERE u.idujian = $1 AND (u.softdelete = '0' OR u.softdelete IS NULL)
	`, examID)
	if err != nil {
		return nil, fmt.Errorf("ujian #%d tidak ditemukan atau tidak aktif: %w", examID, err)
	}

	// 2. Determine kodepeserta
	kodepeserta := req.NomorUjian
	if kodepeserta == "" {
		kodepeserta = req.IDPendaftar
	}
	if len(kodepeserta) > 20 {
		kodepeserta = kodepeserta[:20]
	}

	// 3. Hash password
	hasher := md5.New()
	hasher.Write([]byte(plainPassword))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))

	jk := req.JK
	if jk != "L" && jk != "P" {
		jk = "L"
	}

	// 4. Upsert cat.at_peserta (Defensive: check existence first to be resilient to missing constraints)
	var existingPeserta int
	_ = r.db.GetContext(ctx, &existingPeserta, "SELECT COUNT(*) FROM cat.at_peserta WHERE kodepeserta = $1", kodepeserta)
	if existingPeserta > 0 {
		_, err = r.db.ExecContext(ctx, `
			UPDATE cat.at_peserta SET
				nama = $1,
				jk = $2,
				hp = $3,
				email = $4,
				alamat = $5,
				idpendaftar = $6,
				password = CASE WHEN $7 != '' THEN $7 ELSE cat.at_peserta.password END,
				hint = CASE WHEN $8 != '' THEN $8 ELSE cat.at_peserta.hint END,
				isvalid = 1,
				softdelete = '0',
				t_updatetime = NOW(),
				t_updateact = 'u-spmb-api'
			WHERE kodepeserta = $9
		`, req.Nama, jk, req.HP, req.Email, req.Alamat, req.IDPendaftar,
			hashedPassword, plainPassword, kodepeserta)
		if err != nil {
			return nil, fmt.Errorf("gagal mengupdate data peserta di master: %w", err)
		}
	} else {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO cat.at_peserta (
				kodepeserta, idperiode, nama, jk, hp, email, alamat, idkota,
				password, hint, sumberdata, isvalid, isaktif, islogin,
				kodereferensi, idpendaftar, softdelete, t_updatetime, t_updateact
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8,
				$9, $10, 'P', 1, 0, 1,
				$11, $12, '0', NOW(), 'i-spmb-api'
			)
		`, kodepeserta, exam.IDPeriode, req.Nama, jk, req.HP, req.Email, req.Alamat, req.IDKota,
			hashedPassword, plainPassword, kodepeserta, req.IDPendaftar)
		if err != nil {
			// Fallback update in case of concurrent insert
			_, err = r.db.ExecContext(ctx, `
				UPDATE cat.at_peserta SET
					nama = $1, jk = $2, hp = $3, email = $4, alamat = $5, idpendaftar = $6,
					password = CASE WHEN $7 != '' THEN $7 ELSE cat.at_peserta.password END,
					hint = CASE WHEN $8 != '' THEN $8 ELSE cat.at_peserta.hint END,
					isvalid = 1, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-spmb-api'
				WHERE kodepeserta = $9
			`, req.Nama, jk, req.HP, req.Email, req.Alamat, req.IDPendaftar,
				hashedPassword, plainPassword, kodepeserta)
			if err != nil {
				return nil, fmt.Errorf("gagal mendaftarkan data peserta di master: %w", err)
			}
		}
	}

	// 5. Upsert cat.at_pesertaujian
	var existingPesertaUjian int
	_ = r.db.GetContext(ctx, &existingPesertaUjian, "SELECT COUNT(*) FROM cat.at_pesertaujian WHERE kodepeserta = $1 AND idujian = $2", kodepeserta, exam.IDUjian)
	if existingPesertaUjian > 0 {
		_, err = r.db.ExecContext(ctx, `
			UPDATE cat.at_pesertaujian
			SET softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-spmb-api'
			WHERE kodepeserta = $1 AND idujian = $2
		`, kodepeserta, exam.IDUjian)
		if err != nil {
			return nil, fmt.Errorf("gagal memperbarui status peserta di ujian #%d: %w", exam.IDUjian, err)
		}
	} else {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO cat.at_pesertaujian (kodepeserta, idujian, softdelete, t_updatetime, t_updateact)
			VALUES ($1, $2, '0', NOW(), 'i-spmb-api')
		`, kodepeserta, exam.IDUjian)
		if err != nil {
			_, err = r.db.ExecContext(ctx, `
				UPDATE cat.at_pesertaujian
				SET softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-spmb-api'
				WHERE kodepeserta = $1 AND idujian = $2
			`, kodepeserta, exam.IDUjian)
			if err != nil {
				return nil, fmt.Errorf("gagal mendaftarkan peserta ke ujian #%d: %w", exam.IDUjian, err)
			}
		}
	}

	resp := &dto.SPMBRegisterResponseDTO{
		KodePeserta: kodepeserta,
		IDPendaftar: req.IDPendaftar,
		IDUjian:     exam.IDUjian,
		NamaUjian:   exam.NamaUjian,
		IDPeriode:   exam.IDPeriode,
		NamaPeriode: exam.NamaPeriode,
		CBTCredentials: dto.SPMBCredentialsDTO{
			Username: kodepeserta,
			Password: plainPassword,
		},
		Schedule: dto.SPMBScheduleInfoDTO{
			IsPlotted: false,
		},
	}

	// 6. Auto-plotting into available session & room if requested
	shouldPlot := true
	if req.AutoPlotSession != nil {
		shouldPlot = *req.AutoPlotSession
	}

	if shouldPlot {
		var availableSession struct {
			IDJadwalUjian int            `db:"idjadwalujian"`
			IDRuangUjian  int            `db:"idruangujian"`
			NamaRuang     sql.NullString `db:"namaruang"`
			TglUjian      sql.NullTime   `db:"tglujian"`
			JamMulai      sql.NullString `db:"jammulai"`
			JamSelesai    sql.NullString `db:"jamselesai"`
			Durasi        int            `db:"durasi"`
		}

		var findQuery string
		if exam.IsOnline == 1 {
			// =========================================================================
			// MODE UJIAN ONLINE / DARING: Kapasitas Bebas (Unlimited)
			// Selalu plot peserta ke sesi ujian daring aktif tanpa membatasi kuota PC
			// =========================================================================
			findQuery = `
				SELECT 
					j.idjadwalujian,
					COALESCE(r.idruangujian, 0) as idruangujian,
					COALESCE(rm.namaruang, r.koderuang, 'Daring / Online') as namaruang,
					COALESCE(r.tglmulai, j.tglmulai) as tglujian,
					COALESCE(r.waktumulai, j.waktumulai, '08:00') as jammulai,
					COALESCE(r.waktuselesai, j.waktuselesai, '09:30') as jamselesai,
					COALESCE(j.waktupengerjaan::integer, 90) as durasi
				FROM cat.at_jadwalujian j
				LEFT JOIN cat.at_ruangujian r ON r.idjadwalujian = j.idjadwalujian AND (r.softdelete = '0' OR r.softdelete IS NULL)
				LEFT JOIN cat.at_ruang rm ON rm.koderuang = r.koderuang
				WHERE j.idujian = $1 AND (j.softdelete = '0' OR j.softdelete IS NULL)
				ORDER BY 
					CASE WHEN COALESCE(r.tglselesai, r.tglmulai, j.tglselesai, j.tglmulai, NOW())::date >= CURRENT_DATE THEN 0 ELSE 1 END ASC,
					COALESCE(r.tglmulai, j.tglmulai) ASC, 
					j.idjadwalujian ASC
				LIMIT 1
			`
		} else {
			// =========================================================================
			// MODE UJIAN OFFLINE / LAB FISIK: Waterfall Plotting Berdasarkan Kapasitas PC
			// Mengisi ruangan berurutan (Sesi 1 -> Sesi 2 -> dst.) yang masih ada kursi kosong
			// =========================================================================
			findQuery = `
				SELECT 
					j.idjadwalujian,
					COALESCE(r.idruangujian, 0) as idruangujian,
					COALESCE(rm.namaruang, r.koderuang, 'Lab Utama') as namaruang,
					COALESCE(r.tglmulai, j.tglmulai) as tglujian,
					COALESCE(r.waktumulai, j.waktumulai, '08:00') as jammulai,
					COALESCE(r.waktuselesai, j.waktuselesai, '09:30') as jamselesai,
					COALESCE(j.waktupengerjaan::integer, 90) as durasi
				FROM cat.at_jadwalujian j
				LEFT JOIN cat.at_ruangujian r ON r.idjadwalujian = j.idjadwalujian AND (r.softdelete = '0' OR r.softdelete IS NULL)
				LEFT JOIN cat.at_ruang rm ON rm.koderuang = r.koderuang
				WHERE j.idujian = $1 AND (j.softdelete = '0' OR j.softdelete IS NULL)
				  AND COALESCE(r.tglselesai, r.tglmulai, j.tglselesai, j.tglmulai, NOW())::date >= CURRENT_DATE
				  AND (
				      r.jumlahpeserta IS NULL 
				      OR r.jumlahpeserta = 0 
				      OR (
				          SELECT COUNT(*) FROM cat.at_jadwalpeserta jp 
				          WHERE jp.idruangujian = r.idruangujian 
				            AND jp.idjadwalujian = j.idjadwalujian 
				            AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
				      ) < r.jumlahpeserta
				  )
				ORDER BY COALESCE(r.tglmulai, j.tglmulai) ASC, j.idjadwalujian ASC, COALESCE(r.prioritas, 1) ASC
				LIMIT 1
			`
		}

		err = r.db.GetContext(ctx, &availableSession, findQuery, exam.IDUjian)
		if err == nil && availableSession.IDJadwalUjian > 0 {
			var ruangArg interface{} = nil
			if availableSession.IDRuangUjian > 0 {
				ruangArg = availableSession.IDRuangUjian
			}

			// Plot into cat.at_jadwalpeserta
			var existingJadwalPeserta int
			_ = r.db.GetContext(ctx, &existingJadwalPeserta, "SELECT COUNT(*) FROM cat.at_jadwalpeserta WHERE kodepeserta = $1 AND idjadwalujian = $2", kodepeserta, availableSession.IDJadwalUjian)
			if existingJadwalPeserta > 0 {
				_, plotErr := r.db.ExecContext(ctx, `
					UPDATE cat.at_jadwalpeserta
					SET idruangujian = $3, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-spmb-plot'
					WHERE kodepeserta = $1 AND idjadwalujian = $2
				`, kodepeserta, availableSession.IDJadwalUjian, ruangArg)
				if plotErr == nil {
					resp.Schedule.IsPlotted = true
					resp.Schedule.IDJadwalUjian = availableSession.IDJadwalUjian
				}
			} else {
				_, plotErr := r.db.ExecContext(ctx, `
					INSERT INTO cat.at_jadwalpeserta (
						kodepeserta, idjadwalujian, idruangujian, is_locked, risk_level,
						softdelete, t_updatetime, t_updateact
					) VALUES (
						$1, $2, $3, 0, 'NORMAL', '0', NOW(), 'i-spmb-plot'
					)
				`, kodepeserta, availableSession.IDJadwalUjian, ruangArg)
				if plotErr == nil {
					resp.Schedule.IsPlotted = true
					resp.Schedule.IDJadwalUjian = availableSession.IDJadwalUjian
				} else {
					_, updateErr := r.db.ExecContext(ctx, `
						UPDATE cat.at_jadwalpeserta
						SET idruangujian = $3, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-spmb-plot'
						WHERE kodepeserta = $1 AND idjadwalujian = $2
					`, kodepeserta, availableSession.IDJadwalUjian, ruangArg)
					if updateErr == nil {
						resp.Schedule.IsPlotted = true
						resp.Schedule.IDJadwalUjian = availableSession.IDJadwalUjian
					}
				}
			}
			if availableSession.NamaRuang.Valid && availableSession.NamaRuang.String != "" {
				resp.Schedule.NamaRuang = availableSession.NamaRuang.String
			} else if exam.IsOnline == 1 {
				resp.Schedule.NamaRuang = "Daring / Online"
			} else {
				resp.Schedule.NamaRuang = "Lab Utama"
			}
			if availableSession.TglUjian.Valid {
				resp.Schedule.TglUjian = availableSession.TglUjian.Time.Format("2006-01-02")
			}
			if availableSession.JamMulai.Valid {
				resp.Schedule.JamMulai = formatTimeHi(availableSession.JamMulai.String)
			}
			if availableSession.JamSelesai.Valid && strings.TrimSpace(availableSession.JamSelesai.String) != "" {
				resp.Schedule.JamSelesai = formatTimeHi(availableSession.JamSelesai.String)
			}
			// If JamSelesai is still empty but JamMulai & Durasi are present, calculate JamSelesai
			if resp.Schedule.JamSelesai == "" && resp.Schedule.JamMulai != "" && availableSession.Durasi > 0 {
				parts := strings.Split(resp.Schedule.JamMulai, ":")
				if len(parts) >= 2 {
					hh, _ := strconv.Atoi(parts[0])
					mm, _ := strconv.Atoi(parts[1])
					totalMin := hh*60 + mm + availableSession.Durasi
					resp.Schedule.JamSelesai = fmt.Sprintf("%02d:%02d", (totalMin/60)%24, totalMin%60)
				}
			}
			resp.Schedule.WaktuPengerjaan = availableSession.Durasi
			if exam.IsOnline == 1 {
				resp.Schedule.Catatan = "Ujian dilaksanakan secara Daring / Online."
			} else {
				resp.Schedule.Catatan = "Ujian dilaksanakan secara Luring di Lab Komputer Kampus."
			}
		} else {
			resp.Schedule.IsPlotted = false
			if exam.IsOnline == 1 {
				resp.Schedule.Catatan = "Sesi ujian daring belum aktif atau belum dibuat oleh panitia."
			} else {
				resp.Schedule.Catatan = "Kapasitas sesi lab saat ini telah penuh. Peserta berhasil terdaftar dan masuk antrean pembagian sesi tambahan oleh panitia CBT."
			}
		}
	} else {
		resp.Schedule.IsPlotted = false
		resp.Schedule.Catatan = "Auto-plotting dinonaktifkan atas permintaan klien."
	}

	return resp, nil
}

func formatTimeHi(str string) string {
	s := strings.TrimSpace(str)
	if s == "" {
		return ""
	}
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		if len(parts) >= 2 {
			hh, errH := strconv.Atoi(parts[0])
			mm, errM := strconv.Atoi(parts[1])
			if errH == nil && errM == nil {
				return fmt.Sprintf("%02d:%02d", hh, mm)
			}
		}
	}
	clean := strings.ReplaceAll(s, ":", "")
	if len(clean) == 4 {
		hh, errH := strconv.Atoi(clean[:2])
		mm, errM := strconv.Atoi(clean[2:])
		if errH == nil && errM == nil {
			return fmt.Sprintf("%02d:%02d", hh, mm)
		}
	} else if len(clean) == 3 {
		hh, errH := strconv.Atoi(clean[:1])
		mm, errM := strconv.Atoi(clean[1:])
		if errH == nil && errM == nil {
			return fmt.Sprintf("%02d:%02d", hh, mm)
		}
	}
	return s
}

func (r *integrationRepository) CreateAccessLog(ctx context.Context, log *entity.APIAccessLog) error {
	query := `
		INSERT INTO cat.at_api_access_logs (
			api_key_id, client_name, ip_address, method, endpoint,
			status_code, response_time_ms, user_agent, error_message,
			request_body, response_body, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		log.APIKeyID, log.ClientName, log.IPAddress, log.Method, log.Endpoint,
		log.StatusCode, log.ResponseTimeMS, log.UserAgent, log.ErrorMessage,
		log.RequestBody, log.ResponseBody,
	)
	return err
}

func (r *integrationRepository) GetAccessLogs(ctx context.Context, apiKeyID int, page, perPage int, search string) ([]*entity.APIAccessLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 15
	}
	offset := (page - 1) * perPage

	baseWhere := "1=1"
	args := []interface{}{}
	argIdx := 1

	if apiKeyID > 0 {
		baseWhere += fmt.Sprintf(" AND api_key_id = $%d", argIdx)
		args = append(args, apiKeyID)
		argIdx++
	}

	if search != "" {
		baseWhere += fmt.Sprintf(" AND (client_name ILIKE $%d OR endpoint ILIKE $%d OR ip_address ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM cat.at_api_access_logs WHERE %s", baseWhere)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, api_key_id, client_name, ip_address, method, endpoint, status_code, response_time_ms, user_agent, error_message, request_body, response_body, created_at
		FROM cat.at_api_access_logs
		WHERE %s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d
	`, baseWhere, argIdx, argIdx+1)
	args = append(args, perPage, offset)

	var logs []*entity.APIAccessLog
	err := r.db.SelectContext(ctx, &logs, query, args...)
	if logs == nil {
		logs = []*entity.APIAccessLog{}
	}
	return logs, total, err
}

func (r *integrationRepository) GetSPMBParticipant(ctx context.Context, idPendaftar string) (*dto.SPMBParticipantDetailDTO, error) {
	query := `
		SELECT 
			p.kodepeserta,
			COALESCE(p.idpendaftar, p.kodepeserta) as idpendaftar,
			COALESCE(p.nama, '') as nama,
			COALESCE(p.jk, 'L') as jk,
			COALESCE(p.email, '') as email,
			COALESCE(p.hp, '') as hp,
			COALESCE(p.alamat, '') as alamat,
			p.idkota,
			COALESCE(u.idujian, 0) as idujian,
			COALESCE(u.namaujian, '') as namaujian,
			COALESCE(per.idperiode, 0) as idperiode,
			COALESCE(per.namaperiode, '') as namaperiode,
			COALESCE(per.isonline, 0) as isonline,
			COALESCE(p.hint, p.kodepeserta) as plain_password,
			COALESCE(p.t_updatetime, NOW()) as t_updatetime,
			COALESCE(jp.idjadwalujian, 0) as idjadwalujian,
			COALESCE(rm.namaruang, ru.koderuang, '') as namaruang,
			ru.tglmulai as tglujian_ruang,
			ju.tglmulai as tglujian_jadwal,
			COALESCE(ru.waktumulai, ju.waktumulai, '') as jammulai,
			COALESCE(ru.waktuselesai, ju.waktuselesai, '') as jamselesai,
			COALESCE(ju.waktupengerjaan::integer, 90) as durasi
		FROM cat.at_peserta p
		LEFT JOIN cat.at_pesertaujian pu ON pu.kodepeserta = p.kodepeserta AND (pu.softdelete = '0' OR pu.softdelete IS NULL)
		LEFT JOIN cat.at_ujian u ON u.idujian = pu.idujian AND (u.softdelete = '0' OR u.softdelete IS NULL)
		LEFT JOIN cat.at_periode per ON per.idperiode = u.idperiode
		LEFT JOIN cat.at_jadwalpeserta jp ON jp.kodepeserta = p.kodepeserta AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
		LEFT JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian AND (ju.softdelete = '0' OR ju.softdelete IS NULL)
		LEFT JOIN cat.at_ruangujian ru ON ru.idruangujian = jp.idruangujian AND (ru.softdelete = '0' OR ru.softdelete IS NULL)
		LEFT JOIN cat.at_ruang rm ON rm.koderuang = ru.koderuang
		WHERE (p.idpendaftar = $1 OR p.kodepeserta = $1) AND (p.softdelete = '0' OR p.softdelete IS NULL)
		ORDER BY p.t_updatetime DESC
		LIMIT 1
	`

	var row struct {
		KodePeserta    string       `db:"kodepeserta"`
		IDPendaftar    string       `db:"idpendaftar"`
		Nama           string       `db:"nama"`
		JK             string       `db:"jk"`
		Email          string       `db:"email"`
		HP             string       `db:"hp"`
		Alamat         string       `db:"alamat"`
		IDKota         *int         `db:"idkota"`
		IDUjian        int          `db:"idujian"`
		NamaUjian      string       `db:"namaujian"`
		IDPeriode      int          `db:"idperiode"`
		NamaPeriode    string       `db:"namaperiode"`
		IsOnline       int          `db:"isonline"`
		PlainPassword  string       `db:"plain_password"`
		TUpdateTime    time.Time    `db:"t_updatetime"`
		IDJadwalUjian  int          `db:"idjadwalujian"`
		NamaRuang      string       `db:"namaruang"`
		TglUjianRuang  sql.NullTime `db:"tglujian_ruang"`
		TglUjianJadwal sql.NullTime `db:"tglujian_jadwal"`
		JamMulai       string       `db:"jammulai"`
		JamSelesai     string       `db:"jamselesai"`
		Durasi         int          `db:"durasi"`
	}

	err := r.db.GetContext(ctx, &row, query, idPendaftar)
	if err != nil {
		return nil, fmt.Errorf("peserta dengan ID/Nomor '%s' tidak ditemukan: %w", idPendaftar, err)
	}

	isOnline := (row.IsOnline == 1)
	metode := "Offline / Di Kampus"
	if isOnline {
		metode = "Online / Daring"
	}

	isPlotted := row.IDJadwalUjian > 0
	tglStr := ""
	if row.TglUjianRuang.Valid {
		tglStr = row.TglUjianRuang.Time.Format("2006-01-02")
	} else if row.TglUjianJadwal.Valid {
		tglStr = row.TglUjianJadwal.Time.Format("2006-01-02")
	}

	jamMulai := ""
	if row.JamMulai != "" {
		jamMulai = formatTimeHi(row.JamMulai)
	}
	jamSelesai := ""
	if row.JamSelesai != "" {
		jamSelesai = formatTimeHi(row.JamSelesai)
	}
	if jamSelesai == "" && jamMulai != "" && row.Durasi > 0 {
		parts := strings.Split(jamMulai, ":")
		if len(parts) >= 2 {
			hh, _ := strconv.Atoi(parts[0])
			mm, _ := strconv.Atoi(parts[1])
			totalMin := hh*60 + mm + row.Durasi
			jamSelesai = fmt.Sprintf("%02d:%02d", (totalMin/60)%24, totalMin%60)
		}
	}

	catatan := ""
	if isPlotted {
		if isOnline {
			catatan = "Ujian dilaksanakan secara Daring / Online."
		} else {
			catatan = "Ujian dilaksanakan secara Luring di Lab Komputer Kampus."
		}
	} else {
		if isOnline {
			catatan = "Sesi ujian daring belum aktif."
		} else {
			catatan = "Kapasitas lab saat ini telah penuh. Peserta terdaftar dalam antrean pembagian sesi."
		}
	}

	namaRuang := row.NamaRuang
	if namaRuang == "" {
		if isOnline {
			namaRuang = "Daring / Online"
		} else {
			namaRuang = "Lab Utama"
		}
	}

	return &dto.SPMBParticipantDetailDTO{
		KodePeserta: row.KodePeserta,
		IDPendaftar: row.IDPendaftar,
		Nama:        row.Nama,
		JK:          row.JK,
		Email:       row.Email,
		HP:          row.HP,
		Alamat:      row.Alamat,
		IDKota:      row.IDKota,
		IDUjian:     row.IDUjian,
		NamaUjian:   row.NamaUjian,
		IDPeriode:   row.IDPeriode,
		NamaPeriode: row.NamaPeriode,
		IsOnline:    isOnline,
		Metode:      metode,
		Credentials: dto.SPMBCredentialsDTO{
			Username: row.KodePeserta,
			Password: row.PlainPassword,
		},
		Schedule: dto.SPMBScheduleInfoDTO{
			IsPlotted:       isPlotted,
			IDJadwalUjian:   row.IDJadwalUjian,
			NamaRuang:       namaRuang,
			TglUjian:        tglStr,
			JamMulai:        jamMulai,
			JamSelesai:      jamSelesai,
			WaktuPengerjaan: row.Durasi,
			Catatan:         catatan,
		},
		CreatedAt: row.TUpdateTime,
	}, nil
}

func (r *integrationRepository) UpdateSPMBParticipant(ctx context.Context, idPendaftar string, req *dto.SPMBUpdateParticipantDTO) (*dto.SPMBParticipantDetailDTO, error) {
	existing, err := r.GetSPMBParticipant(ctx, idPendaftar)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("peserta dengan ID/Nomor '%s' tidak ditemukan: %w", idPendaftar, err)
	}

	kodepeserta := existing.KodePeserta

	nama := existing.Nama
	if strings.TrimSpace(req.Nama) != "" {
		nama = strings.TrimSpace(req.Nama)
	}
	email := existing.Email
	if strings.TrimSpace(req.Email) != "" {
		email = strings.TrimSpace(req.Email)
	}
	hp := existing.HP
	if strings.TrimSpace(req.HP) != "" {
		hp = strings.TrimSpace(req.HP)
	}
	alamat := existing.Alamat
	if strings.TrimSpace(req.Alamat) != "" {
		alamat = strings.TrimSpace(req.Alamat)
	}
	jk := existing.JK
	if req.JK == "L" || req.JK == "P" {
		jk = req.JK
	}
	idkota := existing.IDKota
	if req.IDKota != nil {
		idkota = req.IDKota
	}

	if strings.TrimSpace(req.Password) != "" {
		hasher := md5.New()
		hasher.Write([]byte(strings.TrimSpace(req.Password)))
		hashedPassword := hex.EncodeToString(hasher.Sum(nil))

		_, err = r.db.ExecContext(ctx, `
			UPDATE cat.at_peserta
			SET nama = $1, email = $2, hp = $3, alamat = $4, jk = $5, idkota = $6,
			    password = $7, hint = $8, t_updatetime = NOW(), t_updateact = 'u-spmb-api'
			WHERE kodepeserta = $9
		`, nama, email, hp, alamat, jk, idkota, hashedPassword, strings.TrimSpace(req.Password), kodepeserta)
	} else {
		_, err = r.db.ExecContext(ctx, `
			UPDATE cat.at_peserta
			SET nama = $1, email = $2, hp = $3, alamat = $4, jk = $5, idkota = $6,
			    t_updatetime = NOW(), t_updateact = 'u-spmb-api'
			WHERE kodepeserta = $7
		`, nama, email, hp, alamat, jk, idkota, kodepeserta)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui data peserta: %w", err)
	}

	// If IDUjian is provided and different, reassign to new exam & schedule
	if req.IDUjian > 0 && req.IDUjian != existing.IDUjian {
		_, _ = r.db.ExecContext(ctx, `UPDATE cat.at_pesertaujian SET softdelete = '1', t_updatetime = NOW() WHERE kodepeserta = $1 AND idujian = $2`, kodepeserta, existing.IDUjian)
		_, _ = r.db.ExecContext(ctx, `UPDATE cat.at_jadwalpeserta SET softdelete = '1', t_updatetime = NOW() WHERE kodepeserta = $1 AND idjadwalujian IN (SELECT idjadwalujian FROM cat.at_jadwalujian WHERE idujian = $2)`, kodepeserta, existing.IDUjian)

		registerReq := &dto.SPMBRegisterRequestDTO{
			IDUjian:     req.IDUjian,
			IDPendaftar: idPendaftar,
			NomorUjian:  kodepeserta,
			Nama:        nama,
			JK:          jk,
			Email:       email,
			HP:          hp,
			Alamat:      alamat,
			IDKota:      idkota,
		}
		_, err = r.RegisterSPMBParticipant(ctx, registerReq, existing.Credentials.Password)
		if err != nil {
			return nil, fmt.Errorf("gagal memindahkan peserta ke paket ujian baru: %w", err)
		}
	}

	return r.GetSPMBParticipant(ctx, idPendaftar)
}

func (r *integrationRepository) DeleteSPMBParticipant(ctx context.Context, idPendaftar string) error {
	existing, err := r.GetSPMBParticipant(ctx, idPendaftar)
	if err != nil || existing == nil {
		return fmt.Errorf("peserta dengan ID/Nomor '%s' tidak ditemukan: %w", idPendaftar, err)
	}

	kodepeserta := existing.KodePeserta

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi database: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, "UPDATE cat.at_jadwalpeserta SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-spmb-api' WHERE kodepeserta = $1", kodepeserta); err != nil {
		return fmt.Errorf("gagal membatalkan jadwal peserta: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "UPDATE cat.at_pesertaujian SET softdelete = '1', t_updatetime = NOW(), t_updateact = 'd-spmb-api' WHERE kodepeserta = $1", kodepeserta); err != nil {
		return fmt.Errorf("gagal membatalkan pendaftaran ujian: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "UPDATE cat.at_peserta SET softdelete = '1', isaktif = 0, islogin = 0, t_updatetime = NOW(), t_updateact = 'd-spmb-api' WHERE kodepeserta = $1", kodepeserta); err != nil {
		return fmt.Errorf("gagal menonaktifkan peserta: %w", err)
	}

	return tx.Commit()
}

func (r *integrationRepository) GetUnplottedQueueSummary(ctx context.Context) (*dto.UnplottedQueueSummaryDTO, error) {
	query := `
		SELECT 
			u.idujian,
			u.namaujian,
			u.idperiode,
			COALESCE(p.namaperiode, '') as namaperiode,
			COALESCE(SUM(r.jumlahpeserta), 0)::integer as total_capacity,
			(
				SELECT COUNT(DISTINCT pu.kodepeserta)
				FROM cat.at_pesertaujian pu
				WHERE pu.idujian = u.idujian AND (pu.softdelete = '0' OR pu.softdelete IS NULL)
			) as total_registered,
			(
				SELECT COUNT(DISTINCT pu.kodepeserta)
				FROM cat.at_pesertaujian pu
				WHERE pu.idujian = u.idujian 
				  AND (pu.softdelete = '0' OR pu.softdelete IS NULL)
				  AND NOT EXISTS (
				      SELECT 1 
				      FROM cat.at_jadwalpeserta jp 
				      JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian
				      WHERE jp.kodepeserta = pu.kodepeserta 
				        AND ju.idujian = u.idujian 
				        AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
				  )
			) as unplotted_count
		FROM cat.at_ujian u
		LEFT JOIN cat.at_periode p ON p.idperiode = u.idperiode
		LEFT JOIN cat.at_jadwalujian j ON j.idujian = u.idujian AND (j.softdelete = '0' OR j.softdelete IS NULL)
		LEFT JOIN cat.at_ruangujian r ON r.idjadwalujian = j.idjadwalujian AND (r.softdelete = '0' OR r.softdelete IS NULL)
		WHERE (u.softdelete = '0' OR u.softdelete IS NULL)
		  AND COALESCE(p.isonline, 0) = 0
		GROUP BY u.idujian, u.namaujian, u.idperiode, p.namaperiode
		HAVING (
			SELECT COUNT(DISTINCT pu.kodepeserta)
			FROM cat.at_pesertaujian pu
			WHERE pu.idujian = u.idujian 
			  AND (pu.softdelete = '0' OR pu.softdelete IS NULL)
			  AND NOT EXISTS (
			      SELECT 1 
			      FROM cat.at_jadwalpeserta jp 
			      JOIN cat.at_jadwalujian ju ON ju.idjadwalujian = jp.idjadwalujian
			      WHERE jp.kodepeserta = pu.kodepeserta 
			        AND ju.idujian = u.idujian 
			        AND (jp.softdelete = '0' OR jp.softdelete IS NULL)
			  )
		) > 0
		ORDER BY unplotted_count DESC, u.idujian DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unplotted queue summary: %w", err)
	}
	defer rows.Close()

	summary := &dto.UnplottedQueueSummaryDTO{
		TotalUnplotted: 0,
		Exams:          []dto.UnplottedQueueItemDTO{},
	}

	for rows.Next() {
		var item dto.UnplottedQueueItemDTO
		if scanErr := rows.Scan(
			&item.IDUjian,
			&item.NamaUjian,
			&item.IDPeriode,
			&item.NamaPeriode,
			&item.TotalCapacity,
			&item.TotalRegistered,
			&item.UnplottedCount,
		); scanErr != nil {
			return nil, scanErr
		}
		summary.TotalUnplotted += item.UnplottedCount
		summary.Exams = append(summary.Exams, item)
	}

	return summary, nil
}
