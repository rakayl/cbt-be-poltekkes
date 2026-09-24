package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"poltekkes-cat-backend/internal/modules/integration/dto"
	"poltekkes-cat-backend/internal/modules/integration/entity"
	"poltekkes-cat-backend/internal/modules/integration/repository"
	"poltekkes-cat-backend/internal/shared/response"
)

type IntegrationService interface {
	GetAPIKeys(ctx context.Context) ([]*dto.APIKeyResponseDTO, error)
	GetAPIKeyByID(ctx context.Context, id int) (*dto.APIKeyResponseDTO, error)
	CreateAPIKey(ctx context.Context, req *dto.CreateAPIKeyRequestDTO) (*dto.APIKeyResponseDTO, error)
	UpdateAPIKey(ctx context.Context, id int, req *dto.UpdateAPIKeyRequestDTO) (*dto.APIKeyResponseDTO, error)
	DeleteAPIKey(ctx context.Context, id int) error
	RegenerateAPIKey(ctx context.Context, id int) (*dto.APIKeyResponseDTO, error)

	GetActiveExams(ctx context.Context, refDate string) ([]*dto.SPMBActiveExamDTO, error)
	RegisterSPMBParticipant(ctx context.Context, req *dto.SPMBRegisterRequestDTO) (*dto.SPMBRegisterResponseDTO, error)

	GetAccessLogs(ctx context.Context, apiKeyID int, page, perPage int, search string) ([]*dto.APIAccessLogResponseDTO, *response.Pagination, error)
}

type integrationService struct {
	repo repository.IntegrationRepository
}

func NewIntegrationService(repo repository.IntegrationRepository) IntegrationService {
	return &integrationService{repo: repo}
}

func (s *integrationService) GenerateSecureAPIKey() string {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "cbt_live_" + hex.EncodeToString([]byte("fallback_key"))
	}
	return "cbt_live_" + hex.EncodeToString(bytes)
}

func (s *integrationService) toDTO(k *entity.APIKey) *dto.APIKeyResponseDTO {
	ip := ""
	if k.IPWhitelist != nil {
		ip = *k.IPWhitelist
	}
	desc := ""
	if k.Description != nil {
		desc = *k.Description
	}

	return &dto.APIKeyResponseDTO{
		ID:          k.ID,
		Name:        k.Name,
		APIKey:      k.APIKey,
		IPWhitelist: ip,
		IsActive:    k.IsActive,
		Description: desc,
		LastUsedAt:  k.LastUsedAt,
		CreatedAt:   k.CreatedAt,
		UpdatedAt:   k.UpdatedAt,
	}
}

func (s *integrationService) GetAPIKeys(ctx context.Context) ([]*dto.APIKeyResponseDTO, error) {
	keys, err := s.repo.GetAPIKeys(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*dto.APIKeyResponseDTO, 0, len(keys))
	for _, k := range keys {
		res = append(res, s.toDTO(k))
	}
	return res, nil
}

func (s *integrationService) GetAPIKeyByID(ctx context.Context, id int) (*dto.APIKeyResponseDTO, error) {
	k, err := s.repo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toDTO(k), nil
}

func (s *integrationService) CreateAPIKey(ctx context.Context, req *dto.CreateAPIKeyRequestDTO) (*dto.APIKeyResponseDTO, error) {
	keyStr := strings.TrimSpace(req.APIKey)
	if keyStr == "" {
		keyStr = s.GenerateSecureAPIKey()
	}

	var ip *string
	if strings.TrimSpace(req.IPWhitelist) != "" {
		cleaned := strings.TrimSpace(req.IPWhitelist)
		ip = &cleaned
	}

	var desc *string
	if strings.TrimSpace(req.Description) != "" {
		d := strings.TrimSpace(req.Description)
		desc = &d
	}

	newKey := &entity.APIKey{
		Name:        strings.TrimSpace(req.Name),
		APIKey:      keyStr,
		IPWhitelist: ip,
		IsActive:    true,
		Description: desc,
	}

	id, err := s.repo.CreateAPIKey(ctx, newKey)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat API Key: %w", err)
	}

	return s.GetAPIKeyByID(ctx, id)
}

func (s *integrationService) UpdateAPIKey(ctx context.Context, id int, req *dto.UpdateAPIKeyRequestDTO) (*dto.APIKeyResponseDTO, error) {
	existing, err := s.repo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("API Key #%d tidak ditemukan: %w", id, err)
	}

	existing.Name = strings.TrimSpace(req.Name)
	if strings.TrimSpace(req.IPWhitelist) != "" {
		cleaned := strings.TrimSpace(req.IPWhitelist)
		existing.IPWhitelist = &cleaned
	} else {
		existing.IPWhitelist = nil
	}

	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if strings.TrimSpace(req.Description) != "" {
		d := strings.TrimSpace(req.Description)
		existing.Description = &d
	} else {
		existing.Description = nil
	}

	if err := s.repo.UpdateAPIKey(ctx, existing); err != nil {
		return nil, fmt.Errorf("gagal memperbarui API Key: %w", err)
	}

	return s.GetAPIKeyByID(ctx, id)
}

func (s *integrationService) DeleteAPIKey(ctx context.Context, id int) error {
	return s.repo.DeleteAPIKey(ctx, id)
}

func (s *integrationService) RegenerateAPIKey(ctx context.Context, id int) (*dto.APIKeyResponseDTO, error) {
	_, err := s.repo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("API Key #%d tidak ditemukan: %w", id, err)
	}

	newKey := s.GenerateSecureAPIKey()
	if err := s.repo.RegenerateAPIKey(ctx, id, newKey); err != nil {
		return nil, fmt.Errorf("gagal men-generate ulang API Key: %w", err)
	}

	return s.GetAPIKeyByID(ctx, id)
}

// =========================================================================
// SPMB Integration
// =========================================================================

func (s *integrationService) GetActiveExams(ctx context.Context, refDate string) ([]*dto.SPMBActiveExamDTO, error) {
	return s.repo.GetSPMBActiveExams(ctx, refDate)
}

func (s *integrationService) RegisterSPMBParticipant(ctx context.Context, req *dto.SPMBRegisterRequestDTO) (*dto.SPMBRegisterResponseDTO, error) {
	// Clean and validate
	req.IDPendaftar = strings.TrimSpace(req.IDPendaftar)
	req.NomorUjian = strings.TrimSpace(req.NomorUjian)
	req.Nama = strings.TrimSpace(req.Nama)
	req.JK = strings.ToUpper(strings.TrimSpace(req.JK))
	req.Email = strings.TrimSpace(req.Email)
	if req.NomorUjian == "" && req.NoUjian != "" {
		req.NomorUjian = strings.TrimSpace(req.NoUjian)
	}

	if req.IDUjian <= 0 && req.ExamID <= 0 {
		return nil, fmt.Errorf("idujian wajib diisi")
	}
	if req.IDUjian <= 0 {
		req.IDUjian = req.ExamID
	}
	if req.ExamID <= 0 {
		req.ExamID = req.IDUjian
	}
	if req.IDPendaftar == "" {
		return nil, fmt.Errorf("idpendaftar wajib diisi")
	}
	if req.Nama == "" {
		return nil, fmt.Errorf("nama peserta wajib diisi")
	}

	plainPassword := strings.TrimSpace(req.Password)
	if plainPassword == "" {
		if req.NomorUjian != "" {
			plainPassword = req.NomorUjian
		} else {
			plainPassword = req.IDPendaftar
		}
	}

	// Default AutoPlotSession to true if not specified
	if req.AutoPlotSession == nil {
		defaultTrue := true
		req.AutoPlotSession = &defaultTrue
	}

	return s.repo.RegisterSPMBParticipant(ctx, req, plainPassword)
}

func (s *integrationService) GetAccessLogs(ctx context.Context, apiKeyID int, page, perPage int, search string) ([]*dto.APIAccessLogResponseDTO, *response.Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 15
	}

	logs, total, err := s.repo.GetAccessLogs(ctx, apiKeyID, page, perPage, search)
	if err != nil {
		return nil, nil, err
	}

	res := make([]*dto.APIAccessLogResponseDTO, 0, len(logs))
	for _, l := range logs {
		ua := ""
		if l.UserAgent != nil {
			ua = *l.UserAgent
		}
		errMsg := ""
		reqBody := ""
		if l.RequestBody != nil {
			reqBody = *l.RequestBody
		}
		respBody := ""
		if l.ResponseBody != nil {
			respBody = *l.ResponseBody
		}

		res = append(res, &dto.APIAccessLogResponseDTO{
			ID:             l.ID,
			APIKeyID:       l.APIKeyID,
			ClientName:     l.ClientName,
			IPAddress:      l.IPAddress,
			Method:         l.Method,
			Endpoint:       l.Endpoint,
			StatusCode:     l.StatusCode,
			ResponseTimeMS: l.ResponseTimeMS,
			UserAgent:      ua,
			ErrorMessage:   errMsg,
			RequestBody:    reqBody,
			ResponseBody:   respBody,
			CreatedAt:      l.CreatedAt,
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

	return res, pagination, nil
}
