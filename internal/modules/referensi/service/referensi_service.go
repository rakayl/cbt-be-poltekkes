package service

import (
	"context"
	"poltekkes-cat-backend/internal/modules/referensi/dto"
	"poltekkes-cat-backend/internal/modules/referensi/repository"
)

type ReferensiService interface {
	GetRuangList(ctx context.Context) ([]*dto.RuangDTO, error)
	CreateRuang(ctx context.Context, req *dto.CreateRuangRequestDTO) (string, error)
	UpdateRuang(ctx context.Context, kodeRuang string, req *dto.UpdateRuangRequestDTO) error
	DeleteRuang(ctx context.Context, kodeRuang string) error

	GetSkorList(ctx context.Context) ([]*dto.SkorDTO, error)

	GetJenisUjianList(ctx context.Context) ([]*dto.JenisUjianDTO, error)
	CreateJenisUjian(ctx context.Context, req *dto.CreateJenisUjianRequestDTO) (string, error)
	UpdateJenisUjian(ctx context.Context, kodeJenis string, req *dto.UpdateJenisUjianRequestDTO) error
	DeleteJenisUjian(ctx context.Context, kodeJenis string) error

	GetWilayahList(ctx context.Context, parentID string) ([]*dto.WilayahDTO, error)

	GetPanitiaList(ctx context.Context) ([]*dto.PanitiaDTO, error)
	CreatePanitia(ctx context.Context, req *dto.CreatePanitiaRequestDTO) error
	UpdatePanitia(ctx context.Context, nip string, req *dto.UpdatePanitiaRequestDTO) error
	DeletePanitia(ctx context.Context, nip string) error

	GetJenisPeriodeList(ctx context.Context) ([]*dto.JenisPeriodeDTO, error)
	CreateJenisPeriode(ctx context.Context, req *dto.CreateJenisPeriodeRequestDTO) error
	UpdateJenisPeriode(ctx context.Context, jenisPeriode string, req *dto.UpdateJenisPeriodeRequestDTO) error
	DeleteJenisPeriode(ctx context.Context, jenisPeriode string) error

	GetUnitList(ctx context.Context) ([]*dto.UnitDTO, error)
	SalinSoal(ctx context.Context, req *dto.SalinSoalRequestDTO) (*dto.SalinSoalResponseDTO, error)
}

type referensiService struct {
	repo repository.ReferensiRepository
}

func NewReferensiService(repo repository.ReferensiRepository) ReferensiService {
	return &referensiService{repo: repo}
}

func (s *referensiService) GetRuangList(ctx context.Context) ([]*dto.RuangDTO, error) {
	list, err := s.repo.GetRuangList(ctx)
	if err != nil {
		return nil, err
	}
	var res []*dto.RuangDTO
	for _, item := range list {
		res = append(res, &dto.RuangDTO{
			KodeRuang:  item.KodeRuang,
			NamaRuang:  item.NamaRuang,
			Kapasitas:  item.Kapasitas,
			Keterangan: item.Keterangan,
		})
	}
	return res, nil
}

func (s *referensiService) GetSkorList(ctx context.Context) ([]*dto.SkorDTO, error) {
	list, err := s.repo.GetSkorList(ctx)
	if err != nil {
		return nil, err
	}
	var res []*dto.SkorDTO
	for _, item := range list {
		res = append(res, &dto.SkorDTO{
			KodeSkor:  item.KodeSkor,
			SkorBenar: item.SkorBenar,
			SkorSalah: item.SkorSalah,
		})
	}
	return res, nil
}

func (s *referensiService) GetJenisUjianList(ctx context.Context) ([]*dto.JenisUjianDTO, error) {
	list, err := s.repo.GetJenisUjianList(ctx)
	if err != nil {
		return nil, err
	}
	var res []*dto.JenisUjianDTO
	for _, item := range list {
		res = append(res, &dto.JenisUjianDTO{
			KodeJenis: item.KodeJenis,
			NamaJenis: item.NamaJenis,
		})
	}
	return res, nil
}

func (s *referensiService) GetWilayahList(ctx context.Context, parentID string) ([]*dto.WilayahDTO, error) {
	list, err := s.repo.GetWilayahList(ctx, parentID)
	if err != nil {
		return nil, err
	}
	var res []*dto.WilayahDTO
	for _, item := range list {
		res = append(res, &dto.WilayahDTO{
			IDWilayah:     item.IDWilayah,
			ParentWilayah: item.ParentWilayah,
			NamaWilayah:   item.NamaWilayah,
			Level:         item.Level,
		})
	}
	return res, nil
}

func (s *referensiService) GetPanitiaList(ctx context.Context) ([]*dto.PanitiaDTO, error) {
	list, err := s.repo.GetPanitiaList(ctx)
	if err != nil {
		return nil, err
	}
	var res []*dto.PanitiaDTO
	for _, item := range list {
		res = append(res, &dto.PanitiaDTO{
			NIP:         item.NIP,
			NamaPanitia: item.NamaPanitia,
		})
	}
	return res, nil
}

func (s *referensiService) GetJenisPeriodeList(ctx context.Context) ([]*dto.JenisPeriodeDTO, error) {
	list, err := s.repo.GetJenisPeriodeList(ctx)
	if err != nil {
		return nil, err
	}
	var res []*dto.JenisPeriodeDTO
	for _, item := range list {
		res = append(res, &dto.JenisPeriodeDTO{
			JenisPeriode:     item.JenisPeriode,
			KodeJenisPeriode: item.JenisPeriode,
			NamaJenisPeriode: item.NamaJenisPeriode,
		})
	}
	return res, nil
}

func (s *referensiService) GetUnitList(ctx context.Context) ([]*dto.UnitDTO, error) {
	list, err := s.repo.GetUnitList(ctx)
	if err != nil {
		return nil, err
	}
	var res []*dto.UnitDTO
	for _, item := range list {
		res = append(res, &dto.UnitDTO{
			IDSatker:    item.IDSatker,
			NamaSatker:  item.NamaSatker,
			NamaSingkat: item.NamaSingkat,
		})
	}
	return res, nil
}

func (s *referensiService) SalinSoal(ctx context.Context, req *dto.SalinSoalRequestDTO) (*dto.SalinSoalResponseDTO, error) {
	total, err := s.repo.SalinSoal(ctx, req.SumberKodeSoal, req.TargetKodeSoal, req.TargetNamaSoal)
	if err != nil {
		return nil, err
	}
	return &dto.SalinSoalResponseDTO{
		SumberKodeSoal: req.SumberKodeSoal,
		TargetKodeSoal: req.TargetKodeSoal,
		TotalDisalin:   total,
	}, nil
}

func (s *referensiService) CreateRuang(ctx context.Context, req *dto.CreateRuangRequestDTO) (string, error) {
	return s.repo.CreateRuang(ctx, req)
}

func (s *referensiService) UpdateRuang(ctx context.Context, kodeRuang string, req *dto.UpdateRuangRequestDTO) error {
	return s.repo.UpdateRuang(ctx, kodeRuang, req)
}

func (s *referensiService) DeleteRuang(ctx context.Context, kodeRuang string) error {
	return s.repo.DeleteRuang(ctx, kodeRuang)
}

func (s *referensiService) CreateJenisUjian(ctx context.Context, req *dto.CreateJenisUjianRequestDTO) (string, error) {
	return s.repo.CreateJenisUjian(ctx, req)
}

func (s *referensiService) UpdateJenisUjian(ctx context.Context, kodeJenis string, req *dto.UpdateJenisUjianRequestDTO) error {
	return s.repo.UpdateJenisUjian(ctx, kodeJenis, req)
}

func (s *referensiService) DeleteJenisUjian(ctx context.Context, kodeJenis string) error {
	return s.repo.DeleteJenisUjian(ctx, kodeJenis)
}

func (s *referensiService) CreatePanitia(ctx context.Context, req *dto.CreatePanitiaRequestDTO) error {
	return s.repo.CreatePanitia(ctx, req)
}

func (s *referensiService) UpdatePanitia(ctx context.Context, nip string, req *dto.UpdatePanitiaRequestDTO) error {
	return s.repo.UpdatePanitia(ctx, nip, req)
}

func (s *referensiService) DeletePanitia(ctx context.Context, nip string) error {
	return s.repo.DeletePanitia(ctx, nip)
}

func (s *referensiService) CreateJenisPeriode(ctx context.Context, req *dto.CreateJenisPeriodeRequestDTO) error {
	return s.repo.CreateJenisPeriode(ctx, req)
}

func (s *referensiService) UpdateJenisPeriode(ctx context.Context, jenisPeriode string, req *dto.UpdateJenisPeriodeRequestDTO) error {
	return s.repo.UpdateJenisPeriode(ctx, jenisPeriode, req)
}

func (s *referensiService) DeleteJenisPeriode(ctx context.Context, jenisPeriode string) error {
	return s.repo.DeleteJenisPeriode(ctx, jenisPeriode)
}

