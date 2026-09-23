package handler

import (
	"fmt"
	"strconv"
	"strings"

	"poltekkes-cat-backend/internal/modules/examination/dto"
	"poltekkes-cat-backend/internal/modules/examination/service"
	"poltekkes-cat-backend/internal/shared/middleware"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type ExaminationHandler struct {
	service service.ExaminationService
}

func NewExaminationHandler(service service.ExaminationService) *ExaminationHandler {
	return &ExaminationHandler{service: service}
}

func (h *ExaminationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// Admin Exam & Proctoring Routes
	admin := rg.Group("")
	admin.Use(middleware.AdminAuthMiddleware("admin", "cat"))
	{
		admin.GET("", h.GetExams)
		admin.GET("/:exam_id", h.GetExamByID)
		admin.POST("", h.CreateExam)
		admin.PUT("/:exam_id", h.UpdateExam)
		admin.DELETE("/:exam_id", h.DeleteExam)
		admin.GET("/schedules", h.GetSchedules)
		admin.POST("/schedules", h.CreateSchedule)
		admin.GET("/participants", h.GetExamParticipants)
		admin.GET("/:exam_id/participants", h.GetExamParticipants)
		admin.GET("/monitoring/:schedule_id", h.GetLiveMonitoring)
		admin.GET("/monitoring/:schedule_id/events", h.GetRecentSecurityEvents)
		admin.POST("/proctor/action", h.ApplyProctorAction)
		admin.GET("/collusion/:schedule_id", h.GetCollusionAnalysis)
		admin.GET("/:exam_id/completed-participants", h.GetCompletedParticipants)
		admin.GET("/completed-participants", h.GetCompletedParticipants)
		admin.GET("/participant-result/:schedule_id/:participant_code", h.GetParticipantExamResultDetail)

		// === Sesi / Jadwal Ujian CRUD ===
		admin.GET("/:exam_id/sessions", h.GetSessions)
		admin.POST("/:exam_id/sessions", h.CreateSession)
		admin.PUT("/sessions/:session_id", h.UpdateSession)
		admin.DELETE("/sessions/:session_id", h.DeleteSession)
		admin.POST("/sessions/:session_id/refresh-token", h.RefreshToken)

		// === Ruang per Sesi CRUD ===
		admin.GET("/sessions/:session_id/rooms", h.GetSessionRooms)
		admin.POST("/sessions/:session_id/rooms", h.AddRoomToSession)
		admin.PUT("/sessions/rooms/:room_session_id", h.UpdateSessionRoom)
		admin.DELETE("/sessions/rooms/:room_session_id", h.DeleteSessionRoom)

		// === Bank Soal per Sesi ===
		admin.GET("/sessions/:session_id/bank-soal", h.GetSessionBankSoal)
		admin.POST("/sessions/:session_id/bank-soal", h.AddBankSoalToSession)
		admin.DELETE("/sessions/:session_id/bank-soal/:kodesoal", h.RemoveBankSoalFromSession)
		admin.DELETE("/sessions/:session_id/bank-soal", h.RemoveBankSoalFromSession)

		// === Peserta Ujian CRUD ===
		admin.POST("/:exam_id/participants", h.AddExamParticipant)
		admin.DELETE("/:exam_id/participants/:participant_code", h.RemoveExamParticipant)
		admin.GET("/:exam_id/available-participants", h.GetAvailableParticipants)
		admin.POST("/:exam_id/sipenmaru/import", h.ImportSipenmaruToExam)

		// === Plotting & Distribusi Peserta & Ruang ===
		admin.POST("/:exam_id/auto-distribute", h.AutoDistributeParticipants)
		admin.GET("/sessions/rooms/:room_session_id/participants", h.GetRoomParticipants)
		admin.POST("/sessions/rooms/:room_session_id/participants", h.AddRoomParticipants)
		admin.DELETE("/sessions/rooms/:room_session_id/participants/:participant_code", h.RemoveRoomParticipant)

		// === Barcode Attendance Verification ===
		admin.POST("/proctor/verify-barcode", h.VerifyProctorBarcode)
		admin.GET("/sessions/:session_id/attendance-summary", h.GetSessionAttendanceSummary)
		admin.PUT("/proctor/participants/:participant_code/verify", h.SetParticipantVerificationStatus)
	}

	// Participant Exam Taking & Security Routes
	participant := rg.Group("/participant")
	participant.Use(middleware.ParticipantAuthMiddleware())
	{
		participant.GET("/schedules", h.GetParticipantSchedules)
		participant.POST("/verify-barcode", h.VerifyParticipantBarcode)
		participant.POST("/start", h.StartExam)
		participant.GET("/:schedule_id/questions", h.GetSessionQuestions)
		participant.GET("/questions", h.GetSessionQuestions)
		participant.POST("/answers/save", h.SaveAnswer)
		participant.POST("/heartbeat", h.Heartbeat)
		participant.POST("/events", h.RecordSecurityEvent)
		participant.POST("/:schedule_id/finish", h.FinishExam)

		// Dynamic Multi-Item & Multimedia Routes
		participant.GET("/:schedule_id/dynamic-session", h.GetDynamicExamSession)
		participant.POST("/dynamic-answers/save", h.SaveDynamicAnswer)
		participant.POST("/:schedule_id/dynamic-finish", h.FinishDynamicExam)

		// Laporan Hasil Ujian Peserta
		participant.GET("/results", h.GetMyExamResults)
		participant.GET("/results/:schedule_id", h.GetMyExamResultDetail)
	}

	// Direct alias for POST /events
	rg.POST("/events", middleware.ParticipantAuthMiddleware(), h.RecordSecurityEvent)
	rg.POST("/heartbeat", middleware.ParticipantAuthMiddleware(), h.Heartbeat)
}

func (h *ExaminationHandler) GetExams(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	periodID, _ := strconv.Atoi(c.Query("period_id"))
	if periodID == 0 {
		periodID, _ = strconv.Atoi(c.Query("idperiode"))
	}
	search := c.Query("search")

	exams, pagination, err := h.service.GetExams(c.Request.Context(), page, perPage, periodID, search)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat data ujian: "+err.Error())
		return
	}

	response.SuccessWithPagination(c, "Daftar ujian berhasil dimuat", exams, pagination)
}

func (h *ExaminationHandler) GetExamByID(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}

	exam, err := h.service.GetExamByID(c.Request.Context(), examID)
	if err != nil {
		response.NotFound(c, "Data ujian tidak ditemukan: "+err.Error())
		return
	}

	response.Success(c, "Data ujian berhasil dimuat", exam)
}

func (h *ExaminationHandler) GetSchedules(c *gin.Context) {
	examID, _ := strconv.Atoi(c.Query("exam_id"))
	periodID, _ := strconv.Atoi(c.Query("period_id"))
	if periodID == 0 {
		periodID, _ = strconv.Atoi(c.Query("idperiode"))
	}
	schedules, err := h.service.GetSchedules(c.Request.Context(), examID, periodID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat jadwal ujian: "+err.Error())
		return
	}

	response.Success(c, "Daftar jadwal sesi ujian berhasil dimuat", schedules)
}

func (h *ExaminationHandler) GetExamParticipants(c *gin.Context) {
	examID, _ := strconv.Atoi(c.Param("exam_id"))
	if examID == 0 {
		examID, _ = strconv.Atoi(c.Query("exam_id"))
	}
	participants, err := h.service.GetExamParticipants(c.Request.Context(), examID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar peserta ujian: "+err.Error())
		return
	}

	response.Success(c, "Daftar peserta ujian berhasil dimuat", participants)
}

func (h *ExaminationHandler) GetLiveMonitoring(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	data, err := h.service.GetLiveMonitoring(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat monitoring: "+err.Error())
		return
	}

	response.Success(c, "Data monitoring realtime berhasil diperbarui", data)
}

func (h *ExaminationHandler) StartExam(c *gin.Context) {
	var req dto.StartExamRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "ID jadwal ujian wajib disertakan", nil)
		return
	}

	participantCode := c.GetString("participant_code")
	res, err := h.service.StartExam(c.Request.Context(), participantCode, &req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Sesi ujian berhasil dimulai", res)
}

func (h *ExaminationHandler) SaveAnswer(c *gin.Context) {
	var req dto.SaveAnswerRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format lembar jawaban tidak valid", nil)
		return
	}

	participantCode := c.GetString("participant_code")
	res, err := h.service.SaveAnswer(c.Request.Context(), participantCode, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menyimpan jawaban: "+err.Error())
		return
	}

	response.Success(c, "Jawaban berhasil disimpan", res)
}

func (h *ExaminationHandler) Heartbeat(c *gin.Context) {
	var req dto.HeartbeatRequestDTO
	_ = c.ShouldBindJSON(&req)

	participantCode := c.GetString("participant_code")
	res, err := h.service.Heartbeat(c.Request.Context(), participantCode, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memproses detak jantung server: "+err.Error())
		return
	}
	response.Success(c, "Heartbeat OK", res)
}

func (h *ExaminationHandler) RecordSecurityEvent(c *gin.Context) {
	var req dto.SecurityEventRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format security event tidak valid", nil)
		return
	}

	participantCode := c.GetString("participant_code")
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	res, err := h.service.RecordSecurityEvent(c.Request.Context(), participantCode, clientIP, userAgent, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal mencatat event keamanan: "+err.Error())
		return
	}

	response.Success(c, "Event keamanan berhasil dicatat", res)
}

func (h *ExaminationHandler) GetRecentSecurityEvents(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	events, err := h.service.GetRecentSecurityEvents(c.Request.Context(), scheduleID, limit)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat log pelanggaran: "+err.Error())
		return
	}

	response.Success(c, "Log pelanggaran keamanan berhasil dimuat", events)
}

func (h *ExaminationHandler) ApplyProctorAction(c *gin.Context) {
	var req dto.ProctorActionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data aksi pengawas tidak valid", nil)
		return
	}

	err := h.service.ApplyProctorAction(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menjalankan aksi pengawas: "+err.Error())
		return
	}

	response.Success(c, "Aksi pengawas berhasil diterapkan", req)
}

func (h *ExaminationHandler) FinishExam(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	participantCode := c.GetString("participant_code")

	var req dto.FinishExamRequestDTO
	_ = c.ShouldBindJSON(&req)

	submitReason := req.SubmitReason
	if submitReason == "" {
		submitReason = "MANUAL_SUBMIT"
	}

	res, err := h.service.FinishExam(c.Request.Context(), participantCode, scheduleID, submitReason)
	if err != nil {
		response.InternalServerError(c, "Gagal menyelesaikan ujian: "+err.Error())
		return
	}

	response.Success(c, "Ujian telah selesai dikerjakan dan dinilai", res)
}

func (h *ExaminationHandler) CreateExam(c *gin.Context) {
	var req dto.CreateExamRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data ujian tidak valid: "+err.Error(), nil)
		return
	}

	id, err := h.service.CreateExam(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menyimpan data ujian: "+err.Error())
		return
	}

	response.Success(c, "Data ujian baru berhasil disimpan", map[string]interface{}{"exam_id": id})
}

func (h *ExaminationHandler) UpdateExam(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}
	var req dto.UpdateExamRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data ujian tidak valid: "+err.Error(), nil)
		return
	}

	if err := h.service.UpdateExam(c.Request.Context(), examID, &req); err != nil {
		response.InternalServerError(c, "Gagal memperbarui data ujian: "+err.Error())
		return
	}

	response.Success(c, "Data ujian berhasil diperbarui", nil)
}

func (h *ExaminationHandler) DeleteExam(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}

	if err := h.service.DeleteExam(c.Request.Context(), examID); err != nil {
		response.InternalServerError(c, "Gagal menghapus data ujian: "+err.Error())
		return
	}

	response.Success(c, "Data ujian berhasil dihapus", nil)
}

func (h *ExaminationHandler) CreateSchedule(c *gin.Context) {
	var req dto.CreateExamRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data tidak valid", nil)
		return
	}
	examID, _ := strconv.Atoi(c.Param("exam_id"))
	_ = examID
	response.Success(c, "CreateSchedule placeholder", nil)
}

// =====================================================================
// Admin CRUD: Sesi / Jadwal Ujian
// =====================================================================

func (h *ExaminationHandler) GetSessions(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}
	data, err := h.service.GetSessionsByExam(c.Request.Context(), examID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat sesi ujian: "+err.Error())
		return
	}
	response.Success(c, "Daftar sesi ujian berhasil dimuat", data)
}

func (h *ExaminationHandler) CreateSession(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}
	var req dto.CreateScheduleRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data sesi tidak valid: "+err.Error(), nil)
		return
	}
	req.ExamID = examID
	newID, err := h.service.CreateSession(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal membuat sesi ujian: "+err.Error())
		return
	}
	response.Success(c, "Sesi ujian berhasil ditambahkan", map[string]int{"session_id": newID})
}

func (h *ExaminationHandler) UpdateSession(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil || sessionID == 0 {
		response.BadRequest(c, "session_id tidak valid", nil)
		return
	}
	var req dto.UpdateSessionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data sesi tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.UpdateSession(c.Request.Context(), sessionID, &req); err != nil {
		response.InternalServerError(c, "Gagal memperbarui sesi ujian: "+err.Error())
		return
	}
	response.Success(c, "Sesi ujian berhasil diperbarui", nil)
}

func (h *ExaminationHandler) DeleteSession(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil || sessionID == 0 {
		response.BadRequest(c, "session_id tidak valid", nil)
		return
	}
	if err := h.service.DeleteSession(c.Request.Context(), sessionID); err != nil {
		response.InternalServerError(c, "Gagal menghapus sesi ujian: "+err.Error())
		return
	}
	response.Success(c, "Sesi ujian berhasil dihapus", nil)
}

func (h *ExaminationHandler) RefreshToken(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil || sessionID == 0 {
		response.BadRequest(c, "session_id tidak valid", nil)
		return
	}
	res, err := h.service.RefreshToken(c.Request.Context(), sessionID)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui token ujian: "+err.Error())
		return
	}
	response.Success(c, "Token ujian berhasil diperbarui", res)
}

// =====================================================================
// Admin CRUD: Ruang per Sesi
// =====================================================================

func (h *ExaminationHandler) GetSessionRooms(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil || sessionID == 0 {
		response.BadRequest(c, "session_id tidak valid", nil)
		return
	}
	data, err := h.service.GetSessionRooms(c.Request.Context(), sessionID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat ruang sesi: "+err.Error())
		return
	}
	response.Success(c, "Ruang sesi berhasil dimuat", data)
}

func (h *ExaminationHandler) AddRoomToSession(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil || sessionID == 0 {
		response.BadRequest(c, "session_id tidak valid", nil)
		return
	}
	var req dto.CreateRoomSessionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data ruang tidak valid: "+err.Error(), nil)
		return
	}
	req.ScheduleID = sessionID
	newID, err := h.service.AddRoomToSession(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menambah ruang ke sesi: "+err.Error())
		return
	}
	response.Success(c, "Ruang berhasil ditambahkan ke sesi", map[string]int{"room_session_id": newID})
}

func (h *ExaminationHandler) UpdateSessionRoom(c *gin.Context) {
	roomSessionID, err := strconv.Atoi(c.Param("room_session_id"))
	if err != nil || roomSessionID == 0 {
		response.BadRequest(c, "room_session_id tidak valid", nil)
		return
	}
	var req dto.UpdateRoomSessionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data ruang tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.UpdateSessionRoom(c.Request.Context(), roomSessionID, &req); err != nil {
		response.InternalServerError(c, "Gagal memperbarui ruang sesi: "+err.Error())
		return
	}
	response.Success(c, "Ruang sesi berhasil diperbarui", nil)
}

func (h *ExaminationHandler) DeleteSessionRoom(c *gin.Context) {
	roomSessionID, err := strconv.Atoi(c.Param("room_session_id"))
	if err != nil || roomSessionID == 0 {
		response.BadRequest(c, "room_session_id tidak valid", nil)
		return
	}
	if err := h.service.DeleteSessionRoom(c.Request.Context(), roomSessionID); err != nil {
		response.InternalServerError(c, "Gagal menghapus ruang dari sesi: "+err.Error())
		return
	}
	response.Success(c, "Ruang berhasil dihapus dari sesi", nil)
}

// =====================================================================
// Admin CRUD: Peserta Ujian
// =====================================================================

func (h *ExaminationHandler) AddExamParticipant(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}
	var req dto.AddExamParticipantRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data peserta tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.AddExamParticipant(c.Request.Context(), examID, req.ParticipantCode); err != nil {
		response.InternalServerError(c, "Gagal menambahkan peserta ke ujian: "+err.Error())
		return
	}
	response.Success(c, "Peserta berhasil ditambahkan ke ujian", nil)
}

func (h *ExaminationHandler) RemoveExamParticipant(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}
	participantCode := c.Param("participant_code")
	if participantCode == "" {
		response.BadRequest(c, "participant_code tidak valid", nil)
		return
	}
	if err := h.service.RemoveExamParticipant(c.Request.Context(), examID, participantCode); err != nil {
		response.InternalServerError(c, "Gagal menghapus peserta dari ujian: "+err.Error())
		return
	}
	response.Success(c, "Peserta berhasil dihapus dari ujian", nil)
}

func (h *ExaminationHandler) GetAvailableParticipants(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}
	data, err := h.service.GetAvailableParticipants(c.Request.Context(), examID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar peserta tersedia: "+err.Error())
		return
	}
	response.Success(c, "Daftar peserta yang tersedia berhasil dimuat", data)
}

func (h *ExaminationHandler) ImportSipenmaruToExam(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}

	var req dto.ExamImportSipenmaruRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Parameter impor tidak valid: "+err.Error(), nil)
		return
	}

	result, err := h.service.ImportSipenmaruToExam(c.Request.Context(), examID, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal mengimpor pendaftar ke ujian: "+err.Error())
		return
	}

	msg := fmt.Sprintf("Impor selesai: %d peserta terdaftar di ujian", result.ExamRegisteredCount)
	if result.SessionPlottedCount > 0 {
		msg += fmt.Sprintf(", %d peserta berhasil dialokasikan ke sesi & ruangan", result.SessionPlottedCount)
	}

	response.Success(c, msg, result)
}

func (h *ExaminationHandler) GetCollusionAnalysis(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	report, err := h.service.GetCollusionAnalysis(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memproses analisis kolusi: "+err.Error())
		return
	}
	response.Success(c, "Laporan analisis kolusi dan kesamaan jawaban berhasil diproses", report)
}

func (h *ExaminationHandler) GetSessionQuestions(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	if scheduleID == 0 {
		scheduleID, _ = strconv.Atoi(c.Query("schedule_id"))
	}
	participantCode := c.GetString("participant_code")

	snapshot, err := h.service.GetSessionQuestions(c.Request.Context(), participantCode, scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat naskah soal sesi ujian: "+err.Error())
		return
	}

	response.Success(c, "Naskah soal sesi ujian berhasil dimuat dari snapshot", snapshot)
}

func (h *ExaminationHandler) GetDynamicExamSession(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	if scheduleID == 0 {
		scheduleID, _ = strconv.Atoi(c.Query("schedule_id"))
	}
	participantCode := c.GetString("participant_code")

	dynamicSession, err := h.service.GetDynamicExamSession(c.Request.Context(), participantCode, scheduleID)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Sesi lembar ujian dinamis berhasil dimuat", dynamicSession)
}

func (h *ExaminationHandler) SaveDynamicAnswer(c *gin.Context) {
	participantCode := c.GetString("participant_code")
	var req dto.SaveMultiAnswerDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload jawaban tidak valid: "+err.Error(), nil)
		return
	}
	req.ParticipantCode = participantCode

	err := h.service.SaveDynamicAnswer(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, "Gagal menyimpan jawaban: "+err.Error(), nil)
		return
	}

	response.Success(c, "Jawaban sub-pertanyaan berhasil disimpan", req)
}

func (h *ExaminationHandler) FinishDynamicExam(c *gin.Context) {
	scheduleID, _ := strconv.Atoi(c.Param("schedule_id"))
	participantCode := c.GetString("participant_code")
	var req dto.FinishExamRequestDTO
	_ = c.ShouldBindJSON(&req)

	result, err := h.service.FinishDynamicExam(c.Request.Context(), participantCode, scheduleID, req.SubmitReason)
	if err != nil {
		response.InternalServerError(c, "Gagal menyelesaikan ujian: "+err.Error())
		return
	}

	response.Success(c, "Ujian berhasil diselesaikan dan dinilai secara otomatis", result)
}

// =====================================================================
// Admin: Plotting & Ruang Peserta Handlers
// =====================================================================

func (h *ExaminationHandler) AutoDistributeParticipants(c *gin.Context) {
	examID, err := strconv.Atoi(c.Param("exam_id"))
	if err != nil || examID == 0 {
		response.BadRequest(c, "exam_id tidak valid", nil)
		return
	}
	var req dto.AutoDistributeRequestDTO
	_ = c.ShouldBindJSON(&req)

	result, err := h.service.AutoDistributeParticipants(c.Request.Context(), examID, req.OverwriteExisting)
	if err != nil {
		response.BadRequest(c, "Gagal melakukan distribusi peserta: "+err.Error(), nil)
		return
	}
	response.Success(c, result.Message, result)
}

func (h *ExaminationHandler) GetRoomParticipants(c *gin.Context) {
	roomSessionID, err := strconv.Atoi(c.Param("room_session_id"))
	if err != nil || roomSessionID == 0 {
		response.BadRequest(c, "room_session_id tidak valid", nil)
		return
	}
	data, err := h.service.GetRoomParticipants(c.Request.Context(), roomSessionID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat peserta ruang: "+err.Error())
		return
	}
	response.Success(c, "Daftar peserta ruang berhasil dimuat", data)
}

func (h *ExaminationHandler) AddRoomParticipants(c *gin.Context) {
	roomSessionID, err := strconv.Atoi(c.Param("room_session_id"))
	if err != nil || roomSessionID == 0 {
		response.BadRequest(c, "room_session_id tidak valid", nil)
		return
	}
	var req dto.AddRoomParticipantRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.AddRoomParticipants(c.Request.Context(), roomSessionID, req.ParticipantCodes); err != nil {
		response.InternalServerError(c, "Gagal menambahkan peserta ke ruang: "+err.Error())
		return
	}
	response.Success(c, "Peserta berhasil ditambahkan ke ruang sesi", nil)
}

func (h *ExaminationHandler) RemoveRoomParticipant(c *gin.Context) {
	roomSessionID, err := strconv.Atoi(c.Param("room_session_id"))
	if err != nil || roomSessionID == 0 {
		response.BadRequest(c, "room_session_id tidak valid", nil)
		return
	}
	participantCode := c.Param("participant_code")
	if participantCode == "" {
		response.BadRequest(c, "participant_code tidak valid", nil)
		return
	}
	if err := h.service.RemoveRoomParticipant(c.Request.Context(), roomSessionID, participantCode); err != nil {
		response.InternalServerError(c, "Gagal menghapus peserta dari ruang: "+err.Error())
		return
	}
	response.Success(c, "Peserta berhasil dihapus dari ruang sesi", nil)
}

func (h *ExaminationHandler) GetParticipantSchedules(c *gin.Context) {
	participantCode := c.GetString("participant_code")
	if participantCode == "" {
		response.Unauthorized(c, "Peserta belum terautentikasi")
		return
	}

	res, err := h.service.GetParticipantSchedules(c.Request.Context(), participantCode)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat jadwal peserta: "+err.Error())
		return
	}

	response.Success(c, "Jadwal peserta berhasil dimuat", res)
}

func (h *ExaminationHandler) VerifyParticipantBarcode(c *gin.Context) {
	response.Forbidden(c, "Pemindaian barcode kehadiran hanya dapat dilakukan oleh Pengawas/Admin Ujian. Silakan tunjukkan barcode/QR kartu ujian Anda kepada Pengawas.")
}

func (h *ExaminationHandler) VerifyProctorBarcode(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		username = "PROCTOR"
	}

	var req dto.VerifyBarcodeRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload tidak valid: schedule_id dan barcode wajib diisi", nil)
		return
	}

	res, err := h.service.VerifyProctorBarcode(c.Request.Context(), username, &req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, res.Message, res)
}

func (h *ExaminationHandler) GetSessionAttendanceSummary(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil || sessionID <= 0 {
		response.BadRequest(c, "ID sesi tidak valid", nil)
		return
	}

	summary, err := h.service.GetSessionAttendanceSummary(c.Request.Context(), sessionID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, "Data ringkasan kehadiran berhasil dimuat", summary)
}

func (h *ExaminationHandler) SetParticipantVerificationStatus(c *gin.Context) {
	participantCode := c.Param("participant_code")
	if participantCode == "" {
		response.BadRequest(c, "participant_code wajib diisi", nil)
		return
	}

	username := c.GetString("username")
	if username == "" {
		username = "PROCTOR"
	}

	var req dto.ProctorManualVerifyDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload tidak valid: schedule_id wajib diisi", nil)
		return
	}

	err := h.service.SetParticipantVerificationStatus(c.Request.Context(), participantCode, req.ScheduleID, req.IsVerified, username)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui status verifikasi: "+err.Error())
		return
	}

	var msg string
	if req.IsVerified == 1 {
		msg = fmt.Sprintf("Peserta %s berhasil diverifikasi hadir secara manual", participantCode)
	} else {
		msg = fmt.Sprintf("Verifikasi kehadiran peserta %s dibatalkan", participantCode)
	}

	response.Success(c, msg, map[string]interface{}{
		"participant_code": participantCode,
		"schedule_id":      req.ScheduleID,
		"is_verified":      req.IsVerified,
	})
}

// === Bank Soal per Sesi ===

func (h *ExaminationHandler) GetSessionBankSoal(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil {
		response.BadRequest(c, "session_id tidak valid", nil)
		return
	}
	results, err := h.service.GetSessionBankSoal(c.Request.Context(), sessionID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat bank soal sesi: "+err.Error())
		return
	}
	response.Success(c, "Bank soal sesi berhasil dimuat", results)
}

func (h *ExaminationHandler) AddBankSoalToSession(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil {
		response.BadRequest(c, "session_id tidak valid", nil)
		return
	}
	var req dto.AddBankSoalToSessionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.AddBankSoalToSession(c.Request.Context(), sessionID, &req); err != nil {
		response.InternalServerError(c, "Gagal menambahkan bank soal: "+err.Error())
		return
	}
	response.Success(c, fmt.Sprintf("Bank soal %s berhasil ditambahkan ke sesi", req.KodeSoal), nil)
}

func (h *ExaminationHandler) RemoveBankSoalFromSession(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("session_id"))
	if err != nil {
		response.BadRequest(c, "session_id tidak valid", nil)
		return
	}
	kodeSoal := c.Param("kodesoal")
	if kodeSoal == "" {
		kodeSoal = c.Query("kodesoal")
	}
	if kodeSoal == "" {
		kodeSoal = c.Query("kode_soal")
	}
	if kodeSoal == "" {
		var body struct {
			KodeSoal    string `json:"kode_soal"`
			KodeSoalAlt string `json:"kodesoal"`
		}
		_ = c.ShouldBindJSON(&body)
		if body.KodeSoal != "" {
			kodeSoal = body.KodeSoal
		} else if body.KodeSoalAlt != "" {
			kodeSoal = body.KodeSoalAlt
		}
	}
	if kodeSoal == "" {
		response.BadRequest(c, "kodesoal wajib diisi", nil)
		return
	}
	if err := h.service.RemoveBankSoalFromSession(c.Request.Context(), sessionID, kodeSoal); err != nil {
		response.InternalServerError(c, "Gagal menghapus bank soal dari sesi: "+err.Error())
		return
	}
	response.Success(c, fmt.Sprintf("Bank soal %s berhasil dihapus dari sesi", kodeSoal), nil)
}

func (h *ExaminationHandler) GetCompletedParticipants(c *gin.Context) {
	examIDStr := c.Param("exam_id")
	if examIDStr == "" {
		examIDStr = c.Query("exam_id")
	}
	examID, _ := strconv.Atoi(examIDStr)
	scheduleID, _ := strconv.Atoi(c.Query("schedule_id"))

	participants, err := h.service.GetCompletedParticipants(c.Request.Context(), examID, scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat peserta ujian selesai: "+err.Error())
		return
	}

	response.Success(c, "Data peserta selesai berhasil dimuat", participants)
}

func (h *ExaminationHandler) GetParticipantExamResultDetail(c *gin.Context) {
	scheduleID, err := strconv.Atoi(c.Param("schedule_id"))
	if err != nil || scheduleID <= 0 {
		response.BadRequest(c, "ID jadwal ujian tidak valid", nil)
		return
	}

	participantCode := strings.TrimSpace(c.Param("participant_code"))
	if participantCode == "" {
		response.BadRequest(c, "Kode peserta wajib diisi", nil)
		return
	}

	detail, err := h.service.GetParticipantExamResultDetail(c.Request.Context(), scheduleID, participantCode)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat detail hasil ujian peserta: "+err.Error())
		return
	}

	response.Success(c, "Detail hasil dan integritas ujian peserta berhasil dimuat", detail)
}

func (h *ExaminationHandler) GetMyExamResults(c *gin.Context) {
	participantCode := c.GetString("participant_code")
	if participantCode == "" {
		response.Unauthorized(c, "Autentikasi peserta diperlukan")
		return
	}

	results, err := h.service.GetParticipantExamResults(c.Request.Context(), participantCode)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat riwayat hasil ujian: "+err.Error())
		return
	}

	response.Success(c, "Riwayat hasil ujian peserta berhasil dimuat", results)
}

func (h *ExaminationHandler) GetMyExamResultDetail(c *gin.Context) {
	participantCode := c.GetString("participant_code")
	if participantCode == "" {
		response.Unauthorized(c, "Autentikasi peserta diperlukan")
		return
	}

	scheduleID, err := strconv.Atoi(c.Param("schedule_id"))
	if err != nil || scheduleID == 0 {
		response.BadRequest(c, "ID jadwal ujian tidak valid", nil)
		return
	}

	detail, err := h.service.GetParticipantExamResultDetail(c.Request.Context(), scheduleID, participantCode)
	if err != nil {
		response.NotFound(c, "Detail hasil ujian tidak ditemukan: "+err.Error())
		return
	}

	response.Success(c, "Detail laporan hasil ujian berhasil dimuat", detail)
}
