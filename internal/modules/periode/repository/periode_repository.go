package repository

import (
	"context"
	"poltekkes-cat-backend/internal/modules/periode/entity"

	"github.com/jmoiron/sqlx"
)

type PeriodeRepository interface {
	GetPeriods(ctx context.Context, page, perPage int, search string) ([]*entity.Periode, int64, error)
	CreatePeriod(ctx context.Context, p *entity.Periode) (int, error)
	UpdatePeriod(ctx context.Context, id int, p *entity.Periode) error
	DeletePeriod(ctx context.Context, id int) error
}

type periodeRepository struct {
	db *sqlx.DB
}

func NewPeriodeRepository(db *sqlx.DB) PeriodeRepository {
	return &periodeRepository{db: db}
}

func (r *periodeRepository) GetPeriods(ctx context.Context, page, perPage int, search string) ([]*entity.Periode, int64, error) {
	offset := (page - 1) * perPage
	var total int64

	countQuery := `SELECT COUNT(*) FROM cat.at_periode WHERE (softdelete = '0' OR softdelete IS NULL)`
	if search != "" {
		countQuery += ` AND namaperiode ILIKE '%` + search + `%'`
	}
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT idperiode, COALESCE(namaperiode, '') as namaperiode, 
		       COALESCE(jenisperiode, '1') as jenisperiode, COALESCE(isonline, 0) as isonline
		FROM cat.at_periode
		WHERE (softdelete = '0' OR softdelete IS NULL)
	`
	if search != "" {
		query += ` AND namaperiode ILIKE '%` + search + `%'`
	}
	query += ` ORDER BY idperiode DESC LIMIT $1 OFFSET $2`

	var periods []*entity.Periode
	err = r.db.SelectContext(ctx, &periods, query, perPage, offset)
	return periods, total, err
}

func (r *periodeRepository) CreatePeriod(ctx context.Context, p *entity.Periode) (int, error) {
	query := `
		INSERT INTO cat.at_periode (namaperiode, jenisperiode, isonline, softdelete)
		VALUES ($1, $2, $3, '0')
		RETURNING idperiode
	`
	var id int
	err := r.db.QueryRowContext(ctx, query, p.NamaPeriode, p.JenisPeriode, p.IsOnline).Scan(&id)
	return id, err
}

func (r *periodeRepository) UpdatePeriod(ctx context.Context, id int, p *entity.Periode) error {
	query := `
		UPDATE cat.at_periode 
		SET namaperiode = $1, jenisperiode = $2, isonline = $3
		WHERE idperiode = $4
	`
	_, err := r.db.ExecContext(ctx, query, p.NamaPeriode, p.JenisPeriode, p.IsOnline, id)
	return err
}

func (r *periodeRepository) DeletePeriod(ctx context.Context, id int) error {
	query := `
		UPDATE cat.at_periode 
		SET softdelete = '1'
		WHERE idperiode = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
