package service

import (
	"context"
	"fmt"
	"poltekkes-cat-backend/internal/modules/bank_soal/dto"
	"poltekkes-cat-backend/internal/modules/bank_soal/repository"
	"poltekkes-cat-backend/internal/shared/response"
)

type BankSoalService interface {
	GetQuestionBanks(ctx context.Context, page, perPage int, search string) ([]*dto.QuestionBankResponseDTO, *response.Pagination, error)
	GetQuestionBankByCode(ctx context.Context, code string) (*dto.QuestionBankResponseDTO, error)
	GetQuestionsByCode(ctx context.Context, code string) ([]*dto.QuestionItemResponseDTO, error)
	CreateQuestionBank(ctx context.Context, req *dto.CreateQuestionBankDTO) (*dto.QuestionBankResponseDTO, error)
	UpdateQuestionBank(ctx context.Context, code string, req *dto.UpdateQuestionBankDTO) error
	DeleteQuestionBank(ctx context.Context, code string) error
	CopyQuestionBank(ctx context.Context, sourceCode string, req *dto.CopyQuestionBankDTO) error
	CreateQuestionItem(ctx context.Context, code string, req *dto.SaveQuestionItemDTO) (*dto.QuestionItemResponseDTO, error)
	UpdateQuestionItem(ctx context.Context, code string, noUrut int, req *dto.SaveQuestionItemDTO) error
	DeleteQuestionItem(ctx context.Context, code string, noUrut int) error
}

type bankSoalService struct {
	repo repository.BankSoalRepository
}

func NewBankSoalService(repo repository.BankSoalRepository) BankSoalService {
	return &bankSoalService{repo: repo}
}

func (s *bankSoalService) GetQuestionBanks(ctx context.Context, page, perPage int, search string) ([]*dto.QuestionBankResponseDTO, *response.Pagination, error) {
	entities, total, err := s.repo.GetQuestionBanks(ctx, page, perPage, search)
	if err != nil {
		return nil, nil, err
	}

	var results []*dto.QuestionBankResponseDTO
	for _, b := range entities {
		desc := ""
		if b.Keterangan != nil {
			desc = *b.Keterangan
		}
		results = append(results, &dto.QuestionBankResponseDTO{
			QuestionBankCode:     b.KodeSoal,
			SubjectName:          b.NamaSoal,
			Description:          desc,
			TotalQuestions:       b.TotalQuestions,
			TotalQuestionsStatic: b.TotalQuestionsStatic,
			TotalStimuli:         b.TotalStimuli,
			CreatedAt:            b.TInsertTime,
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

func (s *bankSoalService) GetQuestionBankByCode(ctx context.Context, code string) (*dto.QuestionBankResponseDTO, error) {
	b, err := s.repo.GetQuestionBankByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	desc := ""
	if b.Keterangan != nil {
		desc = *b.Keterangan
	}

	return &dto.QuestionBankResponseDTO{
		QuestionBankCode:     b.KodeSoal,
		SubjectName:          b.NamaSoal,
		Description:          desc,
		TotalQuestions:       b.TotalQuestions,
		TotalQuestionsStatic: b.TotalQuestionsStatic,
		TotalStimuli:         b.TotalStimuli,
		CreatedAt:            b.TInsertTime,
	}, nil
}

func resolveMedia(img, aud, vid *string) (string, *string) {
	if img != nil && *img != "" {
		return "IMAGE", img
	}
	if aud != nil && *aud != "" {
		return "AUDIO", aud
	}
	if vid != nil && *vid != "" {
		return "VIDEO", vid
	}
	return "NONE", nil
}

func (s *bankSoalService) GetQuestionsByCode(ctx context.Context, code string) ([]*dto.QuestionItemResponseDTO, error) {
	entities, err := s.repo.GetQuestionsByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	var results []*dto.QuestionItemResponseDTO
	for _, q := range entities {
		qMediaType, qMediaURL := resolveMedia(q.PertanyaanImage, q.PertanyaanAudio, q.PertanyaanVideo)

		var options []dto.QuestionOptionDTO
		if q.Jawaban1 != nil && *q.Jawaban1 != "" {
			mType, mURL := resolveMedia(q.Jawaban1Image, q.Jawaban1Audio, q.Jawaban1Video)
			options = append(options, dto.QuestionOptionDTO{OptionKey: 1, OptionText: *q.Jawaban1, MediaType: mType, MediaURL: mURL})
		}
		if q.Jawaban2 != nil && *q.Jawaban2 != "" {
			mType, mURL := resolveMedia(q.Jawaban2Image, q.Jawaban2Audio, q.Jawaban2Video)
			options = append(options, dto.QuestionOptionDTO{OptionKey: 2, OptionText: *q.Jawaban2, MediaType: mType, MediaURL: mURL})
		}
		if q.Jawaban3 != nil && *q.Jawaban3 != "" {
			mType, mURL := resolveMedia(q.Jawaban3Image, q.Jawaban3Audio, q.Jawaban3Video)
			options = append(options, dto.QuestionOptionDTO{OptionKey: 3, OptionText: *q.Jawaban3, MediaType: mType, MediaURL: mURL})
		}
		if q.Jawaban4 != nil && *q.Jawaban4 != "" {
			mType, mURL := resolveMedia(q.Jawaban4Image, q.Jawaban4Audio, q.Jawaban4Video)
			options = append(options, dto.QuestionOptionDTO{OptionKey: 4, OptionText: *q.Jawaban4, MediaType: mType, MediaURL: mURL})
		}
		if q.Jawaban5 != nil && *q.Jawaban5 != "" {
			mType, mURL := resolveMedia(q.Jawaban5Image, q.Jawaban5Audio, q.Jawaban5Video)
			options = append(options, dto.QuestionOptionDTO{OptionKey: 5, OptionText: *q.Jawaban5, MediaType: mType, MediaURL: mURL})
		}

		catTag := "Umum"
		if q.KataKunci != nil && *q.KataKunci != "" {
			catTag = *q.KataKunci
		}

		results = append(results, &dto.QuestionItemResponseDTO{
			QuestionBankCode:  q.KodeSoal,
			QuestionNumber:    q.NoUrut,
			QuestionText:      q.Pertanyaan,
			QuestionMediaType: qMediaType,
			QuestionMediaURL:  qMediaURL,
			CorrectOption:     q.JawabanBenar,
			CategoryTag:       catTag,
			Bobot:             q.Bobot,
			BobotBenar:        q.BobotBenar,
			BobotSalah:        q.BobotSalah,
			Options:           options,
		})
	}

	return results, nil
}

func (s *bankSoalService) CreateQuestionBank(ctx context.Context, req *dto.CreateQuestionBankDTO) (*dto.QuestionBankResponseDTO, error) {
	err := s.repo.CreateQuestionBank(ctx, req.QuestionBankCode, req.QuestionBankName, req.Description)
	if err != nil {
		return nil, err
	}

	return &dto.QuestionBankResponseDTO{
		QuestionBankCode: req.QuestionBankCode,
		SubjectName:      req.QuestionBankName,
		Description:      req.Description,
		TotalQuestions:   0,
	}, nil
}

func (s *bankSoalService) UpdateQuestionBank(ctx context.Context, code string, req *dto.UpdateQuestionBankDTO) error {
	return s.repo.UpdateQuestionBank(ctx, code, req.QuestionBankName, req.Description)
}

func (s *bankSoalService) DeleteQuestionBank(ctx context.Context, code string) error {
	return s.repo.DeleteQuestionBank(ctx, code)
}

func (s *bankSoalService) CopyQuestionBank(ctx context.Context, sourceCode string, req *dto.CopyQuestionBankDTO) error {
	return s.repo.CopyQuestionBank(ctx, sourceCode, req.NewBankCode, req.NewBankName)
}

func (s *bankSoalService) CreateQuestionItem(ctx context.Context, code string, req *dto.SaveQuestionItemDTO) (*dto.QuestionItemResponseDTO, error) {
	q, err := s.repo.CreateQuestionItem(ctx, code, req)
	if err != nil {
		return nil, err
	}

	var options []dto.QuestionOptionDTO
	if q.Jawaban1 != nil {
		options = append(options, dto.QuestionOptionDTO{OptionKey: 1, OptionText: *q.Jawaban1})
	}
	if q.Jawaban2 != nil {
		options = append(options, dto.QuestionOptionDTO{OptionKey: 2, OptionText: *q.Jawaban2})
	}
	if q.Jawaban3 != nil {
		options = append(options, dto.QuestionOptionDTO{OptionKey: 3, OptionText: *q.Jawaban3})
	}
	if q.Jawaban4 != nil {
		options = append(options, dto.QuestionOptionDTO{OptionKey: 4, OptionText: *q.Jawaban4})
	}
	if q.Jawaban5 != nil {
		options = append(options, dto.QuestionOptionDTO{OptionKey: 5, OptionText: *q.Jawaban5})
	}

	catTag := "Umum"
	if q.KataKunci != nil {
		catTag = *q.KataKunci
	}

	return &dto.QuestionItemResponseDTO{
		QuestionBankCode: q.KodeSoal,
		QuestionNumber:   q.NoUrut,
		QuestionText:     q.Pertanyaan,
		CorrectOption:    q.JawabanBenar,
		CategoryTag:      catTag,
		Bobot:            q.Bobot,
		BobotBenar:       q.BobotBenar,
		BobotSalah:       q.BobotSalah,
		Options:          options,
	}, nil
}

func (s *bankSoalService) UpdateQuestionItem(ctx context.Context, code string, noUrut int, req *dto.SaveQuestionItemDTO) error {
	return s.repo.UpdateQuestionItem(ctx, code, noUrut, req)
}

func (s *bankSoalService) DeleteQuestionItem(ctx context.Context, code string, noUrut int) error {
	if code == "" || noUrut <= 0 {
		return fmt.Errorf("kode bank soal dan nomor urut butir tidak valid")
	}
	return s.repo.DeleteQuestionItem(ctx, code, noUrut)
}
