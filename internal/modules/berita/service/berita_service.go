package service

import (
	"context"
	"poltekkes-cat-backend/internal/modules/berita/dto"
	"poltekkes-cat-backend/internal/modules/berita/entity"
	"poltekkes-cat-backend/internal/modules/berita/repository"
)

type BeritaService interface {
	GetAllBerita(ctx context.Context) ([]*dto.BeritaResponseDTO, error)
	CreateBerita(ctx context.Context, req *dto.CreateBeritaRequestDTO) (int, error)
}

type beritaService struct {
	repo repository.BeritaRepository
}

func NewBeritaService(repo repository.BeritaRepository) BeritaService {
	return &beritaService{repo: repo}
}

func (s *beritaService) GetAllBerita(ctx context.Context) ([]*dto.BeritaResponseDTO, error) {
	entities, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var results []*dto.BeritaResponseDTO
	for _, b := range entities {
		results = append(results, &dto.BeritaResponseDTO{
			IDBerita:    b.IDBerita,
			JudulBerita: b.JudulBerita,
			IsiBerita:   b.IsiBerita,
			WaktuRilis:  b.TInsertTime,
		})
	}
	return results, nil
}

func (s *beritaService) CreateBerita(ctx context.Context, req *dto.CreateBeritaRequestDTO) (int, error) {
	entityData := &entity.Berita{
		JudulBerita: req.JudulBerita,
		IsiBerita:   req.IsiBerita,
		Khusus:      &req.Khusus,
	}
	return s.repo.Create(ctx, entityData)
}
