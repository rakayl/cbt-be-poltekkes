package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/english_proficiency/dto"
	"poltekkes-cat-backend/internal/modules/english_proficiency/entity"
	"poltekkes-cat-backend/internal/modules/english_proficiency/repository"
	"poltekkes-cat-backend/internal/shared/response"
)

type EPService interface {
	GetExamTypes(ctx context.Context) ([]dto.ExamTypeDTO, error)
	GetConversionProfiles(ctx context.Context, examTypeID int) ([]dto.ConversionProfileDTO, error)
	GetConversionTable(ctx context.Context, profileID int) ([]dto.ScoreConversionDTO, error)
	SaveConversionProfile(ctx context.Context, req *dto.SaveConversionProfileRequestDTO) (*dto.ConversionProfileDTO, error)

	CreateExam(ctx context.Context, req *dto.CreateEPExamRequestDTO, createdBy string) (*dto.EPExamResponseDTO, error)
	GetExams(ctx context.Context, page, perPage int) ([]dto.EPExamResponseDTO, *response.Pagination, error)
	GetExamByID(ctx context.Context, examID int) (*dto.EPExamResponseDTO, error)
	UpdateExam(ctx context.Context, examID int, req *dto.CreateEPExamRequestDTO) (*dto.EPExamResponseDTO, error)
	DeleteExam(ctx context.Context, examID int) error

	CreateSchedule(ctx context.Context, req *dto.CreateEPScheduleRequestDTO) (*dto.EPScheduleResponseDTO, error)
	GetSchedules(ctx context.Context, examID int) ([]dto.EPScheduleResponseDTO, error)
	UpdateSchedule(ctx context.Context, scheduleID int, req *dto.UpdateEPScheduleRequestDTO) (*dto.EPScheduleResponseDTO, error)
	DeleteSchedule(ctx context.Context, scheduleID int) error
	RegisterParticipants(ctx context.Context, scheduleID int, codes []string, details []dto.RegisterParticipantItemDTO) (int, error)
	GetScheduleParticipants(ctx context.Context, scheduleID int) ([]*entity.ScheduleParticipant, error)

	StartExamSession(ctx context.Context, scheduleID int, participantCode, sessionToken string) (*dto.EPExamSessionResponseDTO, error)
	GetCurrentSectionSession(ctx context.Context, scheduleID int, participantCode string) (*dto.EPExamSessionResponseDTO, error)
	SaveAnswer(ctx context.Context, participantCode string, req *dto.SaveAnswerDTO) error
	AdvanceToNextSection(ctx context.Context, scheduleID int, participantCode string) (*dto.EPExamSessionResponseDTO, error)
	FinishExam(ctx context.Context, scheduleID int, participantCode, reason string) error
	RemoveParticipantFromSchedule(ctx context.Context, scheduleID int, kodepeserta string) error

	// Anti-Cheat & Proctor Telemetry
	RecordSecurityEvent(ctx context.Context, scheduleID int, participantCode string, req *dto.EPSecurityEventRequestDTO, ip, userAgent string) (*dto.EPSecurityEventResponseDTO, error)
	SyncParticipantHeartbeat(ctx context.Context, scheduleID int, participantCode string, req *dto.EPHeartbeatRequestDTO) (*dto.EPHeartbeatResponseDTO, error)
	ApplyProctorAction(ctx context.Context, req *dto.EPProctorActionRequestDTO) error
	GetSecurityEventsByParticipant(ctx context.Context, scheduleID int, participantCode string) ([]dto.EPSecurityAuditItemDTO, error)
	SaveProctorSnapshot(ctx context.Context, scheduleID int, participantCode string, req *dto.UploadProctorSnapshotRequestDTO) (*dto.ProctorSnapshotItemDTO, error)
	GetProctorSnapshots(ctx context.Context, scheduleID int, participantCode, snapshotType string, limit, offset int) ([]dto.ProctorSnapshotItemDTO, error)
	GetLatestProctorGrid(ctx context.Context, scheduleID int) ([]dto.ProctorLatestGridItemDTO, error)

	CalculateScores(ctx context.Context, scheduleID int) ([]dto.EPScoreDetailDTO, error)
	GetScores(ctx context.Context, scheduleID int) ([]dto.EPScoreDetailDTO, error)

	GenerateCertificates(ctx context.Context, scheduleID int) ([]*entity.EPCertificate, error)
	GetCertificate(ctx context.Context, certNumber string) (*entity.EPCertificate, error)
	VerifyCertificate(ctx context.Context, verificationCode string) (*dto.CertificateVerificationDTO, error)
	GetParticipantScores(ctx context.Context, participantCode string) ([]dto.ParticipantScoreResultDTO, error)

	// Question Banks
	GetEPQuestionBanks(ctx context.Context, examTypeID, sectionID int, search string, page, perPage int) ([]*entity.EPQuestionBank, *response.Pagination, error)
	GetEPQuestionBankByCode(ctx context.Context, kodesoal string) (*entity.EPQuestionBank, error)
	CreateEPQuestionBank(ctx context.Context, req *dto.CreateEPBankDTO) (*entity.EPQuestionBank, error)
	UpdateEPQuestionBank(ctx context.Context, kodesoal string, req *dto.UpdateEPBankDTO) error
	DeleteEPQuestionBank(ctx context.Context, kodesoal string) error
	CopyEPQuestionBank(ctx context.Context, req *dto.CopyEPBankDTO, sourceKode string) (*entity.EPQuestionBank, error)

	GetStimuliByBank(ctx context.Context, kodesoal string) ([]dto.DynamicStimulusDTO, error)
	CreateStimulusWithItems(ctx context.Context, req *dto.CreateStimulusRequestDTO) error
	UpdateStimulus(ctx context.Context, idStimulus int64, req *dto.UpdateStimulusRequestDTO) error
	DeleteStimulus(ctx context.Context, idStimulus int64) error

	CreateSubItem(ctx context.Context, stimulusID int64, req *dto.CreateSubItemRequestDTO) error
	UpdateSubItem(ctx context.Context, itemID int64, req *dto.UpdateSubItemRequestDTO) error
	DeleteSubItem(ctx context.Context, itemID int64) error

	// CSV Import & Export for Question Banks
	ImportBankQuestionsFromCSV(ctx context.Context, kodesoal string, r io.Reader) (*dto.ImportBankQuestionsResponseDTO, error)
	ExportBankQuestionsToCSV(ctx context.Context, kodesoal string) ([]byte, error)
	GetCSVTemplate() []byte

	// Excel / CSV Export for Exam Scores
	ExportScoresToCSV(ctx context.Context, scheduleID int, examID int) ([]byte, string, error)

	// External Participants (Peserta Umum TOEFL)
	GetExternalParticipants(ctx context.Context, search string, page, perPage int) ([]dto.EPExternalParticipantItemDTO, *response.Pagination, error)
	GetExternalParticipantByCode(ctx context.Context, code string) (*dto.EPExternalParticipantItemDTO, error)
	GetNextExternalParticipantCode(ctx context.Context) (string, error)
	CreateExternalParticipant(ctx context.Context, req *dto.CreateEPExternalParticipantRequestDTO) error
	UpdateExternalParticipant(ctx context.Context, code string, req *dto.UpdateEPExternalParticipantRequestDTO) error
	DeleteExternalParticipant(ctx context.Context, code string) error
	ImportExternalParticipantsFromCSV(ctx context.Context, r io.Reader) (int, error)
}

type epService struct {
	repo repository.EPRepository
}

func NewEPService(repo repository.EPRepository) EPService {
	return &epService{repo: repo}
}

func (s *epService) GetExamTypes(ctx context.Context) ([]dto.ExamTypeDTO, error) {
	types, err := s.repo.GetExamTypes(ctx)
	if err != nil {
		return nil, err
	}

	var dtos []dto.ExamTypeDTO
	for _, t := range types {
		secList, _ := s.repo.GetSectionsByExamType(ctx, t.IDExamType)
		var secDTOs []dto.SectionDTO
		for _, sc := range secList {
			parts, _ := s.repo.GetSectionParts(ctx, sc.IDSection)
			var partDTOs []dto.SectionPartDTO
			for _, p := range parts {
				partDTOs = append(partDTOs, dto.SectionPartDTO{
					IDPart:          p.IDPart,
					PartCode:        p.PartCode,
					PartName:        p.PartName,
					PartOrder:       p.PartOrder,
					QuestionStart:   p.QuestionStart,
					QuestionEnd:     p.QuestionEnd,
					InstructionText: p.InstructionText,
					HasPassage:      p.HasPassage,
					HasAudio:        p.HasAudio,
				})
			}
			secDTOs = append(secDTOs, dto.SectionDTO{
				IDSection:        sc.IDSection,
				IDExamType:       sc.IDExamType,
				SectionCode:      sc.SectionCode,
				SectionName:      sc.SectionName,
				SectionOrder:     sc.SectionOrder,
				DefaultQuestions: sc.DefaultQuestions,
				DefaultDuration:  sc.DefaultDuration,
				HasAudio:         sc.HasAudio,
				AllowReplay:      sc.AllowReplay,
				MaxReplayCount:   sc.MaxReplayCount,
				CanGoBack:        sc.CanGoBack,
				ScoreScaleMin:    sc.ScoreScaleMin,
				ScoreScaleMax:    sc.ScoreScaleMax,
				Parts:            partDTOs,
			})
		}

		dtos = append(dtos, dto.ExamTypeDTO{
			IDExamType:     t.IDExamType,
			TypeCode:       t.TypeCode,
			TypeName:       t.TypeName,
			TotalSections:  t.TotalSections,
			ScoreMin:       t.ScoreMin,
			ScoreMax:       t.ScoreMax,
			ScoreFormula:   t.ScoreFormula,
			ValidityMonths: t.ValidityMonths,
			IsActive:       t.IsActive,
			Sections:       secDTOs,
		})
	}
	return dtos, nil
}

func (s *epService) GetConversionProfiles(ctx context.Context, examTypeID int) ([]dto.ConversionProfileDTO, error) {
	profiles, err := s.repo.GetConversionProfiles(ctx, examTypeID)
	if err != nil {
		return nil, err
	}
	var res []dto.ConversionProfileDTO
	for _, p := range profiles {
		res = append(res, dto.ConversionProfileDTO{
			IDProfile:   p.IDProfile,
			IDExamType:  p.IDExamType,
			ProfileCode: p.ProfileCode,
			ProfileName: p.ProfileName,
			IsDefault:   p.IsDefault,
			Description: p.Description,
		})
	}
	return res, nil
}

func (s *epService) GetConversionTable(ctx context.Context, profileID int) ([]dto.ScoreConversionDTO, error) {
	rows, err := s.repo.GetConversionTable(ctx, profileID)
	if err != nil {
		return nil, err
	}
	var res []dto.ScoreConversionDTO
	for _, r := range rows {
		res = append(res, dto.ScoreConversionDTO{
			SectionCode: r.SectionCode,
			RawScore:    r.RawScore,
			ScaledScore: r.ScaledScore,
		})
	}
	return res, nil
}

func (s *epService) SaveConversionProfile(ctx context.Context, req *dto.SaveConversionProfileRequestDTO) (*dto.ConversionProfileDTO, error) {
	p, err := s.repo.SaveConversionProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	return &dto.ConversionProfileDTO{
		IDProfile:   p.IDProfile,
		IDExamType:  p.IDExamType,
		ProfileCode: p.ProfileCode,
		ProfileName: p.ProfileName,
		IsDefault:   p.IsDefault,
		Description: p.Description,
	}, nil
}

func (s *epService) CreateExam(ctx context.Context, req *dto.CreateEPExamRequestDTO, createdBy string) (*dto.EPExamResponseDTO, error) {
	examEntity := &entity.EPExam{
		IDExamType:             req.IDExamType,
		IDProfile:              req.IDProfile,
		IDTemplate:             req.IDTemplate,
		ExamName:               req.ExamName,
		PassingScore:           req.PassingScore,
		ValidityMonthsOverride: req.ValidityMonthsOverride,
		Description:            req.Description,
		CreatedBy:              &createdBy,
	}

	created, err := s.repo.CreateExam(ctx, examEntity, req.Sections)
	if err != nil {
		return nil, err
	}
	return s.GetExamByID(ctx, created.IDEPExam)
}

func (s *epService) GetExams(ctx context.Context, page, perPage int) ([]dto.EPExamResponseDTO, *response.Pagination, error) {
	exams, total, err := s.repo.GetExams(ctx, page, perPage)
	if err != nil {
		return nil, nil, err
	}

	var res []dto.EPExamResponseDTO
	for _, e := range exams {
		var secDTOs []dto.EPExamSectionItemDTO
		for _, sc := range e.Sections {
			secDTOs = append(secDTOs, dto.EPExamSectionItemDTO{
				IDEPExamSection: sc.IDEPExamSection,
				IDSection:       sc.IDSection,
				SectionCode:     sc.SectionCode,
				SectionName:     sc.SectionName,
				KodeSoal:        sc.KodeSoal,
				SectionOrder:    sc.SectionOrder,
				DurationMinutes: sc.DurationMinutes,
				AllowReplay:     sc.AllowReplay,
				MaxReplayCount:  sc.MaxReplayCount,
				ShuffleOptions:  sc.ShuffleOptions,
				HasAudio:        sc.HasAudio,
			})
		}

		res = append(res, dto.EPExamResponseDTO{
			IDEPExam:               e.IDEPExam,
			IDExamType:             e.IDExamType,
			ExamTypeName:           e.ExamTypeName,
			IDProfile:              e.IDProfile,
			ProfileName:            e.ProfileName,
			IDTemplate:             e.IDTemplate,
			ExamName:               e.ExamName,
			PassingScore:           e.PassingScore,
			ValidityMonthsOverride: e.ValidityMonthsOverride,
			Description:            e.Description,
			IsActive:               e.IsActive,
			CreatedAt:              e.CreatedAt,
			Sections:               secDTOs,
		})
	}

	totalPages := 1
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	pagination := &response.Pagination{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalPages:   totalPages,
		TotalRecords: total,
	}
	return res, pagination, nil
}

func (s *epService) GetExamByID(ctx context.Context, examID int) (*dto.EPExamResponseDTO, error) {
	e, err := s.repo.GetExamByID(ctx, examID)
	if err != nil {
		return nil, err
	}

	var secDTOs []dto.EPExamSectionItemDTO
	for _, sc := range e.Sections {
		secDTOs = append(secDTOs, dto.EPExamSectionItemDTO{
			IDEPExamSection: sc.IDEPExamSection,
			IDSection:       sc.IDSection,
			SectionCode:     sc.SectionCode,
			SectionName:     sc.SectionName,
			KodeSoal:        sc.KodeSoal,
			SectionOrder:    sc.SectionOrder,
			DurationMinutes: sc.DurationMinutes,
			AllowReplay:     sc.AllowReplay,
			MaxReplayCount:  sc.MaxReplayCount,
			ShuffleOptions:  sc.ShuffleOptions,
			HasAudio:        sc.HasAudio,
		})
	}

	return &dto.EPExamResponseDTO{
		IDEPExam:               e.IDEPExam,
		IDExamType:             e.IDExamType,
		ExamTypeName:           e.ExamTypeName,
		IDProfile:              e.IDProfile,
		ProfileName:            e.ProfileName,
		IDTemplate:             e.IDTemplate,
		ExamName:               e.ExamName,
		PassingScore:           e.PassingScore,
		ValidityMonthsOverride: e.ValidityMonthsOverride,
		Description:            e.Description,
		IsActive:               e.IsActive,
		CreatedAt:              e.CreatedAt,
		Sections:               secDTOs,
	}, nil
}

func (s *epService) UpdateExam(ctx context.Context, examID int, req *dto.CreateEPExamRequestDTO) (*dto.EPExamResponseDTO, error) {
	err := s.repo.UpdateExam(ctx, examID, req)
	if err != nil {
		return nil, err
	}
	return s.GetExamByID(ctx, examID)
}

func (s *epService) DeleteExam(ctx context.Context, examID int) error {
	return s.repo.DeleteExam(ctx, examID)
}

func (s *epService) CreateSchedule(ctx context.Context, req *dto.CreateEPScheduleRequestDTO) (*dto.EPScheduleResponseDTO, error) {
	examDate, err := time.Parse("2006-01-02", req.ExamDate)
	if err != nil {
		examDate = time.Now()
	}

	capacity := req.Capacity
	if capacity == 0 {
		capacity = 40
	}

	sched := &entity.EPSchedule{
		IDEPExam:     req.IDEPExam,
		IDRuang:      req.IDRuang,
		ExamDate:     examDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		SessionToken: req.SessionToken,
		Capacity:     capacity,
		ProctorName:  &req.ProctorName,
	}

	created, err := s.repo.CreateSchedule(ctx, sched)
	if err != nil {
		return nil, err
	}

	return &dto.EPScheduleResponseDTO{
		IDEPSchedule: created.IDEPSchedule,
		IDEPExam:     created.IDEPExam,
		IDRuang:      created.IDRuang,
		ExamDate:     created.ExamDate,
		StartTime:    created.StartTime,
		EndTime:      created.EndTime,
		SessionToken: created.SessionToken,
		Capacity:     created.Capacity,
		ProctorName:  created.ProctorName,
		IsActive:     created.IsActive,
	}, nil
}

func (s *epService) GetSchedules(ctx context.Context, examID int) ([]dto.EPScheduleResponseDTO, error) {
	list, err := s.repo.GetSchedules(ctx, examID)
	if err != nil {
		return nil, err
	}
	var res []dto.EPScheduleResponseDTO
	for _, sc := range list {
		res = append(res, dto.EPScheduleResponseDTO{
			IDEPSchedule: sc.IDEPSchedule,
			IDEPExam:     sc.IDEPExam,
			ExamName:     sc.ExamName,
			IDRuang:      sc.IDRuang,
			RoomName:     sc.RoomName,
			ExamDate:     sc.ExamDate,
			StartTime:    sc.StartTime,
			EndTime:      sc.EndTime,
			SessionToken: sc.SessionToken,
			Capacity:     sc.Capacity,
			Registered:   sc.Registered,
			ProctorName:  sc.ProctorName,
			IsActive:     sc.IsActive,
		})
	}
	return res, nil
}

func (s *epService) UpdateSchedule(ctx context.Context, scheduleID int, req *dto.UpdateEPScheduleRequestDTO) (*dto.EPScheduleResponseDTO, error) {
	updated, err := s.repo.UpdateSchedule(ctx, scheduleID, req)
	if err != nil {
		return nil, err
	}
	return &dto.EPScheduleResponseDTO{
		IDEPSchedule: updated.IDEPSchedule,
		IDEPExam:     updated.IDEPExam,
		ExamName:     updated.ExamName,
		IDRuang:      updated.IDRuang,
		RoomName:     updated.RoomName,
		ExamDate:     updated.ExamDate,
		StartTime:    updated.StartTime,
		EndTime:      updated.EndTime,
		SessionToken: updated.SessionToken,
		Capacity:     updated.Capacity,
		Registered:   updated.Registered,
		ProctorName:  updated.ProctorName,
		IsActive:     updated.IsActive,
	}, nil
}

func (s *epService) DeleteSchedule(ctx context.Context, scheduleID int) error {
	return s.repo.DeleteSchedule(ctx, scheduleID)
}

func (s *epService) RegisterParticipants(ctx context.Context, scheduleID int, codes []string, details []dto.RegisterParticipantItemDTO) (int, error) {
	return s.repo.RegisterParticipants(ctx, scheduleID, codes, details)
}

func (s *epService) GetScheduleParticipants(ctx context.Context, scheduleID int) ([]*entity.ScheduleParticipant, error) {
	return s.repo.GetScheduleParticipants(ctx, scheduleID)
}

func (s *epService) StartExamSession(ctx context.Context, scheduleID int, participantCode, sessionToken string) (*dto.EPExamSessionResponseDTO, error) {
	_, err := s.repo.StartParticipantSession(ctx, scheduleID, participantCode, sessionToken)
	if err != nil {
		return nil, err
	}
	return s.repo.GetCurrentSectionSession(ctx, scheduleID, participantCode)
}

func (s *epService) GetCurrentSectionSession(ctx context.Context, scheduleID int, participantCode string) (*dto.EPExamSessionResponseDTO, error) {
	return s.repo.GetCurrentSectionSession(ctx, scheduleID, participantCode)
}

func (s *epService) SaveAnswer(ctx context.Context, participantCode string, req *dto.SaveAnswerDTO) error {
	return s.repo.SaveAnswer(ctx, participantCode, req)
}

func (s *epService) AdvanceToNextSection(ctx context.Context, scheduleID int, participantCode string) (*dto.EPExamSessionResponseDTO, error) {
	return s.repo.AdvanceToNextSection(ctx, scheduleID, participantCode)
}

func (s *epService) FinishExam(ctx context.Context, scheduleID int, participantCode, reason string) error {
	return s.repo.FinishExam(ctx, scheduleID, participantCode, reason)
}

func (s *epService) RemoveParticipantFromSchedule(ctx context.Context, scheduleID int, kodepeserta string) error {
	return s.repo.RemoveParticipantFromSchedule(ctx, scheduleID, kodepeserta)
}

func (s *epService) RecordSecurityEvent(ctx context.Context, scheduleID int, participantCode string, req *dto.EPSecurityEventRequestDTO, ip, userAgent string) (*dto.EPSecurityEventResponseDTO, error) {
	return s.repo.RecordSecurityEvent(ctx, scheduleID, participantCode, req, ip, userAgent)
}

func (s *epService) SyncParticipantHeartbeat(ctx context.Context, scheduleID int, participantCode string, req *dto.EPHeartbeatRequestDTO) (*dto.EPHeartbeatResponseDTO, error) {
	return s.repo.SyncParticipantHeartbeat(ctx, scheduleID, participantCode, req)
}

func (s *epService) ApplyProctorAction(ctx context.Context, req *dto.EPProctorActionRequestDTO) error {
	return s.repo.ApplyProctorAction(ctx, req)
}

func (s *epService) GetSecurityEventsByParticipant(ctx context.Context, scheduleID int, participantCode string) ([]dto.EPSecurityAuditItemDTO, error) {
	events, err := s.repo.GetSecurityEventsByParticipant(ctx, scheduleID, participantCode)
	if err != nil {
		return nil, err
	}
	var res []dto.EPSecurityAuditItemDTO
	for _, ev := range events {
		var meta map[string]interface{}
		if ev.Metadata != nil && *ev.Metadata != "" {
			_ = json.Unmarshal([]byte(*ev.Metadata), &meta)
		}
		res = append(res, dto.EPSecurityAuditItemDTO{
			ID:           ev.ID,
			IDEPSchedule: ev.IDEPSchedule,
			KodePeserta:  ev.KodePeserta,
			EventType:    ev.EventType,
			Severity:     ev.Severity,
			RiskScore:    ev.RiskScore,
			Metadata:     meta,
			CreatedAt:    ev.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return res, nil
}

func (s *epService) SaveProctorSnapshot(ctx context.Context, scheduleID int, participantCode string, req *dto.UploadProctorSnapshotRequestDTO) (*dto.ProctorSnapshotItemDTO, error) {
	return s.repo.SaveProctorSnapshot(ctx, scheduleID, participantCode, req)
}

func (s *epService) GetProctorSnapshots(ctx context.Context, scheduleID int, participantCode, snapshotType string, limit, offset int) ([]dto.ProctorSnapshotItemDTO, error) {
	return s.repo.GetProctorSnapshots(ctx, scheduleID, participantCode, snapshotType, limit, offset)
}

func (s *epService) GetLatestProctorGrid(ctx context.Context, scheduleID int) ([]dto.ProctorLatestGridItemDTO, error) {
	return s.repo.GetLatestProctorGrid(ctx, scheduleID)
}

func (s *epService) CalculateScores(ctx context.Context, scheduleID int) ([]dto.EPScoreDetailDTO, error) {
	scores, err := s.repo.CalculateAndSaveScores(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	return s.mapScores(scores), nil
}

func (s *epService) GetScores(ctx context.Context, scheduleID int) ([]dto.EPScoreDetailDTO, error) {
	scores, err := s.repo.GetScoresBySchedule(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	return s.mapScores(scores), nil
}

func (s *epService) mapScores(scores []*entity.EPScore) []dto.EPScoreDetailDTO {
	var res []dto.EPScoreDetailDTO
	for _, sc := range scores {
		res = append(res, dto.EPScoreDetailDTO{
			IDScore:          sc.IDScore,
			ParticipantCode:  sc.KodePeserta,
			ParticipantName:  sc.Nama,
			ListeningRaw:     sc.ListeningRaw,
			ListeningScaled:  sc.ListeningScaled,
			StructureRaw:     sc.StructureRaw,
			StructureScaled:  sc.StructureScaled,
			ReadingRaw:       sc.ReadingRaw,
			ReadingScaled:    sc.ReadingScaled,
			TotalScaledScore: sc.TotalScaledScore,
			CEFRLevel:        sc.CEFRLevel,
			PassingStatus:    sc.PassingStatus,
			CalculatedAt:     sc.CalculatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return res
}

func (s *epService) GenerateCertificates(ctx context.Context, scheduleID int) ([]*entity.EPCertificate, error) {
	return s.repo.GenerateCertificates(ctx, scheduleID)
}

func (s *epService) GetCertificate(ctx context.Context, certNumber string) (*entity.EPCertificate, error) {
	return s.repo.GetCertificateByNumber(ctx, certNumber)
}

func (s *epService) VerifyCertificate(ctx context.Context, verificationCode string) (*dto.CertificateVerificationDTO, error) {
	cert, err := s.repo.VerifyCertificate(ctx, verificationCode)
	if err != nil {
		return nil, fmt.Errorf("sertifikat dengan kode verifikasi tersebut tidak ditemukan")
	}

	isValid := !cert.IsRevoked && time.Now().Before(cert.ExpiryDate)
	msg := "Sertifikat Asli & Terverifikasi Aktif"
	if cert.IsRevoked {
		msg = fmt.Sprintf("Sertifikat telah dicabut: %s", *cert.RevokedReason)
	} else if time.Now().After(cert.ExpiryDate) {
		msg = "Masa berlaku sertifikat telah kedaluwarsa"
	}

	return &dto.CertificateVerificationDTO{
		CertificateNumber: cert.CertificateNumber,
		ParticipantName:   cert.ParticipantName,
		ParticipantCode:   cert.ParticipantCode,
		ExamTypeName:      cert.ExamTypeName,
		ExamDate:          cert.ExamDate.Format("02 January 2006"),
		ExamLocation:      cert.ExamLocation,
		ListeningScore:    cert.ListeningScore,
		StructureScore:    cert.StructureScore,
		ReadingScore:      cert.ReadingScore,
		TotalScore:        cert.TotalScore,
		CEFRLevel:         cert.CEFRLevel,
		IssuedDate:        cert.IssuedDate.Format("02 January 2006"),
		ExpiryDate:        cert.ExpiryDate.Format("02 January 2006"),
		IsValid:           isValid,
		StatusMessage:     msg,
	}, nil
}

func (s *epService) GetParticipantScores(ctx context.Context, participantCode string) ([]dto.ParticipantScoreResultDTO, error) {
	return s.repo.GetParticipantScoresAndCertificates(ctx, participantCode)
}

func (s *epService) GetStimuliByBank(ctx context.Context, kodesoal string) ([]dto.DynamicStimulusDTO, error) {
	return s.repo.GetStimuliByBank(ctx, kodesoal)
}

func (s *epService) CreateStimulusWithItems(ctx context.Context, req *dto.CreateStimulusRequestDTO) error {
	return s.repo.CreateStimulusWithItems(ctx, req)
}

func (s *epService) UpdateStimulus(ctx context.Context, idStimulus int64, req *dto.UpdateStimulusRequestDTO) error {
	return s.repo.UpdateStimulus(ctx, idStimulus, req)
}

func (s *epService) DeleteStimulus(ctx context.Context, idStimulus int64) error {
	return s.repo.DeleteStimulus(ctx, idStimulus)
}

func (s *epService) CreateSubItem(ctx context.Context, stimulusID int64, req *dto.CreateSubItemRequestDTO) error {
	return s.repo.CreateSubItem(ctx, stimulusID, req)
}

func (s *epService) UpdateSubItem(ctx context.Context, itemID int64, req *dto.UpdateSubItemRequestDTO) error {
	return s.repo.UpdateSubItem(ctx, itemID, req)
}

func (s *epService) DeleteSubItem(ctx context.Context, itemID int64) error {
	return s.repo.DeleteSubItem(ctx, itemID)
}

func (s *epService) GetEPQuestionBanks(ctx context.Context, examTypeID, sectionID int, search string, page, perPage int) ([]*entity.EPQuestionBank, *response.Pagination, error) {
	banks, total, err := s.repo.GetEPQuestionBanks(ctx, examTypeID, sectionID, search, page, perPage)
	if err != nil {
		return nil, nil, err
	}
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}
	return banks, &response.Pagination{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalPages:   totalPages,
		TotalRecords: int64(total),
	}, nil
}

func (s *epService) GetEPQuestionBankByCode(ctx context.Context, kodesoal string) (*entity.EPQuestionBank, error) {
	return s.repo.GetEPQuestionBankByCode(ctx, kodesoal)
}

func (s *epService) CreateEPQuestionBank(ctx context.Context, req *dto.CreateEPBankDTO) (*entity.EPQuestionBank, error) {
	return s.repo.CreateEPQuestionBank(ctx, req)
}

func (s *epService) UpdateEPQuestionBank(ctx context.Context, kodesoal string, req *dto.UpdateEPBankDTO) error {
	return s.repo.UpdateEPQuestionBank(ctx, kodesoal, req)
}

func (s *epService) DeleteEPQuestionBank(ctx context.Context, kodesoal string) error {
	return s.repo.DeleteEPQuestionBank(ctx, kodesoal)
}

func (s *epService) CopyEPQuestionBank(ctx context.Context, req *dto.CopyEPBankDTO, sourceKode string) (*entity.EPQuestionBank, error) {
	return s.repo.CopyEPQuestionBank(ctx, req, sourceKode)
}

func (s *epService) GetCSVTemplate() []byte {
	var buf bytes.Buffer
	buf.WriteString("\xef\xbb\xbf") // UTF-8 BOM
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"kode_stimulus", "judul_stimulus", "kategori_part", "naskah_wacana", "media_type", "media_url",
		"no_soal", "teks_pertanyaan", "opsi_a", "opsi_b", "opsi_c", "opsi_d", "kunci_jawaban", "pembahasan",
	})
	_ = w.Write([]string{
		"STIM_01", "Short Conversation 1", "PART_A", "Man: Are you going to the library tonight?\nWoman: I was planning to, but my car broke down.", "AUDIO", "/uploads/media/sample_audio.mp3",
		"1", "What does the woman imply?", "She will go by bus", "She cannot go to the library", "She will fix the car tomorrow", "She already visited the library", "B", "Wanita tersebut membatalkan rencana ke perpustakaan karena mobilnya mogok.",
	})
	_ = w.Write([]string{
		"STIM_02", "Structure Incomplete Sentence", "STRUCTURE", "The Apollo 11 mission ______ the first humans on the Moon.", "NONE", "",
		"2", "Pilihlah kata kerja yang tepat:", "lands", "landed", "is landing", "was landed", "B", "Kalimat lampau (past tense) membutuhkan kata kerja bentuk kedua (landed).",
	})
	_ = w.Write([]string{
		"STIM_03", "Passage 1: Photosynthesis", "READING", "Photosynthesis is the process by which green plants transform light energy into chemical energy. During photosynthesis in green plants, light energy is captured and used to convert water, carbon dioxide, and minerals into oxygen and energy-rich organic compounds.", "NONE", "",
		"3", "According to the passage, what is the primary role of light energy?", "To capture carbon minerals", "To transform light energy into chemical energy", "To decompose oxygen", "To cool the green leaves", "B", "Berdasarkan wacana paragraf 1.",
	})
	w.Flush()
	return buf.Bytes()
}

func (s *epService) ImportBankQuestionsFromCSV(ctx context.Context, kodesoal string, r io.Reader) (*dto.ImportBankQuestionsResponseDTO, error) {
	rawBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file: %w", err)
	}

	if bytes.HasPrefix(rawBytes, []byte("\xef\xbb\xbf")) {
		rawBytes = bytes.TrimPrefix(rawBytes, []byte("\xef\xbb\xbf"))
	}

	rawStr := string(rawBytes)
	if strings.TrimSpace(rawStr) == "" {
		return nil, fmt.Errorf("file CSV kosong")
	}

	firstLine := rawStr
	if idx := strings.Index(firstLine, "\n"); idx != -1 {
		firstLine = firstLine[:idx]
	}
	delimiter := ','
	if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
		delimiter = ';'
	}

	csvReader := csv.NewReader(strings.NewReader(rawStr))
	csvReader.Comma = delimiter
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("format CSV tidak valid: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("file CSV harus memiliki baris header dan minimal 1 baris soal")
	}

	header := records[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		clean := strings.ToLower(strings.TrimSpace(h))
		clean = strings.ReplaceAll(clean, " ", "_")
		colIdx[clean] = i
	}

	getCol := func(row []string, names ...string) string {
		for _, name := range names {
			if idx, ok := colIdx[name]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
		}
		return ""
	}

	type parsedRow struct {
		stimCode    string
		stimTitle   string
		categoryTag string
		narrative   string
		mediaType   string
		mediaURL    string
		question    string
		optA        string
		optB        string
		optC        string
		optD        string
		correct     string
		explanation string
	}

	var rows []parsedRow
	for lineNum, rec := range records[1:] {
		qText := getCol(rec, "teks_pertanyaan", "question_text", "pertanyaan", "soal")
		if qText == "" {
			continue
		}

		stimCode := getCol(rec, "kode_stimulus", "stimulus_code", "kode")
		stimTitle := getCol(rec, "judul_stimulus", "title", "judul")
		catTag := getCol(rec, "kategori_part", "category_tag", "kategori", "part")
		narrative := getCol(rec, "naskah_wacana", "narrative_text", "wacana", "bacaan")
		mediaType := strings.ToUpper(getCol(rec, "media_type", "tipe_media"))
		mediaURL := getCol(rec, "media_url", "url_media", "audio_file", "file")
		optA := getCol(rec, "opsi_a", "option_a", "a")
		optB := getCol(rec, "opsi_b", "option_b", "b")
		optC := getCol(rec, "opsi_c", "option_c", "c")
		optD := getCol(rec, "opsi_d", "option_d", "d")
		correct := strings.ToUpper(getCol(rec, "kunci_jawaban", "correct_answer", "kunci", "jawaban"))
		explanation := getCol(rec, "pembahasan", "explanation")

		if correct == "" {
			correct = "A"
		}
		if catTag == "" {
			catTag = "Umum"
		}
		if mediaType == "" {
			lowerURL := strings.ToLower(mediaURL)
			if lowerURL != "" && (strings.HasSuffix(lowerURL, ".mp3") || strings.HasSuffix(lowerURL, ".wav") || strings.HasSuffix(lowerURL, ".m4a")) {
				mediaType = "AUDIO"
			} else if lowerURL != "" {
				mediaType = "IMAGE"
			} else {
				mediaType = "NONE"
			}
		}

		if stimCode == "" {
			stimCode = fmt.Sprintf("STIM_%03d", lineNum+1)
		}

		rows = append(rows, parsedRow{
			stimCode:    stimCode,
			stimTitle:   stimTitle,
			categoryTag: catTag,
			narrative:   narrative,
			mediaType:   mediaType,
			mediaURL:    mediaURL,
			question:    qText,
			optA:        optA,
			optB:        optB,
			optC:        optC,
			optD:        optD,
			correct:     correct,
			explanation: explanation,
		})
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("tidak ada baris soal yang valid dalam file CSV")
	}

	stimMap := make(map[string]*dto.CreateStimulusRequestDTO)
	var orderedCodes []string

	for _, r := range rows {
		stim, exists := stimMap[r.stimCode]
		if !exists {
			var mediaURLPtr *string
			if r.mediaURL != "" {
				mediaURLPtr = &r.mediaURL
			}
			narr := r.narrative
			if narr == "" {
				narr = r.question
			}
			title := r.stimTitle
			if title == "" {
				title = fmt.Sprintf("Stimulus %s", r.stimCode)
			}
			stim = &dto.CreateStimulusRequestDTO{
				KodeSoal:        kodesoal,
				StimulusCode:    r.stimCode,
				Title:           title,
				NarrativeText:   narr,
				MediaType:       r.mediaType,
				MediaURL:        mediaURLPtr,
				CategoryTag:     r.categoryTag,
				DifficultyLevel: 2,
				Items:           nil,
			}
			stimMap[r.stimCode] = stim
			orderedCodes = append(orderedCodes, r.stimCode)
		}

		var options []dto.CreateStimulusItemOptionDTO
		labels := []string{"A", "B", "C", "D"}
		texts := []string{r.optA, r.optB, r.optC, r.optD}
		for i := 0; i < 4; i++ {
			t := texts[i]
			if t == "" {
				t = fmt.Sprintf("Pilihan %s", labels[i])
			}
			options = append(options, dto.CreateStimulusItemOptionDTO{
				OptionLabel: labels[i],
				OptionText:  t,
				MediaType:   "NONE",
			})
		}

		var explPtr *string
		if r.explanation != "" {
			explPtr = &r.explanation
		}

		stim.Items = append(stim.Items, dto.CreateStimulusSubItemDTO{
			ItemOrder:     len(stim.Items) + 1,
			QuestionText:  r.question,
			CorrectAnswer: r.correct,
			Explanation:   explPtr,
			Options:       options,
		})
	}

	var stimuliToInsert []dto.CreateStimulusRequestDTO
	for _, code := range orderedCodes {
		stimuliToInsert = append(stimuliToInsert, *stimMap[code])
	}

	sCount, qCount, err := s.repo.ImportBankStimuli(ctx, kodesoal, stimuliToInsert)
	if err != nil {
		return nil, err
	}

	return &dto.ImportBankQuestionsResponseDTO{
		StimulusCount: sCount,
		QuestionCount: qCount,
	}, nil
}

func (s *epService) ExportBankQuestionsToCSV(ctx context.Context, kodesoal string) ([]byte, error) {
	stimuli, err := s.repo.GetStimuliByBank(ctx, kodesoal)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("\xef\xbb\xbf") // UTF-8 BOM
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{
		"kode_stimulus", "judul_stimulus", "kategori_part", "naskah_wacana", "media_type", "media_url",
		"no_soal", "teks_pertanyaan", "opsi_a", "opsi_b", "opsi_c", "opsi_d", "kunci_jawaban", "pembahasan",
	})

	qNum := 1
	for _, stim := range stimuli {
		mediaURL := ""
		if stim.MediaURL != nil {
			mediaURL = *stim.MediaURL
		}
		for _, it := range stim.SubItems {
			optMap := make(map[string]string)
			for _, o := range it.Options {
				optMap[strings.ToUpper(o.OptionLabel)] = o.OptionText
			}
			expl := ""
			if it.Explanation != nil {
				expl = *it.Explanation
			}

			_ = w.Write([]string{
				stim.StimulusCode,
				stim.Title,
				stim.CategoryTag,
				stim.NarrativeText,
				stim.MediaType,
				mediaURL,
				strconv.Itoa(qNum),
				it.QuestionText,
				optMap["A"],
				optMap["B"],
				optMap["C"],
				optMap["D"],
				it.CorrectAnswer,
				expl,
			})
			qNum++
		}
	}
	w.Flush()
	return buf.Bytes(), nil
}

func (s *epService) ExportScoresToCSV(ctx context.Context, scheduleID int, examID int) ([]byte, string, error) {
	var scores []*entity.EPScore
	var filename string

	if scheduleID > 0 {
		_, _ = s.repo.CalculateAndSaveScores(ctx, scheduleID)
		sc, err := s.repo.GetScoresBySchedule(ctx, scheduleID)
		if err != nil {
			return nil, "", err
		}
		scores = sc
		filename = fmt.Sprintf("rekap_nilai_toefl_sesi_%d_%s.csv", scheduleID, time.Now().Format("20060102_150405"))
	} else if examID > 0 {
		schedules, err := s.repo.GetSchedules(ctx, examID)
		if err == nil {
			for _, sch := range schedules {
				_, _ = s.repo.CalculateAndSaveScores(ctx, sch.IDEPSchedule)
			}
		}
		sc, err := s.repo.GetScoresByExam(ctx, examID)
		if err != nil {
			return nil, "", err
		}
		scores = sc
		filename = fmt.Sprintf("rekap_nilai_toefl_paket_%d_%s.csv", examID, time.Now().Format("20060102_150405"))
	} else {
		return nil, "", fmt.Errorf("id jadwal atau id ujian harus ditentukan")
	}

	var buf bytes.Buffer
	// UTF-8 BOM for Microsoft Excel native compatibility
	buf.WriteString("\xef\xbb\xbf")

	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"REKAPITULASI HASIL UJIAN ENGLISH PROFICIENCY (TOEFL / TOEIC)"})
	_ = w.Write([]string{"POLTEKKES KEMENKES"})
	_ = w.Write([]string{fmt.Sprintf("Tanggal Cetak: %s", time.Now().Format("02/01/2006 15:04:05"))})
	_ = w.Write([]string{""})

	headers := []string{
		"No",
		"NIM / Kode Peserta",
		"Nama Peserta",
		"Paket Ujian",
		"Sesi & Ruangan",
		"Listening (Benar)",
		"Listening (Skor)",
		"Structure (Benar)",
		"Structure (Skor)",
		"Reading (Benar)",
		"Reading (Skor)",
		"Total Skor TOEFL",
		"Level CEFR",
		"Status Kelulusan",
		"Waktu Kalkulasi",
	}
	_ = w.Write(headers)

	for i, sc := range scores {
		cefr := "-"
		if sc.CEFRLevel != nil && *sc.CEFRLevel != "" {
			cefr = *sc.CEFRLevel
		}
		status := "TIDAK LULUS"
		if sc.PassingStatus == "PASS" {
			status = "LULUS"
		}
		sesiRuangan := sc.RoomName
		if sc.ExamDate != "" {
			sesiRuangan = fmt.Sprintf("%s (%s %s)", sc.RoomName, sc.ExamDate, sc.StartTime)
		}

		row := []string{
			strconv.Itoa(i + 1),
			sc.KodePeserta,
			sc.Nama,
			sc.ExamName,
			sesiRuangan,
			strconv.Itoa(sc.ListeningRaw),
			strconv.Itoa(sc.ListeningScaled),
			strconv.Itoa(sc.StructureRaw),
			strconv.Itoa(sc.StructureScaled),
			strconv.Itoa(sc.ReadingRaw),
			strconv.Itoa(sc.ReadingScaled),
			strconv.Itoa(sc.TotalScaledScore),
			cefr,
			status,
			sc.CalculatedAt.Format("2006-01-02 15:04"),
		}
		_ = w.Write(row)
	}

	w.Flush()
	return buf.Bytes(), filename, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 12. External Participants (Peserta Umum TOEFL)
// ─────────────────────────────────────────────────────────────────────────────

func (s *epService) GetExternalParticipants(ctx context.Context, search string, page, perPage int) ([]dto.EPExternalParticipantItemDTO, *response.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	list, total, err := s.repo.GetExternalParticipants(ctx, search, perPage, offset)
	if err != nil {
		return nil, nil, err
	}

	var dtos []dto.EPExternalParticipantItemDTO
	for _, p := range list {
		var tglStr *string
		if p.TanggalLahir != nil {
			s := p.TanggalLahir.Format("2006-01-02")
			tglStr = &s
		}
		dtos = append(dtos, dto.EPExternalParticipantItemDTO{
			IDPesertaUmum: p.IDPesertaUmum,
			KodePeserta:   p.KodePeserta,
			NIK:           p.NIK,
			Nama:          p.Nama,
			JenisKelamin:  p.JenisKelamin,
			Email:         p.Email,
			HP:            p.HP,
			Instansi:      p.Instansi,
			TempatLahir:   p.TempatLahir,
			TanggalLahir:  tglStr,
			Alamat:        p.Alamat,
			IsActive:      p.IsActive,
			CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04"),
			UpdatedAt:     p.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}

	totalPages := (total + perPage - 1) / perPage
	pagination := &response.Pagination{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalRecords: int64(total),
		TotalPages:   totalPages,
	}

	return dtos, pagination, nil
}

func (s *epService) GetExternalParticipantByCode(ctx context.Context, code string) (*dto.EPExternalParticipantItemDTO, error) {
	p, err := s.repo.GetExternalParticipantByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	var tglStr *string
	if p.TanggalLahir != nil {
		s := p.TanggalLahir.Format("2006-01-02")
		tglStr = &s
	}
	return &dto.EPExternalParticipantItemDTO{
		IDPesertaUmum: p.IDPesertaUmum,
		KodePeserta:   p.KodePeserta,
		NIK:           p.NIK,
		Nama:          p.Nama,
		JenisKelamin:  p.JenisKelamin,
		Email:         p.Email,
		HP:            p.HP,
		Instansi:      p.Instansi,
		TempatLahir:   p.TempatLahir,
		TanggalLahir:  tglStr,
		Alamat:        p.Alamat,
		IsActive:      p.IsActive,
		CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02 15:04"),
	}, nil
}

func (s *epService) GetNextExternalParticipantCode(ctx context.Context) (string, error) {
	return s.repo.GetNextExternalParticipantCode(ctx)
}

func (s *epService) CreateExternalParticipant(ctx context.Context, req *dto.CreateEPExternalParticipantRequestDTO) error {
	var tgl *time.Time
	if req.TanggalLahir != nil && *req.TanggalLahir != "" {
		parsed, err := time.Parse("2006-01-02", *req.TanggalLahir)
		if err == nil {
			tgl = &parsed
		}
	}
	jk := "L"
	if req.JenisKelamin != "" {
		jk = req.JenisKelamin
	}
	pwd := req.Password
	if pwd == "" {
		if tgl != nil {
			pwd = tgl.Format("02012006")
		} else if req.KodePeserta != "" {
			pwd = req.KodePeserta
		}
	}

	p := &entity.EPPesertaUmum{
		KodePeserta:  req.KodePeserta,
		NIK:          req.NIK,
		Nama:         req.Nama,
		JenisKelamin: jk,
		Email:        req.Email,
		HP:           req.HP,
		Instansi:     req.Instansi,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tgl,
		Alamat:       req.Alamat,
		Password:     pwd,
		IsActive:     true,
	}
	return s.repo.CreateExternalParticipant(ctx, p)
}

func (s *epService) UpdateExternalParticipant(ctx context.Context, code string, req *dto.UpdateEPExternalParticipantRequestDTO) error {
	var tgl *time.Time
	if req.TanggalLahir != nil && *req.TanggalLahir != "" {
		parsed, err := time.Parse("2006-01-02", *req.TanggalLahir)
		if err == nil {
			tgl = &parsed
		}
	}
	jk := "L"
	if req.JenisKelamin != "" {
		jk = req.JenisKelamin
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	pwd := ""
	if req.Password != nil {
		pwd = *req.Password
	}

	p := &entity.EPPesertaUmum{
		KodePeserta:  code,
		NIK:          req.NIK,
		Nama:         req.Nama,
		JenisKelamin: jk,
		Email:        req.Email,
		HP:           req.HP,
		Instansi:     req.Instansi,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tgl,
		Alamat:       req.Alamat,
		Password:     pwd,
		IsActive:     isActive,
	}
	return s.repo.UpdateExternalParticipant(ctx, p)
}

func (s *epService) DeleteExternalParticipant(ctx context.Context, code string) error {
	return s.repo.DeleteExternalParticipant(ctx, code)
}

func (s *epService) ImportExternalParticipantsFromCSV(ctx context.Context, r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("format CSV tidak valid: %w", err)
	}
	if len(rows) < 2 {
		return 0, fmt.Errorf("file CSV kosong atau hanya berisi header")
	}

	header := rows[0]
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	var list []*entity.EPPesertaUmum
	for _, row := range rows[1:] {
		getVal := func(keys ...string) string {
			for _, k := range keys {
				if idx, ok := colIdx[k]; ok && idx < len(row) {
					return strings.TrimSpace(row[idx])
				}
			}
			return ""
		}

		code := getVal("kodepeserta", "kode_peserta", "nomor_peserta", "kode", "no_peserta")
		nama := getVal("nama", "name", "nama_lengkap")
		nik := getVal("nik", "no_ktp", "ktp")
		email := getVal("email")
		hp := getVal("hp", "telepon", "whatsapp", "no_hp")
		instansi := getVal("instansi", "institusi", "asal_instansi", "unit")
		alamat := getVal("alamat")
		jk := getVal("jenis_kelamin", "jk")
		if jk == "" {
			jk = "L"
		}

		if code == "" && nik != "" {
			code = nik
		}
		if code == "" || nama == "" {
			continue
		}
		if email == "" {
			email = code + "@ep.poltekkes-sby.ac.id"
		}
		if hp == "" {
			hp = "08000000000"
		}

		var nikPtr *string
		if nik != "" {
			nikPtr = &nik
		}
		var instansiPtr *string
		if instansi != "" {
			instansiPtr = &instansi
		}
		var alamatPtr *string
		if alamat != "" {
			alamatPtr = &alamat
		}

		tglStr := getVal("tanggal_lahir", "tgl_lahir")
		var tgl *time.Time
		if tglStr != "" {
			if parsed, err := time.Parse("2006-01-02", tglStr); err == nil {
				tgl = &parsed
			}
		}

		list = append(list, &entity.EPPesertaUmum{
			KodePeserta:  code,
			NIK:          nikPtr,
			Nama:         nama,
			JenisKelamin: jk,
			Email:        email,
			HP:           hp,
			Instansi:     instansiPtr,
			TanggalLahir: tgl,
			Alamat:       alamatPtr,
			Password:     code,
			IsActive:     true,
		})
	}

	return s.repo.BatchImportExternalParticipants(ctx, list)
}

