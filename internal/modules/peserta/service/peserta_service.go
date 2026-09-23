package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"poltekkes-cat-backend/internal/modules/peserta/dto"
	"poltekkes-cat-backend/internal/modules/peserta/entity"
	"poltekkes-cat-backend/internal/modules/peserta/repository"
	"poltekkes-cat-backend/internal/shared/response"
)

type PesertaService interface {
	GetParticipants(ctx context.Context, periodID, page, perPage int, search string) ([]*dto.PesertaResponseDTO, *response.Pagination, error)
	GetParticipantByCode(ctx context.Context, code string) (*dto.PesertaResponseDTO, error)
	CreateParticipant(ctx context.Context, req *dto.SavePesertaRequestDTO) error
	UpdateParticipant(ctx context.Context, code string, req *dto.SavePesertaRequestDTO) error
	DeleteParticipant(ctx context.Context, code string) error
	ResetLogin(ctx context.Context, req *dto.ResetParticipantLoginDTO) error
	SearchParticipants(ctx context.Context, query string) ([]*dto.PesertaResponseDTO, error)
	ChangePassword(ctx context.Context, code string, req *dto.ChangePasswordRequestDTO) error

	// Sipenmaru Integration
	GetSipenmaruFilterOptions(ctx context.Context) (*dto.SipenmaruFilterOptionsDTO, error)
	PreviewSipenmaruCandidates(ctx context.Context, req *dto.SipenmaruPreviewRequestDTO) (*dto.SipenmaruPreviewResponseDTO, error)
	ExecuteImportSipenmaru(ctx context.Context, req *dto.ImportSipenmaruRequestDTO) (*dto.ImportSipenmaruResponseDTO, error)
}

type pesertaService struct {
	repo repository.PesertaRepository
}

func NewPesertaService(repo repository.PesertaRepository) PesertaService {
	return &pesertaService{repo: repo}
}

func (s *pesertaService) GetParticipants(ctx context.Context, periodID, page, perPage int, search string) ([]*dto.PesertaResponseDTO, *response.Pagination, error) {
	entities, total, err := s.repo.GetParticipants(ctx, periodID, page, perPage, search)
	if err != nil {
		return nil, nil, err
	}

	var results []*dto.PesertaResponseDTO
	for _, p := range entities {
		results = append(results, s.mapEntityToDTO(p))
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

func (s *pesertaService) GetParticipantByCode(ctx context.Context, code string) (*dto.PesertaResponseDTO, error) {
	p, err := s.repo.GetParticipantByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.mapEntityToDTO(p), nil
}

func (s *pesertaService) CreateParticipant(ctx context.Context, req *dto.SavePesertaRequestDTO) error {
	rawPass := req.KodePeserta
	if req.Password != nil && *req.Password != "" {
		rawPass = *req.Password
	} else if req.Hint != nil && *req.Hint != "" {
		rawPass = *req.Hint
	}

	hasher := md5.New()
	hasher.Write([]byte(rawPass))
	hashedPass := hex.EncodeToString(hasher.Sum(nil))

	hintVal := rawPass
	if req.Hint != nil && *req.Hint != "" {
		hintVal = *req.Hint
	}

	p := &entity.Peserta{
		KodePeserta:    req.KodePeserta,
		IDPeriode:      req.IDPeriode,
		Nama:           req.Nama,
		JK:             req.JK,
		HP:             req.HP,
		Email:          req.Email,
		SumberData:     req.SumberData,
		Alamat:         req.Alamat,
		IDKota:         req.IDKota,
		IDKotaLama:     req.NamaWilayah,
		Profesi:        req.Profesi,
		Status:         req.Status,
		FilePembayaran: req.FilePembayaran,
		Hint:           &hintVal,
		Password:       &hashedPass,
		IsValid:        req.IsValid,
		IsAktif:        req.IsAktif,
		Keterangan:     req.Keterangan,
		KodeReferensi:  req.KodeReferensi,
	}

	return s.repo.CreateParticipant(ctx, p)
}

func (s *pesertaService) UpdateParticipant(ctx context.Context, code string, req *dto.SavePesertaRequestDTO) error {
	var hashedPass *string
	var hintVal *string

	if (req.Password != nil && *req.Password != "") || (req.Hint != nil && *req.Hint != "") {
		passToUse := ""
		if req.Password != nil && *req.Password != "" {
			passToUse = *req.Password
		} else {
			passToUse = *req.Hint
		}
		hasher := md5.New()
		hasher.Write([]byte(passToUse))
		h := hex.EncodeToString(hasher.Sum(nil))
		hashedPass = &h
		hintVal = &passToUse
	}

	p := &entity.Peserta{
		KodePeserta:    code,
		IDPeriode:      req.IDPeriode,
		Nama:           req.Nama,
		JK:             req.JK,
		HP:             req.HP,
		Email:          req.Email,
		SumberData:     req.SumberData,
		Alamat:         req.Alamat,
		IDKota:         req.IDKota,
		IDKotaLama:     req.NamaWilayah,
		Profesi:        req.Profesi,
		Status:         req.Status,
		FilePembayaran: req.FilePembayaran,
		Hint:           hintVal,
		Password:       hashedPass,
		IsValid:        req.IsValid,
		IsAktif:        req.IsAktif,
		Keterangan:     req.Keterangan,
		KodeReferensi:  req.KodeReferensi,
	}

	return s.repo.UpdateParticipant(ctx, code, p)
}

func (s *pesertaService) DeleteParticipant(ctx context.Context, code string) error {
	return s.repo.DeleteParticipant(ctx, code)
}

func (s *pesertaService) ResetLogin(ctx context.Context, req *dto.ResetParticipantLoginDTO) error {
	code := req.KodePeserta
	if code == "" {
		code = req.Code
	}
	return s.repo.ResetLogin(ctx, code)
}

func (s *pesertaService) SearchParticipants(ctx context.Context, query string) ([]*dto.PesertaResponseDTO, error) {
	entities, err := s.repo.SearchParticipants(ctx, query)
	if err != nil {
		return nil, err
	}
	var results []*dto.PesertaResponseDTO
	for _, p := range entities {
		results = append(results, s.mapEntityToDTO(p))
	}
	return results, nil
}

func (s *pesertaService) ChangePassword(ctx context.Context, code string, req *dto.ChangePasswordRequestDTO) error {
	hasher := md5.New()
	hasher.Write([]byte(req.Password))
	hashedPass := hex.EncodeToString(hasher.Sum(nil))

	return s.repo.ChangePassword(ctx, code, hashedPass, req.Password)
}

func (s *pesertaService) mapEntityToDTO(p *entity.Peserta) *dto.PesertaResponseDTO {
	return &dto.PesertaResponseDTO{
		KodePeserta:     p.KodePeserta,
		ParticipantCode: p.KodePeserta,
		IDPeriode:       p.IDPeriode,
		NamaPeriode:     p.NamaPeriode,
		Nama:            p.Nama,
		Name:            p.Nama,
		JK:              p.JK,
		Email:           p.Email,
		HP:              p.HP,
		Phone:           p.HP,
		Alamat:          p.Alamat,
		Address:         p.Alamat,
		IDKota:          p.IDKota,
		NamaWilayah:     p.NamaWilayah,
		Profesi:         p.Profesi,
		Status:          p.Status,
		SumberData:      p.SumberData,
		KodeReferensi:   p.KodeReferensi,
		Hint:            p.Hint,
		IsValid:         p.IsValid,
		IsAktif:         p.IsAktif,
		IsActive:        p.IsAktif,
		IsLogin:         p.IsLogin,
		FilePembayaran:  p.FilePembayaran,
		Keterangan:      p.Keterangan,
	}
}

func (s *pesertaService) GetSipenmaruFilterOptions(ctx context.Context) (*dto.SipenmaruFilterOptionsDTO, error) {
	return s.repo.GetSipenmaruFilterOptions(ctx)
}

func (s *pesertaService) PreviewSipenmaruCandidates(ctx context.Context, req *dto.SipenmaruPreviewRequestDTO) (*dto.SipenmaruPreviewResponseDTO, error) {
	candidates, err := s.repo.GetSipenmaruCandidates(ctx, req)
	if err != nil {
		return nil, err
	}

	resp := &dto.SipenmaruPreviewResponseDTO{
		TotalCount: len(candidates),
		Candidates: candidates,
	}

	for _, c := range candidates {
		if c.IsAdministrasi == 1 {
			resp.EligibleCount++
		}
		if c.IsImported {
			resp.ImportedCount++
		} else {
			if !req.OnlyAdministrasi || c.IsAdministrasi == 1 {
				resp.ReadyCount++
			}
		}
	}

	return resp, nil
}

func (s *pesertaService) ExecuteImportSipenmaru(ctx context.Context, req *dto.ImportSipenmaruRequestDTO) (*dto.ImportSipenmaruResponseDTO, error) {
	return s.repo.ImportSipenmaruCandidates(ctx, req)
}

