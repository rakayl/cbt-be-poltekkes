package service

import (
	"context"
	"poltekkes-cat-backend/internal/modules/periode/dto"
	"poltekkes-cat-backend/internal/modules/periode/entity"
	"poltekkes-cat-backend/internal/modules/periode/repository"
	"poltekkes-cat-backend/internal/shared/response"
)

type PeriodeService interface {
	GetPeriods(ctx context.Context, page, perPage int, search string) ([]*dto.PeriodResponseDTO, *response.Pagination, error)
	CreatePeriod(ctx context.Context, req *dto.CreatePeriodRequestDTO) (int, error)
}

type periodeService struct {
	repo repository.PeriodeRepository
}

func NewPeriodeService(repo repository.PeriodeRepository) PeriodeService {
	return &periodeService{repo: repo}
}

func (s *periodeService) GetPeriods(ctx context.Context, page, perPage int, search string) ([]*dto.PeriodResponseDTO, *response.Pagination, error) {
	entities, total, err := s.repo.GetPeriods(ctx, page, perPage, search)
	if err != nil {
		return nil, nil, err
	}

	var results []*dto.PeriodResponseDTO
	for _, p := range entities {
		results = append(results, &dto.PeriodResponseDTO{
			PeriodID:    p.IDPeriode,
			IDPeriode:   p.IDPeriode,
			PeriodName:  p.NamaPeriode,
			NamaPeriode: p.NamaPeriode,
			PeriodType:  p.JenisPeriode,
			IsOnline:    p.IsOnline,
		})
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	pagination := &response.Pagination{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalPages:   totalPages,
		TotalRecords: total,
	}

	return results, pagination, nil
}

func (s *periodeService) CreatePeriod(ctx context.Context, req *dto.CreatePeriodRequestDTO) (int, error) {
	entityData := &entity.Periode{
		NamaPeriode:  req.PeriodName,
		JenisPeriode: req.PeriodType,
		IsOnline:     req.IsOnline,
	}
	if entityData.JenisPeriode == "" {
		entityData.JenisPeriode = "S"
	}
	return s.repo.CreatePeriod(ctx, entityData)
}
