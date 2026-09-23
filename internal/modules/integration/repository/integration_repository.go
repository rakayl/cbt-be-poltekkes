package repository

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
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
		GROUP BY u.idujian, u.namaujian, u.idperiode, p.namaperiode, u.nilaiminimal, u.keterangan
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
		var rawCapacity, regCount int
		if scanErr := rows.Scan(
			&item.ExamID,
			&item.ExamName,
			&item.PeriodID,
			&item.PeriodName,
			&item.PassingGrade,
			&item.Description,
			&item.StartDate,
			&item.EndDate,
			&item.TotalSessions,
			&rawCapacity,
			&regCount,
		); scanErr != nil {
			return nil, scanErr
		}

		item.TotalCapacity = rawCapacity
		item.RegisteredCount = regCount
		rem := rawCapacity - regCount
		if rem < 0 {
			rem = 0
		}
		item.RemainingQuota = rem
		item.IsOpen = (rawCapacity == 0 || rem > 0)

		results = append(results, &item)
	}

	return results, nil
}

func (r *integrationRepository) RegisterSPMBParticipant(ctx context.Context, req *dto.SPMBRegisterRequestDTO, plainPassword string) (*dto.SPMBRegisterResponseDTO, error) {
	// 1. Validate exam exists and get period
	var exam struct {
		IDUjian     int    `db:"idujian"`
		NamaUjian   string `db:"namaujian"`
		IDPeriode   int    `db:"idperiode"`
		NamaPeriode string `db:"namaperiode"`
	}
	err := r.db.GetContext(ctx, &exam, `
		SELECT u.idujian, u.namaujian, u.idperiode, COALESCE(p.namaperiode, '') as namaperiode
		FROM cat.at_ujian u
		LEFT JOIN cat.at_periode p ON p.idperiode = u.idperiode
		WHERE u.idujian = $1 AND (u.softdelete = '0' OR u.softdelete IS NULL)
	`, req.ExamID)
	if err != nil {
		return nil, fmt.Errorf("ujian #%d tidak ditemukan atau tidak aktif: %w", req.ExamID, err)
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

	// 4. Upsert cat.at_peserta
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
		ON CONFLICT (kodepeserta) DO UPDATE SET
			nama = EXCLUDED.nama,
			jk = EXCLUDED.jk,
			hp = EXCLUDED.hp,
			email = EXCLUDED.email,
			alamat = EXCLUDED.alamat,
			idpendaftar = EXCLUDED.idpendaftar,
			password = CASE WHEN $9 != '' THEN $9 ELSE cat.at_peserta.password END,
			hint = CASE WHEN $10 != '' THEN $10 ELSE cat.at_peserta.hint END,
			isvalid = 1,
			softdelete = '0',
			t_updatetime = NOW(),
			t_updateact = 'u-spmb-api'
	`, kodepeserta, exam.IDPeriode, req.Nama, jk, req.HP, req.Email, req.Alamat, req.IDKota,
		hashedPassword, plainPassword, kodepeserta, req.IDPendaftar)
	if err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan data peserta di master: %w", err)
	}

	// 5. Upsert cat.at_pesertaujian
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO cat.at_pesertaujian (kodepeserta, idujian, softdelete, t_updatetime, t_updateact)
		VALUES ($1, $2, '0', NOW(), 'i-spmb-api')
		ON CONFLICT (kodepeserta, idujian) DO UPDATE
		SET softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-spmb-api'
	`, kodepeserta, exam.IDUjian)
	if err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan peserta ke ujian #%d: %w", exam.IDUjian, err)
	}

	resp := &dto.SPMBRegisterResponseDTO{
		ParticipantCode: kodepeserta,
		IDPendaftar:     req.IDPendaftar,
		ExamID:          exam.IDUjian,
		ExamName:        exam.NamaUjian,
		PeriodID:        exam.IDPeriode,
		PeriodName:      exam.NamaPeriode,
		CBTCredentials: dto.SPMBCredentialsDTO{
			Username: kodepeserta,
			Password: plainPassword,
		},
		Schedule: dto.SPMBScheduleInfoDTO{
			IsPlotted: false,
		},
	}

	// 6. Auto-plotting into available session & room if requested
	if req.AutoPlotSession {
		var availableSession struct {
			IDJadwalUjian int            `db:"idjadwalujian"`
			IDRuangUjian  int            `db:"idruangujian"`
			NamaRuang     sql.NullString `db:"namaruang"`
			TglUjian      sql.NullTime   `db:"tglujian"`
			JamMulai      sql.NullString `db:"jammulai"`
			JamSelesai    sql.NullString `db:"jamselesai"`
			Durasi        int            `db:"durasi"`
		}

		// Find first session & room with remaining capacity
		findQuery := `
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
			ORDER BY COALESCE(r.tglmulai, j.tglmulai) ASC, j.idjadwalujian ASC
			LIMIT 1
		`
		err = r.db.GetContext(ctx, &availableSession, findQuery, exam.IDUjian)
		if err == nil && availableSession.IDJadwalUjian > 0 {
			// Plot into cat.at_jadwalpeserta
			_, _ = r.db.ExecContext(ctx, `
				INSERT INTO cat.at_jadwalpeserta (
					kodepeserta, idjadwalujian, idruangujian, is_locked, risk_level,
					softdelete, t_updatetime, t_updateact
				) VALUES (
					$1, $2, $3, 0, 'NORMAL', '0', NOW(), 'i-spmb-plot'
				)
				ON CONFLICT (kodepeserta, idjadwalujian) DO UPDATE
				SET idruangujian = $3, softdelete = '0', t_updatetime = NOW(), t_updateact = 'u-spmb-plot'
			`, kodepeserta, availableSession.IDJadwalUjian, availableSession.IDRuangUjian)

			resp.Schedule.IsPlotted = true
			resp.Schedule.SessionID = availableSession.IDJadwalUjian
			if availableSession.NamaRuang.Valid {
				resp.Schedule.RoomName = availableSession.NamaRuang.String
			}
			if availableSession.TglUjian.Valid {
				resp.Schedule.ExamDate = availableSession.TglUjian.Time.Format("2006-01-02")
			}
			if availableSession.JamMulai.Valid {
				resp.Schedule.StartTime = availableSession.JamMulai.String
			}
			if availableSession.JamSelesai.Valid {
				resp.Schedule.EndTime = availableSession.JamSelesai.String
			}
			resp.Schedule.DurationMinutes = availableSession.Durasi
		}
	}

	return resp, nil
}

func (r *integrationRepository) CreateAccessLog(ctx context.Context, log *entity.APIAccessLog) error {
	query := `
		INSERT INTO cat.at_api_access_logs (
			api_key_id, client_name, ip_address, method, endpoint,
			status_code, response_time_ms, user_agent, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		log.APIKeyID, log.ClientName, log.IPAddress, log.Method, log.Endpoint,
		log.StatusCode, log.ResponseTimeMS, log.UserAgent, log.ErrorMessage,
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
		SELECT id, api_key_id, client_name, ip_address, method, endpoint, status_code, response_time_ms, user_agent, error_message, created_at
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

