package service

import (
	"context"
	"fmt"
	"strings"

	"poltekkes-cat-backend/internal/modules/bank_soal/dto"
	"poltekkes-cat-backend/internal/modules/bank_soal/entity"
	"poltekkes-cat-backend/internal/modules/bank_soal/repository"
)

type DynamicBankSoalService interface {
	GetStimuliByBankCode(ctx context.Context, bankCode string) ([]*entity.BankStimulus, error)
	GetStimulusByID(ctx context.Context, id int64) (*entity.BankStimulus, error)
	CreateStimulus(ctx context.Context, req *dto.CreateStimulusDTO) (*entity.BankStimulus, error)
	UpdateStimulus(ctx context.Context, id int64, req *dto.UpdateStimulusDTO) error
	DeleteStimulus(ctx context.Context, id int64) error

	GetBlueprints(ctx context.Context) ([]*entity.ExamBlueprint, error)
	GetBlueprintByID(ctx context.Context, id int64) (*entity.ExamBlueprint, error)
	CreateBlueprint(ctx context.Context, req *dto.CreateBlueprintDTO) (*entity.ExamBlueprint, error)

	GetExamScheduleExt(ctx context.Context, scheduleID int) (*entity.ExamScheduleExt, error)
	SetExamScheduleExt(ctx context.Context, req *dto.SetExamScheduleExtDTO) error
}

type dynamicBankSoalService struct {
	repo repository.DynamicBankSoalRepository
}

func NewDynamicBankSoalService(repo repository.DynamicBankSoalRepository) DynamicBankSoalService {
	return &dynamicBankSoalService{repo: repo}
}

func (s *dynamicBankSoalService) GetStimuliByBankCode(ctx context.Context, bankCode string) ([]*entity.BankStimulus, error) {
	return s.repo.GetStimuliByBankCode(ctx, bankCode)
}

func (s *dynamicBankSoalService) GetStimulusByID(ctx context.Context, id int64) (*entity.BankStimulus, error) {
	return s.repo.GetStimulusByID(ctx, id)
}

func (s *dynamicBankSoalService) CreateStimulus(ctx context.Context, req *dto.CreateStimulusDTO) (*entity.BankStimulus, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, fmt.Errorf("judul stimulus/kasus wajib diisi")
	}
	if strings.TrimSpace(req.NarrativeText) == "" {
		return nil, fmt.Errorf("naskah kasus/stimulus wajib diisi")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("stimulus harus memiliki minimal 1 butir pertanyaan")
	}

	for i, item := range req.Items {
		if strings.TrimSpace(item.QuestionText) == "" {
			return nil, fmt.Errorf("teks pertanyaan nomor %d wajib diisi", i+1)
		}
		if len(item.Options) < 2 {
			return nil, fmt.Errorf("pertanyaan nomor %d harus memiliki minimal 2 opsi pilihan", i+1)
		}
		if strings.TrimSpace(item.CorrectAnswer) == "" {
			return nil, fmt.Errorf("kunci jawaban untuk pertanyaan nomor %d wajib ditentukan", i+1)
		}
	}

	return s.repo.CreateStimulus(ctx, req)
}

func (s *dynamicBankSoalService) UpdateStimulus(ctx context.Context, id int64, req *dto.UpdateStimulusDTO) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("judul stimulus/kasus wajib diisi")
	}
	if strings.TrimSpace(req.NarrativeText) == "" {
		return fmt.Errorf("naskah kasus/stimulus wajib diisi")
	}
	return s.repo.UpdateStimulus(ctx, id, req)
}

func (s *dynamicBankSoalService) DeleteStimulus(ctx context.Context, id int64) error {
	return s.repo.DeleteStimulus(ctx, id)
}

func (s *dynamicBankSoalService) GetBlueprints(ctx context.Context) ([]*entity.ExamBlueprint, error) {
	return s.repo.GetBlueprints(ctx)
}

func (s *dynamicBankSoalService) GetBlueprintByID(ctx context.Context, id int64) (*entity.ExamBlueprint, error) {
	return s.repo.GetBlueprintByID(ctx, id)
}

func (s *dynamicBankSoalService) CreateBlueprint(ctx context.Context, req *dto.CreateBlueprintDTO) (*entity.ExamBlueprint, error) {
	if strings.TrimSpace(req.BlueprintCode) == "" {
		return nil, fmt.Errorf("kode blueprint wajib diisi")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, fmt.Errorf("nama blueprint wajib diisi")
	}
	if len(req.Rules) == 0 {
		return nil, fmt.Errorf("blueprint harus memiliki minimal 1 aturan kuota tag")
	}
	return s.repo.CreateBlueprint(ctx, req)
}

func (s *dynamicBankSoalService) GetExamScheduleExt(ctx context.Context, scheduleID int) (*entity.ExamScheduleExt, error) {
	return s.repo.GetExamScheduleExt(ctx, scheduleID)
}

func (s *dynamicBankSoalService) SetExamScheduleExt(ctx context.Context, req *dto.SetExamScheduleExtDTO) error {
	if req.IDJadwalUjian <= 0 {
		return fmt.Errorf("id jadwal ujian tidak valid")
	}
	return s.repo.SetExamScheduleExt(ctx, req)
}
