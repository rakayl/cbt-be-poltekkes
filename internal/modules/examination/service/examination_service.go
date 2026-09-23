package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/examination/dto"
	"poltekkes-cat-backend/internal/modules/examination/entity"
	"poltekkes-cat-backend/internal/modules/examination/repository"
	"poltekkes-cat-backend/internal/shared/response"
)

type ExaminationService interface {
	GetExams(ctx context.Context, page, perPage, periodID int, search string) ([]*dto.ExamResponseDTO, *response.Pagination, error)
	GetExamByID(ctx context.Context, examID int) (*dto.ExamResponseDTO, error)
	CreateExam(ctx context.Context, req *dto.CreateExamRequestDTO) (int, error)
	UpdateExam(ctx context.Context, examID int, req *dto.UpdateExamRequestDTO) error
	DeleteExam(ctx context.Context, examID int) error
	GetSchedules(ctx context.Context, examID, periodID int) ([]*dto.ScheduleResponseDTO, error)
	GetExamParticipants(ctx context.Context, examID int) ([]*dto.ExamParticipantDTO, error)
	StartExam(ctx context.Context, participantCode string, req *dto.StartExamRequestDTO) (*dto.StartExamResponseDTO, error)
	GetSessionQuestions(ctx context.Context, participantCode string, scheduleID int) (*dto.SessionQuestionsResponseDTO, error)
	SaveAnswer(ctx context.Context, participantCode string, req *dto.SaveAnswerRequestDTO) (*dto.SaveAnswerResponseDTO, error)
	GetLiveMonitoring(ctx context.Context, scheduleID int) (*dto.LiveMonitoringResponseDTO, error)
	FinishExam(ctx context.Context, participantCode string, scheduleID int, submitReason string) (*dto.FinishExamResponseDTO, error)

	// CBT Security & Proctor Control
	RecordSecurityEvent(ctx context.Context, participantCode, clientIP, userAgent string, req *dto.SecurityEventRequestDTO) (*dto.SecurityEventResponseDTO, error)
	GetRecentSecurityEvents(ctx context.Context, scheduleID int, limit int) ([]*dto.SecurityEventItemDTO, error)
	ApplyProctorAction(ctx context.Context, req *dto.ProctorActionRequestDTO) error
	Heartbeat(ctx context.Context, participantCode string, req *dto.HeartbeatRequestDTO) (*dto.HeartbeatResponseDTO, error)
	GetCollusionAnalysis(ctx context.Context, scheduleID int) (*dto.CollusionReportResponseDTO, error)

	// Dynamic Multimedia & Multi-Item Examination
	GetDynamicExamSession(ctx context.Context, participantCode string, scheduleID int) (*dto.DynamicExamSessionResponseDTO, error)
	SaveDynamicAnswer(ctx context.Context, req *dto.SaveMultiAnswerDTO) error
	FinishDynamicExam(ctx context.Context, participantCode string, scheduleID int, submitReason string) (*dto.FinishExamResponseDTO, error)

	// === Admin CRUD: Jadwal Sesi ===
	GetSessionsByExam(ctx context.Context, examID int) ([]*dto.SessionResponseDTO, error)
	CreateSession(ctx context.Context, req *dto.CreateScheduleRequestDTO) (int, error)
	UpdateSession(ctx context.Context, sessionID int, req *dto.UpdateSessionRequestDTO) error
	DeleteSession(ctx context.Context, sessionID int) error
	RefreshToken(ctx context.Context, scheduleID int) (*dto.RefreshTokenResponseDTO, error)

	// === Admin CRUD: Ruang per Sesi ===
	GetSessionRooms(ctx context.Context, sessionID int) ([]*dto.SessionRoomResponseDTO, error)
	AddRoomToSession(ctx context.Context, req *dto.CreateRoomSessionRequestDTO) (int, error)
	UpdateSessionRoom(ctx context.Context, roomSessionID int, req *dto.UpdateRoomSessionRequestDTO) error
	DeleteSessionRoom(ctx context.Context, roomSessionID int) error

	// === Admin CRUD: Peserta Ujian ===
	AddExamParticipant(ctx context.Context, examID int, participantCode string) error
	RemoveExamParticipant(ctx context.Context, examID int, participantCode string) error
	GetAvailableParticipants(ctx context.Context, examID int) ([]*dto.AvailableParticipantDTO, error)
	ImportSipenmaruToExam(ctx context.Context, examID int, req *dto.ExamImportSipenmaruRequestDTO) (*dto.ExamImportSipenmaruResponseDTO, error)

	// === Admin: Plotting & Ruang Peserta ===
	AutoDistributeParticipants(ctx context.Context, examID int, overwrite bool) (*dto.AutoDistributeResponseDTO, error)
	GetRoomParticipants(ctx context.Context, roomSessionID int) ([]*dto.RoomParticipantDTO, error)
	AddRoomParticipants(ctx context.Context, roomSessionID int, participantCodes []string) error
	RemoveRoomParticipant(ctx context.Context, roomSessionID int, participantCode string) error

	// === Barcode Attendance Verification ===
	VerifyParticipantBarcode(ctx context.Context, participantCode string, req *dto.VerifyBarcodeRequestDTO) (*dto.VerifyBarcodeResponseDTO, error)
	VerifyProctorBarcode(ctx context.Context, proctorUser string, req *dto.VerifyBarcodeRequestDTO) (*dto.VerifyBarcodeResponseDTO, error)
	GetSessionAttendanceSummary(ctx context.Context, scheduleID int) (*dto.AttendanceSummaryResponseDTO, error)
	SetParticipantVerificationStatus(ctx context.Context, participantCode string, scheduleID int, isVerified int, verifier string) error
	GetParticipantSchedules(ctx context.Context, participantCode string) (*dto.ParticipantScheduleListResponseDTO, error)

	// === Admin CRUD: Bank Soal per Sesi ===
	GetSessionBankSoal(ctx context.Context, sessionID int) ([]*dto.SessionBankSoalDTO, error)
	AddBankSoalToSession(ctx context.Context, sessionID int, req *dto.AddBankSoalToSessionDTO) error
	RemoveBankSoalFromSession(ctx context.Context, sessionID int, kodeSoal string) error

	// === Completed Participants & Detailed Review ===
	GetCompletedParticipants(ctx context.Context, examID int, scheduleID int) ([]*dto.CompletedParticipantDTO, error)
	GetParticipantExamResults(ctx context.Context, participantCode string) ([]*dto.ParticipantExamSummaryDTO, error)
	GetParticipantExamResultDetail(ctx context.Context, scheduleID int, participantCode string) (*dto.ParticipantExamResultDetailDTO, error)
}

type examinationService struct {
	repo repository.ExaminationRepository
}

func NewExaminationService(repo repository.ExaminationRepository) ExaminationService {
	return &examinationService{repo: repo}
}

func (s *examinationService) GetExams(ctx context.Context, page, perPage, periodID int, search string) ([]*dto.ExamResponseDTO, *response.Pagination, error) {
	entities, total, err := s.repo.GetExams(ctx, page, perPage, periodID, search)
	if err != nil {
		return nil, nil, err
	}

	var results []*dto.ExamResponseDTO
	for _, e := range entities {
		grade := 0.0
		if e.NilaiMinimal != nil {
			grade = *e.NilaiMinimal
		}

		var updatedAt *string
		if e.TUpdateTime != nil {
			formatted := e.TUpdateTime.Format(time.RFC3339)
			updatedAt = &formatted
		}

		results = append(results, &dto.ExamResponseDTO{
			ExamID:        e.IDUjian,
			ExamName:      e.NamaUjian,
			PeriodID:      e.IDPeriode,
			PeriodName:    e.NamaPeriode,
			PassingGrade:  grade,
			KodeJenis:     e.KodeJenis,
			NamaJenis:     e.NamaJenis,
			IDSatker:      e.IDSatker,
			NamaSatker:    e.NamaSatker,
			IsPercobaan:   e.IsPercobaan,
			Description:   e.Keterangan,
			MaxViolations: e.MaxViolations,
			UpdatedAt:     updatedAt,
			UpdatedBy:     e.TUpdateUser,
		})
	}

	if results == nil {
		results = []*dto.ExamResponseDTO{}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	pagination := &response.Pagination{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalRecords: total,
		TotalPages:   totalPages,
	}

	return results, pagination, nil
}

func (s *examinationService) GetExamByID(ctx context.Context, examID int) (*dto.ExamResponseDTO, error) {
	e, err := s.repo.GetExamByID(ctx, examID)
	if err != nil {
		return nil, err
	}

	grade := 0.0
	if e.NilaiMinimal != nil {
		grade = *e.NilaiMinimal
	}

	var updatedAt *string
	if e.TUpdateTime != nil {
		formatted := e.TUpdateTime.Format(time.RFC3339)
		updatedAt = &formatted
	}

	return &dto.ExamResponseDTO{
		ExamID:        e.IDUjian,
		ExamName:      e.NamaUjian,
		PeriodID:      e.IDPeriode,
		PeriodName:    e.NamaPeriode,
		PassingGrade:  grade,
		KodeJenis:     e.KodeJenis,
		NamaJenis:     e.NamaJenis,
		IDSatker:      e.IDSatker,
		NamaSatker:    e.NamaSatker,
		IsPercobaan:   e.IsPercobaan,
		Description:   e.Keterangan,
		MaxViolations: e.MaxViolations,
		UpdatedAt:     updatedAt,
		UpdatedBy:     e.TUpdateUser,
	}, nil
}

func (s *examinationService) CreateExam(ctx context.Context, req *dto.CreateExamRequestDTO) (int, error) {
	return s.repo.CreateExam(ctx, req)
}

func (s *examinationService) UpdateExam(ctx context.Context, examID int, req *dto.UpdateExamRequestDTO) error {
	return s.repo.UpdateExam(ctx, examID, req)
}

func (s *examinationService) DeleteExam(ctx context.Context, examID int) error {
	return s.repo.DeleteExam(ctx, examID)
}

func (s *examinationService) GetSchedules(ctx context.Context, examID, periodID int) ([]*dto.ScheduleResponseDTO, error) {
	entities, err := s.repo.GetSchedules(ctx, examID, periodID)
	if err != nil {
		return nil, err
	}

	var results []*dto.ScheduleResponseDTO
	for _, sc := range entities {
		results = append(results, &dto.ScheduleResponseDTO{
			ScheduleID:      sc.IDJadwalUjian,
			ExamID:          sc.IDUjian,
			ExamName:        sc.NamaUjian,
			PeriodID:        sc.IDPeriode,
			PeriodName:      sc.NamaPeriode,
			QuestionCode:    sc.KodeSoal,
			RoomName:        sc.NamaRuang,
			ExamDate:        sc.TglUjian,
			StartTime:       sc.JamMulai,
			EndTime:         sc.JamSelesai,
			DurationMinutes: sc.WaktuPengerjaan,
			SessionToken:    sc.TokenUjian,
			Capacity:        sc.Kapasitas,
			MaxViolations:   sc.MaxViolations,
		})
	}
	if results == nil {
		results = []*dto.ScheduleResponseDTO{}
	}

	return results, nil
}

var indoDays = map[time.Weekday]string{
	time.Sunday:    "Minggu",
	time.Monday:    "Senin",
	time.Tuesday:   "Selasa",
	time.Wednesday: "Rabu",
	time.Thursday:  "Kamis",
	time.Friday:    "Jumat",
	time.Saturday:  "Sabtu",
}

func getIndonesianDayName(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			return ""
		}
	}
	dayName, ok := indoDays[t.Weekday()]
	if !ok {
		return ""
	}
	return dayName
}

func (s *examinationService) GetExamParticipants(ctx context.Context, examID int) ([]*dto.ExamParticipantDTO, error) {
	entities, err := s.repo.GetExamParticipants(ctx, examID)
	if err != nil {
		return nil, err
	}

	var results []*dto.ExamParticipantDTO
	for _, p := range entities {
		var waktuMulaiPtr, waktuSelesaiPtr, hariPtr, tglMulaiPtr *string
		if p.IsPlotted == 1 && p.TglMulai != nil && *p.TglMulai != "" {
			tglMulaiPtr = p.TglMulai
			day := getIndonesianDayName(*p.TglMulai)
			if day != "" {
				hariPtr = &day
			}

			dur := 60
			if p.WaktuPengerjaanStr != nil && len(*p.WaktuPengerjaanStr) == 4 {
				hh := int((*p.WaktuPengerjaanStr)[0]-'0')*10 + int((*p.WaktuPengerjaanStr)[1]-'0')
				mm := int((*p.WaktuPengerjaanStr)[2]-'0')*10 + int((*p.WaktuPengerjaanStr)[3]-'0')
				dur = hh*60 + mm
			}

			startStr := ""
			if p.WaktuMulai != nil {
				startStr = *p.WaktuMulai
			}
			endStr := ""
			if p.WaktuSelesai != nil {
				endStr = *p.WaktuSelesai
			}

			sStart, sEnd := formatTimeAndCalculateEnd(startStr, endStr, dur)
			if sStart != "" {
				waktuMulaiPtr = &sStart
			}
			if sEnd != "" {
				waktuSelesaiPtr = &sEnd
			}
		}

		results = append(results, &dto.ExamParticipantDTO{
			ParticipantCode: p.KodePeserta,
			Name:            p.Nama,
			Email:           p.Email,
			CityName:        p.NamaWilayah,
			Score:           p.Nilai,
			IsLoggedIn:      p.IsLoggedIn,
			IsPlotted:       p.IsPlotted == 1,
			SessionNumber:   p.SessionNumber,
			RoomName:        p.RoomName,
			TglMulai:        tglMulaiPtr,
			Hari:            hariPtr,
			WaktuMulai:      waktuMulaiPtr,
			WaktuSelesai:    waktuSelesaiPtr,
			IsVerified:      p.IsVerified,
			BarcodeScannedAt: func() *string {
				if p.BarcodeScannedAt != nil {
					s := p.BarcodeScannedAt.Format(time.RFC3339)
					return &s
				}
				return nil
			}(),
			ScannedBy: p.ScannedBy,
		})
	}

	if results == nil {
		results = []*dto.ExamParticipantDTO{}
	}

	return results, nil
}

func (s *examinationService) StartExam(ctx context.Context, participantCode string, req *dto.StartExamRequestDTO) (*dto.StartExamResponseDTO, error) {
	sched, remainingSeconds, startedAtStr, err := s.repo.StartExam(ctx, participantCode, req.ScheduleID, req.SessionToken, req.DeviceID, req.DeviceFingerprint)
	if err != nil {
		return nil, err
	}

	return &dto.StartExamResponseDTO{
		ScheduleID:       sched.IDJadwalUjian,
		ExamID:           sched.IDUjian,
		QuestionBankCode: sched.KodeSoal,
		DurationMinutes:  sched.WaktuPengerjaan,
		RemainingSeconds: remainingSeconds,
		StartedAt:        startedAtStr,
		ServerTime:       time.Now().Format("15:04:05"),
		IsLocked:         0,
	}, nil
}

func (s *examinationService) GetSessionQuestions(ctx context.Context, participantCode string, scheduleID int) (*dto.SessionQuestionsResponseDTO, error) {
	return s.repo.GetSessionQuestions(ctx, participantCode, scheduleID)
}

func (s *examinationService) SaveAnswer(ctx context.Context, participantCode string, req *dto.SaveAnswerRequestDTO) (*dto.SaveAnswerResponseDTO, error) {
	err := s.repo.SaveAnswer(ctx, participantCode, req.ScheduleID, req.QuestionCode, req.QuestionNumber, req.SelectedOption, req.IsDoubtful)
	if err != nil {
		return nil, err
	}

	return &dto.SaveAnswerResponseDTO{
		QuestionNumber: req.QuestionNumber,
		SelectedOption: req.SelectedOption,
		IsDoubtful:     req.IsDoubtful,
		SavedAt:        time.Now().Format(time.RFC3339),
	}, nil
}

func (s *examinationService) RecordSecurityEvent(ctx context.Context, participantCode, clientIP, userAgent string, req *dto.SecurityEventRequestDTO) (*dto.SecurityEventResponseDTO, error) {
	// 1. Calculate risk score weight based on event type
	riskScore := 5
	severity := "LOW"

	switch req.EventType {
	case "PROCTOR_SNAP":
		riskScore = 0
		severity = "INFO"
	case "TAB_SWITCH", "FULLSCREEN_EXIT", "FULLSCREEN_REQUIRED":
		riskScore = 10
		severity = "MEDIUM"
	case "WINDOW_BLUR":
		riskScore = 5
		severity = "LOW"
	case "COPY_ATTEMPT", "PASTE_ATTEMPT", "CUT_ATTEMPT":
		riskScore = 5
		severity = "MEDIUM"
	case "PRINT_ATTEMPT":
		riskScore = 15
		severity = "HIGH"
	case "DEVTOOLS_ATTEMPT", "FORBIDDEN_SHORTCUT", "DEVTOOLS_OPEN":
		riskScore = 15
		severity = "HIGH"
	case "SCREENSHOT_ATTEMPT":
		riskScore = 20
		severity = "HIGH"
	case "CAMERA_DISCONNECTED", "CAMERA_BLACKOUT":
		riskScore = 15
		severity = "HIGH"
	case "MULTI_MONITOR":
		riskScore = 20
		severity = "HIGH"
	case "VM_DETECTED":
		riskScore = 25
		severity = "HIGH"
	case "UNTRUSTED_EVENT":
		riskScore = 20
		severity = "HIGH"
	case "SUSPICIOUS_AUDIO":
		riskScore = 15
		severity = "MEDIUM"
	case "HEAD_POSE_ALERT":
		riskScore = 15
		severity = "MEDIUM"
	case "MULTI_FACE_DETECTED":
		riskScore = 25
		severity = "HIGH"
	case "NO_FACE_DETECTED":
		riskScore = 15
		severity = "MEDIUM"
	case "NETWORK_DISCONNECTED":
		riskScore = 10
		severity = "MEDIUM"
	case "LONG_IDLE":
		riskScore = 10
		severity = "MEDIUM"
	case "DEVICE_CHANGE":
		riskScore = 25
		severity = "HIGH"
	case "CONCURRENT_LOGIN":
		riskScore = 30
		severity = "CRITICAL"
	}

	var metaStr *string
	if req.Metadata != nil {
		bytes, _ := json.Marshal(req.Metadata)
		s := string(bytes)
		metaStr = &s
	}

	eventEntity := &entity.SecurityEvent{
		IDJadwal:    req.ScheduleID,
		KodePeserta: participantCode,
		EventType:   req.EventType,
		Severity:    severity,
		RiskScore:   riskScore,
		Metadata:    metaStr,
		IPAddress:   &clientIP,
		UserAgent:   &userAgent,
		DeviceID:    &req.DeviceID,
	}

	return s.repo.RecordSecurityEvent(ctx, eventEntity)
}

func (s *examinationService) GetRecentSecurityEvents(ctx context.Context, scheduleID int, limit int) ([]*dto.SecurityEventItemDTO, error) {
	events, err := s.repo.GetRecentSecurityEvents(ctx, scheduleID, limit)
	if err != nil {
		return nil, err
	}

	var results []*dto.SecurityEventItemDTO
	for _, ev := range events {
		var meta map[string]interface{}
		if ev.Metadata != nil {
			_ = json.Unmarshal([]byte(*ev.Metadata), &meta)
		}
		ip := ""
		if ev.IPAddress != nil {
			ip = *ev.IPAddress
		}

		results = append(results, &dto.SecurityEventItemDTO{
			ID:              ev.ID,
			ScheduleID:      ev.IDJadwal,
			ParticipantCode: ev.KodePeserta,
			EventType:       ev.EventType,
			Severity:        ev.Severity,
			RiskScore:       ev.RiskScore,
			Metadata:        meta,
			IPAddress:       ip,
			CreatedAt:       ev.CreatedAt.Format("15:04:05"),
		})
	}
	return results, nil
}

func (s *examinationService) ApplyProctorAction(ctx context.Context, req *dto.ProctorActionRequestDTO) error {
	return s.repo.ApplyProctorAction(ctx, req.ScheduleID, req.ParticipantCode, req.Action, req.Reason, req.WarningMessage)
}

func (s *examinationService) Heartbeat(ctx context.Context, participantCode string, req *dto.HeartbeatRequestDTO) (*dto.HeartbeatResponseDTO, error) {
	return s.repo.UpdateHeartbeatAndCheckStatus(ctx, req.ScheduleID, participantCode, req)
}

func (s *examinationService) GetLiveMonitoring(ctx context.Context, scheduleID int) (*dto.LiveMonitoringResponseDTO, error) {
	rows, err := s.repo.GetLiveMonitoring(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	var participants []dto.LiveMonitoringParticipantDTO
	var inProgress, completed, ready, locked, highRisk int
	now := time.Now()

	for _, item := range rows {
		if item.StatusUjian == "COMPLETED" {
			completed++
		} else if item.StatusUjian == "IN_PROGRESS" || item.JumlahTerjawab > 0 {
			inProgress++
		} else {
			ready++
		}

		if item.IsLocked == 1 {
			locked++
		}
		if item.RiskScore >= 41 {
			highRisk++
		}

		// Calculate remaining seconds
		remSec := 5400
		if item.WaktuMulai != nil {
			elapsed := int(now.Sub(*item.WaktuMulai).Seconds())
			remSec = 5400 - elapsed
			if remSec < 0 {
				remSec = 0
			}
		}

		connStatus := "CONNECTED"
		if item.LastHeartbeat != nil {
			diff := now.Sub(*item.LastHeartbeat).Seconds()
			if diff > 45 {
				connStatus = "DISCONNECTED"
			}
		}

		lastAct := now.Format("15:04:05")
		if item.LastHeartbeat != nil {
			lastAct = item.LastHeartbeat.Format("15:04:05")
		}

		totQ := item.TotalQuestions
		if totQ <= 0 {
			totQ = 10
		}
		participants = append(participants, dto.LiveMonitoringParticipantDTO{
			ParticipantCode:           item.KodePeserta,
			Name:                      item.Nama,
			DeskNumber:                item.NoUrutPeserta,
			ExamStatus:                item.StatusUjian,
			TotalQuestions:            totQ,
			AnsweredCount:             item.JumlahTerjawab,
			RemainingSeconds:          remSec,
			ConnectionStatus:          connStatus,
			LastActivity:              lastAct,
			RiskScore:                 item.RiskScore,
			RiskLevel:                 item.RiskLevel,
			TabSwitchCount:            item.TabSwitchCount,
			FullscreenExitCount:       item.FullscreenExitCount,
			ActiveTabSwitchCount:      item.ActiveTabSwitchCount,
			ActiveFullscreenExitCount: item.ActiveFullscreenExitCount,
			TotalViolations:           item.TotalViolations,
			UnlockCount:               item.UnlockCount,
			DisconnectCount:           item.DisconnectCount,
			ReconnectCount:            item.ReconnectCount,
			IsLocked:                  item.IsLocked,
			LockReason:                item.LockReason,
			ProctorWarning:            item.ProctorWarning,
			IsVerified:                item.IsVerified,
			BarcodeScannedAt: func() *string {
				if item.BarcodeScannedAt != nil {
					s := item.BarcodeScannedAt.Format(time.RFC3339)
					return &s
				}
				return nil
			}(),
		})
	}

	stats := map[string]interface{}{
		"total_participants": len(participants),
		"in_progress":        inProgress,
		"completed":          completed,
		"not_started":        ready,
		"locked_count":       locked,
		"high_risk_count":    highRisk,
	}

	return &dto.LiveMonitoringResponseDTO{
		ScheduleID:   scheduleID,
		Stats:        stats,
		Participants: participants,
	}, nil
}

func (s *examinationService) FinishExam(ctx context.Context, participantCode string, scheduleID int, submitReason string) (*dto.FinishExamResponseDTO, error) {
	if submitReason == "" {
		submitReason = "MANUAL_SUBMIT"
	}
	_, total, correct, wrong, empty, finalScore, statusLulus, err := s.repo.FinishExam(ctx, participantCode, scheduleID, submitReason)
	if err != nil {
		return nil, err
	}

	return &dto.FinishExamResponseDTO{
		ParticipantCode: participantCode,
		ScheduleID:      scheduleID,
		TotalQuestions:  total,
		CorrectCount:    correct,
		WrongCount:      wrong,
		EmptyCount:      empty,
		FinalScore:      finalScore,
		PassingStatus:   statusLulus,
		SubmittedAt:     time.Now().Format(time.RFC3339),
		SubmitReason:    submitReason,
	}, nil
}

func (s *examinationService) GetCollusionAnalysis(ctx context.Context, scheduleID int) (*dto.CollusionReportResponseDTO, error) {
	return s.repo.GetCollusionAnalysis(ctx, scheduleID)
}

func (s *examinationService) GetDynamicExamSession(ctx context.Context, participantCode string, scheduleID int) (*dto.DynamicExamSessionResponseDTO, error) {
	return s.repo.GetDynamicExamSession(ctx, participantCode, scheduleID)
}

func (s *examinationService) SaveDynamicAnswer(ctx context.Context, req *dto.SaveMultiAnswerDTO) error {
	return s.repo.SaveDynamicAnswer(ctx, req)
}

func (s *examinationService) FinishDynamicExam(ctx context.Context, participantCode string, scheduleID int, submitReason string) (*dto.FinishExamResponseDTO, error) {
	return s.repo.FinishDynamicExam(ctx, participantCode, scheduleID, submitReason)
}

// =====================================================================
// Admin CRUD: Jadwal Sesi
// =====================================================================

func formatTimeToHHMM(str string) string {
	clean := strings.TrimSpace(strings.ReplaceAll(str, ":", ""))
	if len(clean) == 4 {
		hh, errH := strconv.Atoi(clean[:2])
		mm, errM := strconv.Atoi(clean[2:])
		if errH == nil && errM == nil {
			return fmt.Sprintf("%02d:%02d", hh, mm)
		}
	}
	return str
}

func formatTimeAndCalculateEnd(waktuMulaiStr, waktuSelesaiStr string, durationMinutes int) (string, string) {
	if waktuMulaiStr == "" {
		return "", ""
	}
	startFormatted := formatTimeToHHMM(waktuMulaiStr)

	// If explicit waktu_selesai is provided/stored, use it directly!
	if strings.TrimSpace(waktuSelesaiStr) != "" {
		return startFormatted, formatTimeToHHMM(waktuSelesaiStr)
	}

	// Fallback calculation using duration
	clean := strings.TrimSpace(strings.ReplaceAll(waktuMulaiStr, ":", ""))
	if len(clean) == 4 {
		hh, _ := strconv.Atoi(clean[:2])
		mm, _ := strconv.Atoi(clean[2:])
		totalMinutes := hh*60 + mm + durationMinutes
		endH := (totalMinutes / 60) % 24
		endM := totalMinutes % 60
		return startFormatted, fmt.Sprintf("%02d:%02d", endH, endM)
	}

	return startFormatted, ""
}

func (s *examinationService) GetSessionsByExam(ctx context.Context, examID int) ([]*dto.SessionResponseDTO, error) {
	sessions, err := s.repo.GetSessionsByExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	var results []*dto.SessionResponseDTO
	for _, sess := range sessions {
		// Parse waktupengerjaan from HHMM string to minutes
		waktuMenit := 0
		if len(sess.WaktuPengerjaan) == 4 {
			hh := int(sess.WaktuPengerjaan[0]-'0')*10 + int(sess.WaktuPengerjaan[1]-'0')
			mm := int(sess.WaktuPengerjaan[2]-'0')*10 + int(sess.WaktuPengerjaan[3]-'0')
			waktuMenit = hh*60 + mm
		}

		waktuMulai := ""
		waktuSelesai := ""
		if sess.WaktuMulai != nil && *sess.WaktuMulai != "" {
			endStr := ""
			if sess.WaktuSelesai != nil {
				endStr = *sess.WaktuSelesai
			}
			waktuMulai, waktuSelesai = formatTimeAndCalculateEnd(*sess.WaktuMulai, endStr, waktuMenit)
		}

		tglMulai := ""
		if sess.TglMulai != nil {
			tglMulai = *sess.TglMulai
		}
		tglSelesai := ""
		if sess.TglSelesai != nil {
			tglSelesai = *sess.TglSelesai
		} else if tglMulai != "" {
			tglSelesai = tglMulai
		}

		results = append(results, &dto.SessionResponseDTO{
			SessionID:          sess.IDJadwalUjian,
			ExamID:             sess.IDUjian,
			NoJadwal:           sess.NoJadwal,
			WaktuPengerjaan:    waktuMenit,
			Bobot:              sess.Bobot,
			TampilkanNilai:     sess.TampilkanNilai,
			ExamDeadlineAt:     sess.ExamDeadlineAt,
			GracePeriodSeconds: sess.GracePeriodSeconds,
			RoomCount:          sess.RoomCount,
			WaktuMulai:         waktuMulai,
			WaktuSelesai:       waktuSelesai,
			TglMulai:           tglMulai,
			TglSelesai:         tglSelesai,
			TokenUjian:         sess.TokenUjian,
			MaxViolations:      sess.MaxViolations,
		})
	}
	if results == nil {
		results = []*dto.SessionResponseDTO{}
	}
	return results, nil
}

func (s *examinationService) CreateSession(ctx context.Context, req *dto.CreateScheduleRequestDTO) (int, error) {
	return s.repo.CreateSession(ctx, req)
}

func (s *examinationService) UpdateSession(ctx context.Context, sessionID int, req *dto.UpdateSessionRequestDTO) error {
	return s.repo.UpdateSession(ctx, sessionID, req)
}

func (s *examinationService) DeleteSession(ctx context.Context, sessionID int) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

func (s *examinationService) RefreshToken(ctx context.Context, scheduleID int) (*dto.RefreshTokenResponseDTO, error) {
	newToken, err := s.repo.RefreshToken(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	return &dto.RefreshTokenResponseDTO{
		ScheduleID: scheduleID,
		TokenUjian: newToken,
		Message:    "Token ujian berhasil diperbarui",
	}, nil
}

// =====================================================================
// Admin CRUD: Ruang per Sesi
// =====================================================================

func (s *examinationService) GetSessionRooms(ctx context.Context, sessionID int) ([]*dto.SessionRoomResponseDTO, error) {
	rooms, err := s.repo.GetSessionRooms(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	var results []*dto.SessionRoomResponseDTO
	for _, r := range rooms {
		durationMenit := 60
		if len(r.WaktuPengerjaanStr) == 4 {
			hh := int(r.WaktuPengerjaanStr[0]-'0')*10 + int(r.WaktuPengerjaanStr[1]-'0')
			mm := int(r.WaktuPengerjaanStr[2]-'0')*10 + int(r.WaktuPengerjaanStr[3]-'0')
			durationMenit = hh*60 + mm
		}
		waktuMulai, waktuSelesai := formatTimeAndCalculateEnd(r.WaktuMulai, r.WaktuSelesai, durationMenit)

		results = append(results, &dto.SessionRoomResponseDTO{
			RoomSessionID: r.IDRuangUjian,
			SessionID:     r.IDJadwalUjian,
			KodeRuang:     r.KodeRuang,
			NamaRuang:     r.NamaRuang,
			TglMulai:      r.TglMulai,
			TglSelesai:    r.TglSelesai,
			WaktuMulai:    waktuMulai,
			WaktuSelesai:  waktuSelesai,
			JumlahPeserta: r.JumlahPeserta,
			Prioritas:     r.Prioritas,
		})
	}
	if results == nil {
		results = []*dto.SessionRoomResponseDTO{}
	}
	return results, nil
}

func (s *examinationService) AddRoomToSession(ctx context.Context, req *dto.CreateRoomSessionRequestDTO) (int, error) {
	return s.repo.AddRoomToSession(ctx, req)
}

func (s *examinationService) UpdateSessionRoom(ctx context.Context, roomSessionID int, req *dto.UpdateRoomSessionRequestDTO) error {
	return s.repo.UpdateSessionRoom(ctx, roomSessionID, req)
}

func (s *examinationService) DeleteSessionRoom(ctx context.Context, roomSessionID int) error {
	return s.repo.DeleteSessionRoom(ctx, roomSessionID)
}

// =====================================================================
// Admin CRUD: Peserta Ujian
// =====================================================================

func (s *examinationService) AddExamParticipant(ctx context.Context, examID int, participantCode string) error {
	return s.repo.AddExamParticipant(ctx, examID, participantCode)
}

func (s *examinationService) RemoveExamParticipant(ctx context.Context, examID int, participantCode string) error {
	return s.repo.RemoveExamParticipant(ctx, examID, participantCode)
}

func (s *examinationService) GetAvailableParticipants(ctx context.Context, examID int) ([]*dto.AvailableParticipantDTO, error) {
	list, err := s.repo.GetAvailableParticipants(ctx, examID)
	if err != nil {
		return nil, err
	}
	var results []*dto.AvailableParticipantDTO
	for _, p := range list {
		results = append(results, &dto.AvailableParticipantDTO{
			ParticipantCode: p.KodePeserta,
			Name:            p.Nama,
			Email:           p.Email,
			HP:              p.HP,
		})
	}
	if results == nil {
		results = []*dto.AvailableParticipantDTO{}
	}
	return results, nil
}

func (s *examinationService) ImportSipenmaruToExam(ctx context.Context, examID int, req *dto.ExamImportSipenmaruRequestDTO) (*dto.ExamImportSipenmaruResponseDTO, error) {
	return s.repo.ImportSipenmaruToExam(ctx, examID, req)
}

// =====================================================================
// Admin: Plotting & Ruang Peserta
// =====================================================================

func (s *examinationService) AutoDistributeParticipants(ctx context.Context, examID int, overwrite bool) (*dto.AutoDistributeResponseDTO, error) {
	return s.repo.AutoDistributeParticipants(ctx, examID, overwrite)
}

func (s *examinationService) GetRoomParticipants(ctx context.Context, roomSessionID int) ([]*dto.RoomParticipantDTO, error) {
	list, err := s.repo.GetRoomParticipants(ctx, roomSessionID)
	if err != nil {
		return nil, err
	}
	var results []*dto.RoomParticipantDTO
	for _, p := range list {
		results = append(results, &dto.RoomParticipantDTO{
			ParticipantCode: p.KodePeserta,
			Name:            p.Nama,
			Email:           p.Email,
			CityName:        p.NamaWilayah,
			SessionID:       p.IDJadwalUjian,
			RoomSessionID:   p.IDRuangUjian,
			RoomName:        p.NamaRuang,
			IsLoggedIn:      p.IsLogin,
			IsVerified:      p.IsVerified,
			BarcodeScannedAt: func() *string {
				if p.BarcodeScannedAt != nil {
					s := p.BarcodeScannedAt.Format(time.RFC3339)
					return &s
				}
				return nil
			}(),
		})
	}
	if results == nil {
		results = []*dto.RoomParticipantDTO{}
	}
	return results, nil
}

func (s *examinationService) AddRoomParticipants(ctx context.Context, roomSessionID int, participantCodes []string) error {
	return s.repo.AddRoomParticipants(ctx, roomSessionID, participantCodes)
}

func (s *examinationService) RemoveRoomParticipant(ctx context.Context, roomSessionID int, participantCode string) error {
	return s.repo.RemoveRoomParticipant(ctx, roomSessionID, participantCode)
}

// =====================================================================
// Barcode Attendance Verification
// =====================================================================

func (s *examinationService) VerifyParticipantBarcode(ctx context.Context, participantCode string, req *dto.VerifyBarcodeRequestDTO) (*dto.VerifyBarcodeResponseDTO, error) {
	return s.repo.VerifyParticipantBarcode(ctx, req.ScheduleID, participantCode, req.Barcode)
}

func (s *examinationService) VerifyProctorBarcode(ctx context.Context, proctorUser string, req *dto.VerifyBarcodeRequestDTO) (*dto.VerifyBarcodeResponseDTO, error) {
	return s.repo.VerifyProctorBarcode(ctx, req.ScheduleID, req.Barcode, proctorUser)
}

func (s *examinationService) GetSessionAttendanceSummary(ctx context.Context, scheduleID int) (*dto.AttendanceSummaryResponseDTO, error) {
	return s.repo.GetSessionAttendanceSummary(ctx, scheduleID)
}

func (s *examinationService) SetParticipantVerificationStatus(ctx context.Context, participantCode string, scheduleID int, isVerified int, verifier string) error {
	return s.repo.SetParticipantVerificationStatus(ctx, scheduleID, participantCode, isVerified, verifier)
}

func (s *examinationService) GetParticipantSchedules(ctx context.Context, participantCode string) (*dto.ParticipantScheduleListResponseDTO, error) {
	schedules, err := s.repo.GetParticipantSchedules(ctx, participantCode)
	if err != nil {
		return nil, err
	}
	return &dto.ParticipantScheduleListResponseDTO{
		ActiveSchedules: schedules,
	}, nil
}

// helper: keep entity import alive
var _ = entity.AvailableParticipant{}

// === Bank Soal per Sesi ===

func (s *examinationService) GetSessionBankSoal(ctx context.Context, sessionID int) ([]*dto.SessionBankSoalDTO, error) {
	entities, err := s.repo.GetSessionBankSoal(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat bank soal sesi: %w", err)
	}
	results := make([]*dto.SessionBankSoalDTO, 0)
	for _, e := range entities {
		results = append(results, &dto.SessionBankSoalDTO{
			KodeSoal:               e.KodeSoal,
			NamaSoal:               e.NamaSoal,
			NoUrut:                 e.NoUrut,
			JumlahPertanyaan:       e.JumlahPertanyaan,
			JumlahSoalStatis:       e.JumlahSoalStatis,
			JumlahKasusDinamis:     e.JumlahKasusDinamis,
			Bobot:                  e.Bobot,
			KodeSkor:               e.KodeSkor,
			TotalAvailable:         e.TotalAvailable,
			TotalStatisAvailable:   e.TotalStatisAvailable,
			TotalStimulusAvailable: e.TotalStimulusAvailable,
		})
	}
	return results, nil
}

func (s *examinationService) AddBankSoalToSession(ctx context.Context, sessionID int, req *dto.AddBankSoalToSessionDTO) error {
	return s.repo.AddBankSoalToSession(ctx, sessionID, req)
}

func (s *examinationService) RemoveBankSoalFromSession(ctx context.Context, sessionID int, kodeSoal string) error {
	return s.repo.RemoveBankSoalFromSession(ctx, sessionID, kodeSoal)
}

func (s *examinationService) GetCompletedParticipants(ctx context.Context, examID int, scheduleID int) ([]*dto.CompletedParticipantDTO, error) {
	return s.repo.GetCompletedParticipants(ctx, examID, scheduleID)
}

func (s *examinationService) GetParticipantExamResults(ctx context.Context, participantCode string) ([]*dto.ParticipantExamSummaryDTO, error) {
	return s.repo.GetParticipantCompletedExams(ctx, participantCode)
}

func (s *examinationService) GetParticipantExamResultDetail(ctx context.Context, scheduleID int, participantCode string) (*dto.ParticipantExamResultDetailDTO, error) {
	return s.repo.GetParticipantExamResultDetail(ctx, scheduleID, participantCode)
}
