package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/english_proficiency/dto"
	"poltekkes-cat-backend/internal/modules/english_proficiency/service"
	"poltekkes-cat-backend/internal/shared/middleware"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EPHandler struct {
	service service.EPService
}

func NewEPHandler(service service.EPService) *EPHandler {
	return &EPHandler{service: service}
}

func (h *EPHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// Public Verification Endpoint
	rg.GET("/verify/:verification_code", h.VerifyCertificate)

	// Admin Routes
	admin := rg.Group("")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		// Exam Types & Sections
		admin.GET("/exam-types", h.GetExamTypes)

		// Conversion Profiles
		admin.GET("/conversion-profiles", h.GetConversionProfiles)
		admin.GET("/conversion-profiles/:profile_id/table", h.GetConversionTable)
		admin.POST("/conversion-profiles", h.SaveConversionProfile)

		// Exams
		admin.POST("/exams", h.CreateExam)
		admin.GET("/exams", h.GetExams)
		admin.GET("/exams/:exam_id", h.GetExamByID)
		admin.PUT("/exams/:exam_id", h.UpdateExam)
		admin.DELETE("/exams/:exam_id", h.DeleteExam)

		// Schedules & Participants
		admin.POST("/schedules", h.CreateSchedule)
		admin.GET("/schedules", h.GetSchedules)
		admin.PUT("/schedules/:schedule_id", h.UpdateSchedule)
		admin.DELETE("/schedules/:schedule_id", h.DeleteSchedule)
		admin.POST("/schedules/:schedule_id/participants", h.RegisterParticipants)
		admin.GET("/schedules/:schedule_id/participants", h.GetScheduleParticipants)
		admin.DELETE("/schedules/:schedule_id/participants/:kodepeserta", h.RemoveParticipant)

		// External Participants (Peserta Umum TOEFL)
		admin.GET("/participants/external", h.GetExternalParticipants)
		admin.GET("/participants/external/next-code", h.GetNextExternalParticipantCode)
		admin.GET("/participants/external/:code", h.GetExternalParticipantByCode)
		admin.POST("/participants/external", h.CreateExternalParticipant)
		admin.PUT("/participants/external/:code", h.UpdateExternalParticipant)
		admin.DELETE("/participants/external/:code", h.DeleteExternalParticipant)
		admin.POST("/participants/external/import", h.ImportExternalParticipants)

		// Scoring & Certificates
		admin.POST("/scoring/:schedule_id/calculate", h.CalculateScores)
		admin.GET("/scoring/:schedule_id", h.GetScores)
		admin.GET("/schedules/:schedule_id/scores/export", h.ExportScheduleScores)
		admin.GET("/exams/:exam_id/scores/export", h.ExportExamScores)
		admin.POST("/certificates/:schedule_id/generate", h.GenerateCertificates)

		// Question Banks (Header & Details)
		admin.GET("/banks", h.GetEPQuestionBanks)
		admin.POST("/banks", h.CreateEPQuestionBank)
		admin.GET("/banks/:kodesoal", h.GetEPQuestionBankByCode)
		admin.PUT("/banks/:kodesoal", h.UpdateEPQuestionBank)
		admin.DELETE("/banks/:kodesoal", h.DeleteEPQuestionBank)
		admin.POST("/banks/:kodesoal/copy", h.CopyEPQuestionBank)

		// Question Bank Items (Stimuli & Sub-Questions)
		admin.GET("/banks/:kodesoal/questions", h.GetBankQuestions)
		admin.POST("/banks/stimulus", h.CreateBankStimulus)
		admin.PUT("/banks/stimulus/:id_stimulus", h.UpdateBankStimulus)
		admin.DELETE("/banks/stimulus/:id_stimulus", h.DeleteBankStimulus)

		// Sub-Questions (Dynamic Item Level)
		admin.POST("/banks/stimulus/:id_stimulus/items", h.CreateSubItem)
		admin.PUT("/banks/items/:id_item", h.UpdateSubItem)
		admin.DELETE("/banks/items/:id_item", h.DeleteSubItem)

		// CSV Template, Import & Export for Question Banks
		admin.GET("/banks/template/csv", h.DownloadCSVTemplate)
		admin.GET("/banks/:kodesoal/export", h.ExportBankQuestions)
		admin.POST("/banks/:kodesoal/import", h.ImportBankQuestions)

		// Multimedia File Upload (Audio, Images, Video)
		admin.POST("/upload-media", h.UploadMedia)

		// Anti-Cheat & Proctor Live Control
		admin.POST("/proctor/action", h.ApplyProctorAction)
		admin.GET("/monitoring/:schedule_id/events/:kodepeserta", h.GetSecurityEventsByParticipant)
		admin.GET("/monitoring/:schedule_id/events", h.GetSecurityEventsByParticipant)
		admin.GET("/monitoring/:schedule_id/snapshots", h.GetProctorSnapshots)
		admin.GET("/monitoring/:schedule_id/snapshots/latest-grid", h.GetLatestProctorGrid)
	}

	// Participant Routes
	participant := rg.Group("/session")
	participant.Use(middleware.ParticipantAuthMiddleware())
	{
		participant.POST("/start", h.StartExamSession)
		participant.GET("/:schedule_id/current", h.GetCurrentSectionSession)
		participant.POST("/answer", h.SaveAnswer)
		participant.POST("/next-section", h.AdvanceToNextSection)
		participant.POST("/finish", h.FinishExam)
		participant.GET("/my-results", h.GetMyResults)

		// Anti-Cheat Telemetry & Proctor Snapshots
		participant.POST("/events", h.RecordSecurityEvent)
		participant.POST("/heartbeat", h.SyncParticipantHeartbeat)
		participant.POST("/snapshot", h.UploadProctorSnapshot)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Admin Handler Methods
// ─────────────────────────────────────────────────────────────────────────────
func (h *EPHandler) GetExamTypes(c *gin.Context) {
	types, err := h.service.GetExamTypes(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat jenis ujian: "+err.Error())
		return
	}
	response.Success(c, "Data jenis ujian berhasil dimuat", types)
}

func (h *EPHandler) GetConversionProfiles(c *gin.Context) {
	typeID, _ := strconv.Atoi(c.DefaultQuery("exam_type_id", "1"))
	profiles, err := h.service.GetConversionProfiles(c.Request.Context(), typeID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat profil konversi: "+err.Error())
		return
	}
	response.Success(c, "Data profil konversi skor berhasil dimuat", profiles)
}

func (h *EPHandler) GetConversionTable(c *gin.Context) {
	profileID, _ := strconv.Atoi(c.Param("profile_id"))
	tables, err := h.service.GetConversionTable(c.Request.Context(), profileID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat tabel konversi: "+err.Error())
		return
	}
	response.Success(c, "Tabel konversi skor berhasil dimuat", tables)
}

func (h *EPHandler) SaveConversionProfile(c *gin.Context) {
	var req dto.SaveConversionProfileRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data profil konversi tidak valid: "+err.Error(), nil)
		return
	}

	p, err := h.service.SaveConversionProfile(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menyimpan profil konversi: "+err.Error())
		return
	}
	response.Created(c, "Profil konversi skor berhasil disimpan", p)
}

func (h *EPHandler) CreateExam(c *gin.Context) {
	var req dto.CreateEPExamRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data ujian tidak valid: "+err.Error(), nil)
		return
	}

	username, _ := c.Get("username")
	exam, err := h.service.CreateExam(c.Request.Context(), &req, username.(string))
	if err != nil {
		response.InternalServerError(c, "Gagal membuat ujian: "+err.Error())
		return
	}
	response.Created(c, "Ujian TOEFL/TOEIC berhasil dibuat", exam)
}

func (h *EPHandler) GetExams(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	exams, meta, err := h.service.GetExams(c.Request.Context(), page, perPage)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar ujian: "+err.Error())
		return
	}
	response.SuccessWithPagination(c, "Daftar ujian berhasil dimuat", exams, meta)
}

func (h *EPHandler) GetExamByID(c *gin.Context) {
	examID, _ := strconv.Atoi(c.Param("exam_id"))
	exam, err := h.service.GetExamByID(c.Request.Context(), examID)
	if err != nil {
		response.NotFound(c, "Ujian tidak ditemukan")
		return
	}
	response.Success(c, "Detail ujian berhasil dimuat", exam)
}

func (h *EPHandler) UpdateExam(c *gin.Context) {
	examID, _ := strconv.Atoi(c.Param("exam_id"))
	var req dto.CreateEPExamRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data ujian tidak valid: "+err.Error(), nil)
		return
	}

	exam, err := h.service.UpdateExam(c.Request.Context(), examID, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui paket ujian: "+err.Error())
		return
	}
	response.Success(c, "Paket ujian berhasil diperbarui", exam)
}

func (h *EPHandler) DeleteExam(c *gin.Context) {
	examID, _ := strconv.Atoi(c.Param("exam_id"))
	err := h.service.DeleteExam(c.Request.Context(), examID)
	if err != nil {
		response.InternalServerError(c, "Gagal menghapus paket ujian: "+err.Error())
		return
	}
	response.Success(c, "Paket ujian berhasil dihapus", nil)
}

func (h *EPHandler) CreateSchedule(c *gin.Context) {
	var req dto.CreateEPScheduleRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format jadwal tidak valid: "+err.Error(), nil)
		return
	}

	sched, err := h.service.CreateSchedule(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal membuat jadwal sesi: "+err.Error())
		return
	}
	response.Created(c, "Jadwal sesi ujian berhasil dibuat", sched)
}

func (h *EPHandler) GetSchedules(c *gin.Context) {
	examID, _ := strconv.Atoi(c.DefaultQuery("exam_id", "0"))
	schedules, err := h.service.GetSchedules(c.Request.Context(), examID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat jadwal ujian: "+err.Error())
		return
	}
	response.Success(c, "Daftar jadwal ujian berhasil dimuat", schedules)
}

func (h *EPHandler) UpdateSchedule(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	var req dto.UpdateEPScheduleRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data jadwal tidak valid: "+err.Error(), nil)
		return
	}

	sched, err := h.service.UpdateSchedule(c.Request.Context(), scheduleID, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui jadwal sesi: "+err.Error())
		return
	}
	response.Success(c, "Jadwal sesi ujian berhasil diperbarui", sched)
}

func (h *EPHandler) DeleteSchedule(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	err := h.service.DeleteSchedule(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal menghapus jadwal sesi: "+err.Error())
		return
	}
	response.Success(c, "Jadwal sesi ujian berhasil dihapus", nil)
}

func (h *EPHandler) RegisterParticipants(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	var req dto.RegisterParticipantsRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data peserta tidak valid: "+err.Error(), nil)
		return
	}

	count, err := h.service.RegisterParticipants(c.Request.Context(), scheduleID, req.ParticipantCodes, req.Participants)
	if err != nil {
		response.InternalServerError(c, "Gagal mendaftarkan peserta: "+err.Error())
		return
	}
	response.Success(c, "Peserta berhasil didaftarkan ke jadwal", map[string]int{"total_registered": count})
}

func (h *EPHandler) GetScheduleParticipants(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	list, err := h.service.GetScheduleParticipants(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat peserta jadwal: "+err.Error())
		return
	}
	response.Success(c, "Daftar peserta jadwal berhasil dimuat", list)
}

func (h *EPHandler) RemoveParticipant(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	kodepeserta := c.Param("kodepeserta")
	if scheduleID == 0 || kodepeserta == "" {
		response.BadRequest(c, "schedule_id dan kodepeserta wajib diisi", nil)
		return
	}
	err := h.service.RemoveParticipantFromSchedule(c.Request.Context(), scheduleID, kodepeserta)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Success(c, "Peserta berhasil dihapus dari sesi ujian", nil)
}

func (h *EPHandler) CalculateScores(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	scores, err := h.service.CalculateScores(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal menghitung skor ujian: "+err.Error())
		return
	}
	response.Success(c, "Kalkulasi skor konversi TOEFL/TOEIC berhasil dihitung", scores)
}

func (h *EPHandler) GetScores(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	scores, err := h.service.GetScores(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat skor ujian: "+err.Error())
		return
	}
	response.Success(c, "Rekapitulasi skor ujian berhasil dimuat", scores)
}

func (h *EPHandler) GenerateCertificates(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	certs, err := h.service.GenerateCertificates(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal menerbitkan sertifikat: "+err.Error())
		return
	}
	response.Success(c, "Sertifikat resmi berhasil diterbitkan", certs)
}

func (h *EPHandler) ExportScheduleScores(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	data, filename, err := h.service.ExportScoresToCSV(c.Request.Context(), scheduleID, 0)
	if err != nil {
		response.InternalServerError(c, "Gagal mengekspor rekapitulasi nilai jadwal: "+err.Error())
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *EPHandler) ExportExamScores(c *gin.Context) {
	examID, _ := strconv.Atoi(c.Param("exam_id"))
	data, filename, err := h.service.ExportScoresToCSV(c.Request.Context(), 0, examID)
	if err != nil {
		response.InternalServerError(c, "Gagal mengekspor rekapitulasi nilai paket ujian: "+err.Error())
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// ─────────────────────────────────────────────────────────────────────────────
// Test Taker (Participant) Handler Methods
// ─────────────────────────────────────────────────────────────────────────────
func (h *EPHandler) StartExamSession(c *gin.Context) {
	var req dto.StartEPExamRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format permintaan tidak valid: "+err.Error(), nil)
		return
	}

	pCodeVal, _ := c.Get("participant_code")
	session, err := h.service.StartExamSession(c.Request.Context(), req.ScheduleID, pCodeVal.(string), req.SessionToken)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Success(c, "Sesi ujian TOEFL/TOEIC berhasil dimulai", session)
}

func (h *EPHandler) GetCurrentSectionSession(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	pCodeVal, _ := c.Get("participant_code")

	session, err := h.service.GetCurrentSectionSession(c.Request.Context(), scheduleID, pCodeVal.(string))
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Success(c, "Seksi ujian berhasil dimuat", session)
}

func (h *EPHandler) SaveAnswer(c *gin.Context) {
	var req dto.SaveAnswerDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format jawaban tidak valid: "+err.Error(), nil)
		return
	}

	pCodeVal, _ := c.Get("participant_code")
	err := h.service.SaveAnswer(c.Request.Context(), pCodeVal.(string), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menyimpan jawaban: "+err.Error())
		return
	}
	response.Success(c, "Jawaban berhasil disimpan", nil)
}

func (h *EPHandler) AdvanceToNextSection(c *gin.Context) {
	var req dto.NextSectionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Permintaan seksi berikutnya tidak valid: "+err.Error(), nil)
		return
	}

	pCodeVal, _ := c.Get("participant_code")
	session, err := h.service.AdvanceToNextSection(c.Request.Context(), req.ScheduleID, pCodeVal.(string))
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Success(c, "Berhasil beralih ke seksi berikutnya", session)
}

func (h *EPHandler) FinishExam(c *gin.Context) {
	var req dto.FinishEPExamDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Permintaan submit ujian tidak valid: "+err.Error(), nil)
		return
	}

	pCodeVal, _ := c.Get("participant_code")
	err := h.service.FinishExam(c.Request.Context(), req.ScheduleID, pCodeVal.(string), req.SubmitReason)
	if err != nil {
		response.InternalServerError(c, "Gagal menyelesaikan ujian: "+err.Error())
		return
	}
	response.Success(c, "Ujian TOEFL/TOEIC telah berhasil diselesaikan", nil)
}

func (h *EPHandler) GetMyResults(c *gin.Context) {
	pCodeVal, exists := c.Get("participant_code")
	if !exists || pCodeVal == nil || pCodeVal.(string) == "" {
		response.Unauthorized(c, "Sesi autentikasi peserta tidak valid")
		return
	}
	participantCode := pCodeVal.(string)
	results, err := h.service.GetParticipantScores(c.Request.Context(), participantCode)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat rekap nilai dan sertifikat: "+err.Error())
		return
	}
	response.Success(c, "Data hasil ujian dan sertifikat berhasil dimuat", results)
}

// ─────────────────────────────────────────────────────────────────────────────
// Public Certificate Verification
// ─────────────────────────────────────────────────────────────────────────────
func (h *EPHandler) VerifyCertificate(c *gin.Context) {
	code := c.Param("verification_code")
	verif, err := h.service.VerifyCertificate(c.Request.Context(), code)
	if err != nil {
		response.NotFound(c, "Sertifikat tidak ditemukan atau kode verifikasi tidak valid")
		return
	}
	response.Success(c, "Data verifikasi sertifikat resmi ditemukan", verif)
}

// ─────────────────────────────────────────────────────────────────────────────
// Question Banks
// ─────────────────────────────────────────────────────────────────────────────
func (h *EPHandler) GetEPQuestionBanks(c *gin.Context) {
	examTypeID, _ := strconv.Atoi(c.DefaultQuery("exam_type_id", "0"))
	sectionID, _ := strconv.Atoi(c.DefaultQuery("section_id", "0"))
	search := c.DefaultQuery("search", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	banks, pagination, err := h.service.GetEPQuestionBanks(c.Request.Context(), examTypeID, sectionID, search, page, perPage)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar bank soal: "+err.Error())
		return
	}
	response.SuccessWithPagination(c, "Daftar bank soal TOEFL/TOEIC berhasil dimuat", banks, pagination)
}

func (h *EPHandler) CreateEPQuestionBank(c *gin.Context) {
	var req dto.CreateEPBankDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data bank soal tidak valid: "+err.Error(), nil)
		return
	}

	bank, err := h.service.CreateEPQuestionBank(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal membuat bank soal baru: "+err.Error())
		return
	}
	response.Created(c, "Bank soal baru berhasil dibuat", bank)
}

func (h *EPHandler) GetEPQuestionBankByCode(c *gin.Context) {
	kodesoal := c.Param("kodesoal")
	bank, err := h.service.GetEPQuestionBankByCode(c.Request.Context(), kodesoal)
	if err != nil {
		response.NotFound(c, "Bank soal tidak ditemukan: "+err.Error())
		return
	}
	response.Success(c, "Detail bank soal berhasil dimuat", bank)
}

func (h *EPHandler) UpdateEPQuestionBank(c *gin.Context) {
	kodesoal := c.Param("kodesoal")
	var req dto.UpdateEPBankDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data edit bank soal tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.UpdateEPQuestionBank(c.Request.Context(), kodesoal, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui bank soal: "+err.Error())
		return
	}
	response.Success(c, "Bank soal berhasil diperbarui", nil)
}

func (h *EPHandler) DeleteEPQuestionBank(c *gin.Context) {
	kodesoal := c.Param("kodesoal")
	err := h.service.DeleteEPQuestionBank(c.Request.Context(), kodesoal)
	if err != nil {
		response.InternalServerError(c, "Gagal menghapus bank soal: "+err.Error())
		return
	}
	response.Success(c, "Bank soal berhasil dihapus", nil)
}

func (h *EPHandler) CopyEPQuestionBank(c *gin.Context) {
	kodesoal := c.Param("kodesoal")
	var req dto.CopyEPBankDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data salin bank soal tidak valid: "+err.Error(), nil)
		return
	}

	bank, err := h.service.CopyEPQuestionBank(c.Request.Context(), &req, kodesoal)
	if err != nil {
		response.InternalServerError(c, "Gagal menyalin bank soal: "+err.Error())
		return
	}
	response.Created(c, "Bank soal berhasil diduplikasi", bank)
}

func (h *EPHandler) GetBankQuestions(c *gin.Context) {
	kodesoal := c.Param("kodesoal")
	stimuli, err := h.service.GetStimuliByBank(c.Request.Context(), kodesoal)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat bank soal: "+err.Error())
		return
	}
	response.Success(c, "Daftar naskah stimulus dan butir soal berhasil dimuat", stimuli)
}

func (h *EPHandler) CreateBankStimulus(c *gin.Context) {
	var req dto.CreateStimulusRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data stimulus tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.CreateStimulusWithItems(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menyimpan stimulus dan butir soal: "+err.Error())
		return
	}
	response.Created(c, "Stimulus dan butir soal berhasil ditambahkan ke bank soal", nil)
}

func (h *EPHandler) UpdateBankStimulus(c *gin.Context) {
	idStimulus, _ := strconv.ParseInt(c.Param("id_stimulus"), 10, 64)
	var req dto.UpdateStimulusRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data edit stimulus tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.UpdateStimulus(c.Request.Context(), idStimulus, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui stimulus: "+err.Error())
		return
	}
	response.Success(c, "Stimulus berhasil diperbarui", nil)
}

func (h *EPHandler) DeleteBankStimulus(c *gin.Context) {
	idStimulus, _ := strconv.ParseInt(c.Param("id_stimulus"), 10, 64)
	err := h.service.DeleteStimulus(c.Request.Context(), idStimulus)
	if err != nil {
		response.InternalServerError(c, "Gagal menghapus stimulus: "+err.Error())
		return
	}
	response.Success(c, "Stimulus berhasil dihapus", nil)
}

func (h *EPHandler) CreateSubItem(c *gin.Context) {
	idStimulus, _ := strconv.ParseInt(c.Param("id_stimulus"), 10, 64)
	var req dto.CreateSubItemRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data butir soal tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.CreateSubItem(c.Request.Context(), idStimulus, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menambahkan butir soal: "+err.Error())
		return
	}
	response.Created(c, "Butir soal berhasil ditambahkan ke stimulus", nil)
}

func (h *EPHandler) UpdateSubItem(c *gin.Context) {
	idItem, _ := strconv.ParseInt(c.Param("id_item"), 10, 64)
	var req dto.UpdateSubItemRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data butir soal tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.UpdateSubItem(c.Request.Context(), idItem, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui butir soal: "+err.Error())
		return
	}
	response.Success(c, "Butir soal berhasil diperbarui", nil)
}

func (h *EPHandler) DeleteSubItem(c *gin.Context) {
	idItem, _ := strconv.ParseInt(c.Param("id_item"), 10, 64)
	err := h.service.DeleteSubItem(c.Request.Context(), idItem)
	if err != nil {
		response.InternalServerError(c, "Gagal menghapus butir soal: "+err.Error())
		return
	}
	response.Success(c, "Butir soal berhasil dihapus", nil)
}

func (h *EPHandler) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File tidak ditemukan dalam request form: "+err.Error(), nil)
		return
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		// Audio formats
		".mp3": true, ".wav": true, ".ogg": true, ".m4a": true, ".aac": true, ".flac": true,
		// Image formats
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true,
		// Video formats
		".mp4": true, ".webm": true, ".mov": true, ".mkv": true,
	}

	if !allowed[ext] {
		response.BadRequest(c, fmt.Sprintf("Format file '%s' tidak didukung. Format didukung: Audio (MP3, WAV, OGG, M4A), Gambar (JPG, PNG, WEBP, SVG), Video (MP4, WEBM)", ext), nil)
		return
	}

	// 50MB maximum size
	if file.Size > 50*1024*1024 {
		response.BadRequest(c, "Ukuran file terlalu besar (maksimal 50MB)", nil)
		return
	}

	uploadDir := "./uploads/media"
	_ = os.MkdirAll(uploadDir, 0755)

	cleanBase := strings.TrimSuffix(filepath.Base(file.Filename), ext)
	cleanBase = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, cleanBase)
	if len(cleanBase) > 35 {
		cleanBase = cleanBase[:35]
	}

	newFileName := fmt.Sprintf("%d_%s_%s%s", time.Now().UnixNano()/1e6, uuid.New().String()[:8], cleanBase, ext)
	dst := filepath.Join(uploadDir, newFileName)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.InternalServerError(c, "Gagal menyimpan file di server: "+err.Error())
		return
	}

	fileURL := fmt.Sprintf("/uploads/media/%s", newFileName)

	response.Success(c, "File multimedia berhasil diunggah", gin.H{
		"file_url":      fileURL,
		"original_name": file.Filename,
		"file_size":     file.Size,
		"file_ext":      ext,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Anti-Cheat, Telemetry & Proctor Live Action Handlers
// ─────────────────────────────────────────────────────────────────────────────

func (h *EPHandler) RecordSecurityEvent(c *gin.Context) {
	var req dto.EPSecurityEventRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data event keamanan tidak valid: "+err.Error(), nil)
		return
	}

	participantCode := c.GetString("participant_code")
	if participantCode == "" {
		participantCode = c.GetString("kodepeserta")
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	res, err := h.service.RecordSecurityEvent(c.Request.Context(), req.ScheduleID, participantCode, &req, ip, userAgent)
	if err != nil {
		response.InternalServerError(c, "Gagal mencatat event keamanan: "+err.Error())
		return
	}
	response.Success(c, "Event keamanan berhasil dicatat", res)
}

func (h *EPHandler) SyncParticipantHeartbeat(c *gin.Context) {
	var req dto.EPHeartbeatRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data heartbeat tidak valid: "+err.Error(), nil)
		return
	}

	participantCode := c.GetString("participant_code")
	if participantCode == "" {
		participantCode = c.GetString("kodepeserta")
	}

	res, err := h.service.SyncParticipantHeartbeat(c.Request.Context(), req.ScheduleID, participantCode, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal sinkronisasi heartbeat: "+err.Error())
		return
	}
	response.Success(c, "Heartbeat berhasil disinkronkan", res)
}

func (h *EPHandler) ApplyProctorAction(c *gin.Context) {
	var req dto.EPProctorActionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data aksi pengawas tidak valid: "+err.Error(), nil)
		return
	}

	if req.ParticipantCode == "" {
		req.ParticipantCode = req.KodePeserta
	}
	if req.ParticipantCode == "" {
		response.BadRequest(c, "Kode peserta wajib diisi (participant_code / kodepeserta)", nil)
		return
	}

	err := h.service.ApplyProctorAction(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menerapkan aksi pengawas: "+err.Error())
		return
	}
	response.Success(c, "Aksi pengawas berhasil diterapkan", nil)
}

func (h *EPHandler) GetSecurityEventsByParticipant(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	participantCode := c.Param("kodepeserta")
	if participantCode == "all" {
		participantCode = ""
	}

	events, err := h.service.GetSecurityEventsByParticipant(c.Request.Context(), scheduleID, participantCode)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat jejak audit pelanggaran: "+err.Error())
		return
	}
	response.Success(c, "Jejak audit pelanggaran berhasil dimuat", events)
}

func (h *EPHandler) UploadProctorSnapshot(c *gin.Context) {
	participantCode := c.GetString("participant_code")
	if participantCode == "" {
		response.Unauthorized(c, "Sesi peserta tidak valid")
		return
	}

	var req dto.UploadProctorSnapshotRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data snapshot tidak valid: "+err.Error(), nil)
		return
	}

	res, err := h.service.SaveProctorSnapshot(c.Request.Context(), req.ScheduleID, participantCode, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menyimpan snapshot pengawas: "+err.Error())
		return
	}

	response.Success(c, "Snapshot pengawas berhasil disimpan", res)
}

func (h *EPHandler) GetProctorSnapshots(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	participantCode := c.Query("kodepeserta")
	snapshotType := c.DefaultQuery("type", "ALL")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	snapshots, err := h.service.GetProctorSnapshots(c.Request.Context(), scheduleID, participantCode, snapshotType, limit, offset)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat galeri snapshot pengawas: "+err.Error())
		return
	}

	response.Success(c, "Galeri snapshot pengawas berhasil dimuat", snapshots)
}

func (h *EPHandler) GetLatestProctorGrid(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))

	grid, err := h.service.GetLatestProctorGrid(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat matriks snapshot pengawas: "+err.Error())
		return
	}

	response.Success(c, "Matriks snapshot pengawas berhasil dimuat", grid)
}

func (h *EPHandler) DownloadCSVTemplate(c *gin.Context) {
	data := h.service.GetCSVTemplate()
	c.Header("Content-Disposition", "attachment; filename=toefl_question_bank_template.csv")
	c.Data(200, "text/csv; charset=utf-8", data)
}

func (h *EPHandler) ExportBankQuestions(c *gin.Context) {
	kodesoal := c.Param("kodesoal")
	data, err := h.service.ExportBankQuestionsToCSV(c.Request.Context(), kodesoal)
	if err != nil {
		response.InternalServerError(c, "Gagal mengunduh soal CSV: "+err.Error())
		return
	}
	filename := fmt.Sprintf("bank_soal_%s.csv", kodesoal)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(200, "text/csv; charset=utf-8", data)
}

func (h *EPHandler) ImportBankQuestions(c *gin.Context) {
	kodesoal := c.Param("kodesoal")
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File CSV wajib diunggah: "+err.Error(), nil)
		return
	}

	src, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "Gagal membuka file: "+err.Error())
		return
	}
	defer src.Close()

	res, err := h.service.ImportBankQuestionsFromCSV(c.Request.Context(), kodesoal, src)
	if err != nil {
		response.BadRequest(c, "Gagal mengimpor soal: "+err.Error(), nil)
		return
	}

	response.Success(c, fmt.Sprintf("Berhasil mengimpor %d stimulus dan %d butir soal", res.StimulusCount, res.QuestionCount), res)
}

func (h *EPHandler) GetExternalParticipants(c *gin.Context) {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	list, pag, err := h.service.GetExternalParticipants(c.Request.Context(), search, page, perPage)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat peserta umum: "+err.Error())
		return
	}
	response.SuccessWithPagination(c, "Berhasil memuat peserta umum", list, pag)
}

func (h *EPHandler) GetNextExternalParticipantCode(c *gin.Context) {
	code, err := h.service.GetNextExternalParticipantCode(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal men-generate kode peserta: "+err.Error())
		return
	}
	response.Success(c, "Kode peserta berhasil digenerate", map[string]string{
		"next_code": code,
	})
}

func (h *EPHandler) GetExternalParticipantByCode(c *gin.Context) {
	code := c.Param("code")
	item, err := h.service.GetExternalParticipantByCode(c.Request.Context(), code)
	if err != nil {
		response.NotFound(c, "Peserta umum tidak ditemukan")
		return
	}
	response.Success(c, "Detail peserta umum", item)
}

func (h *EPHandler) CreateExternalParticipant(c *gin.Context) {
	var req dto.CreateEPExternalParticipantRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data peserta umum tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.CreateExternalParticipant(c.Request.Context(), &req); err != nil {
		response.InternalServerError(c, "Gagal menyimpan peserta umum: "+err.Error())
		return
	}
	response.Created(c, "Peserta umum berhasil ditambahkan", nil)
}

func (h *EPHandler) UpdateExternalParticipant(c *gin.Context) {
	code := c.Param("code")
	var req dto.UpdateEPExternalParticipantRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data peserta umum tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.UpdateExternalParticipant(c.Request.Context(), code, &req); err != nil {
		response.InternalServerError(c, "Gagal memperbarui peserta umum: "+err.Error())
		return
	}
	response.Success(c, "Peserta umum berhasil diperbarui", nil)
}

func (h *EPHandler) DeleteExternalParticipant(c *gin.Context) {
	code := c.Param("code")
	if err := h.service.DeleteExternalParticipant(c.Request.Context(), code); err != nil {
		response.InternalServerError(c, "Gagal menghapus peserta umum: "+err.Error())
		return
	}
	response.Success(c, "Peserta umum berhasil dinonaktifkan", nil)
}

func (h *EPHandler) ImportExternalParticipants(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File CSV wajib diunggah: "+err.Error(), nil)
		return
	}
	src, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "Gagal membuka file: "+err.Error())
		return
	}
	defer src.Close()

	count, err := h.service.ImportExternalParticipantsFromCSV(c.Request.Context(), src)
	if err != nil {
		response.BadRequest(c, "Gagal mengimpor peserta umum: "+err.Error(), nil)
		return
	}
	response.Success(c, fmt.Sprintf("Berhasil mengimpor %d peserta umum", count), gin.H{"count": count})
}

