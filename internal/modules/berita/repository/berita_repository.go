package repository

import (
	"context"
	"poltekkes-cat-backend/internal/modules/berita/entity"

	"github.com/jmoiron/sqlx"
)

type BeritaRepository interface {
	GetAll(ctx context.Context) ([]*entity.Berita, error)
	Create(ctx context.Context, b *entity.Berita) (int, error)
}

type beritaRepository struct {
	db *sqlx.DB
}

func NewBeritaRepository(db *sqlx.DB) BeritaRepository {
	return &beritaRepository{db: db}
}

func (r *beritaRepository) GetAll(ctx context.Context) ([]*entity.Berita, error) {
	var list []*entity.Berita
	query := `
		SELECT idberita, judulberita, isiberita, extimage, khusus, t_inserttime
		FROM cat.at_berita
		WHERE softdelete = '0' OR softdelete IS NULL
		ORDER BY idberita DESC LIMIT 20
	`
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *beritaRepository) Create(ctx context.Context, b *entity.Berita) (int, error) {
	query := `
		INSERT INTO cat.at_berita (judulberita, isiberita, khusus, t_inserttime, softdelete)
		VALUES ($1, $2, $3, NOW(), '0')
		RETURNING idberita
	`
	var id int
	err := r.db.QueryRowContext(ctx, query, b.JudulBerita, b.IsiBerita, b.Khusus).Scan(&id)
	return id, err
}
