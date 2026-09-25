package handler

import (
	"strconv"

	"poltekkes-cat-backend/internal/modules/integration/dto"
	"poltekkes-cat-backend/internal/modules/integration/service"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type IntegrationHandler struct {
	service service.IntegrationService
}

func NewIntegrationHandler(service service.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{service: service}
}

// =========================================================================
// Admin API Key Management Handlers
// =========================================================================

func (h *IntegrationHandler) GetAPIKeys(c *gin.Context) {
	keys, err := h.service.GetAPIKeys(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar API Key: "+err.Error())
		return
	}
	response.Success(c, "Daftar API Key integrasi berhasil dimuat", keys)
}

func (h *IntegrationHandler) CreateAPIKey(c *gin.Context) {
	var req dto.CreateAPIKeyRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data permintaan tidak valid: "+err.Error(), nil)
		return
	}

	created, err := h.service.CreateAPIKey(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal membuat API Key baru: "+err.Error())
		return
	}

	response.Created(c, "API Key baru berhasil dibuat", created)
}

func (h *IntegrationHandler) UpdateAPIKey(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.BadRequest(c, "ID API Key tidak valid", nil)
		return
	}

	var req dto.UpdateAPIKeyRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data permintaan tidak valid: "+err.Error(), nil)
		return
	}

	updated, err := h.service.UpdateAPIKey(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui API Key: "+err.Error())
		return
	}

	response.Success(c, "API Key berhasil diperbarui", updated)
}

func (h *IntegrationHandler) DeleteAPIKey(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.BadRequest(c, "ID API Key tidak valid", nil)
		return
	}

	if err := h.service.DeleteAPIKey(c.Request.Context(), id); err != nil {
		response.InternalServerError(c, "Gagal menghapus API Key: "+err.Error())
		return
	}

	response.Success(c, "API Key berhasil dihapus", nil)
}

func (h *IntegrationHandler) RegenerateAPIKey(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.BadRequest(c, "ID API Key tidak valid", nil)
		return
	}

	regenerated, err := h.service.RegenerateAPIKey(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, "Gagal men-generate ulang API Key: "+err.Error())
		return
	}

	response.Success(c, "API Key baru berhasil di-generate", regenerated)
}

func (h *IntegrationHandler) GetAccessLogs(c *gin.Context) {
	apiKeyID, _ := strconv.Atoi(c.Query("api_key_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "15"))
	search := c.Query("search")

	logs, pagination, err := h.service.GetAccessLogs(c.Request.Context(), apiKeyID, page, perPage, search)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat log akses API: "+err.Error())
		return
	}

	response.SuccessWithPagination(c, "Log akses API integrasi berhasil dimuat", logs, pagination)
}

// =========================================================================
// SPMB Integration Handlers (Protected by APIKeyAuthMiddleware)
// =========================================================================

func (h *IntegrationHandler) GetSPMBActiveExams(c *gin.Context) {
	refDate := c.Query("date")
	exams, err := h.service.GetActiveExams(c.Request.Context(), refDate)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar ujian aktif SPMB: "+err.Error())
		return
	}

	response.Success(c, "Daftar ujian aktif berhasil dimuat", exams)
}

func (h *IntegrationHandler) RegisterSPMBParticipant(c *gin.Context) {
	var req dto.SPMBRegisterRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data pendaftaran tidak valid: "+err.Error(), nil)
		return
	}

	res, err := h.service.RegisterSPMBParticipant(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal mendaftarkan peserta dari SPMB: "+err.Error())
		return
	}

	response.Created(c, "Peserta SPMB berhasil didaftarkan ke ujian CBT", res)
}

func (h *IntegrationHandler) RegisterSPMBBatchParticipants(c *gin.Context) {
	var req dto.SPMBBatchRegisterRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data pendaftaran batch tidak valid: "+err.Error(), nil)
		return
	}

	res, err := h.service.RegisterSPMBBatchParticipants(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, "Gagal memproses pendaftaran batch: "+err.Error(), nil)
		return
	}

	response.Success(c, "Pendaftaran batch peserta SPMB berhasil diproses", res)
}

func (h *IntegrationHandler) GetSPMBParticipant(c *gin.Context) {
	idPendaftar := c.Param("idpendaftar")
	if idPendaftar == "" {
		response.BadRequest(c, "Parameter idpendaftar / nomor ujian wajib disertakan", nil)
		return
	}

	res, err := h.service.GetSPMBParticipant(c.Request.Context(), idPendaftar)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, "Data peserta CBT berhasil ditemukan", res)
}

func (h *IntegrationHandler) UpdateSPMBParticipant(c *gin.Context) {
	idPendaftar := c.Param("idpendaftar")
	if idPendaftar == "" {
		response.BadRequest(c, "Parameter idpendaftar / nomor ujian wajib disertakan", nil)
		return
	}

	var req dto.SPMBUpdateParticipantDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembaharuan tidak valid: "+err.Error(), nil)
		return
	}

	res, err := h.service.UpdateSPMBParticipant(c.Request.Context(), idPendaftar, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui data peserta: "+err.Error())
		return
	}

	response.Success(c, "Data peserta CBT berhasil diperbarui", res)
}

func (h *IntegrationHandler) DeleteSPMBParticipant(c *gin.Context) {
	idPendaftar := c.Param("idpendaftar")
	if idPendaftar == "" {
		response.BadRequest(c, "Parameter idpendaftar / nomor ujian wajib disertakan", nil)
		return
	}

	if err := h.service.DeleteSPMBParticipant(c.Request.Context(), idPendaftar); err != nil {
		response.InternalServerError(c, "Gagal membatalkan pendaftaran peserta: "+err.Error())
		return
	}

	response.Success(c, "Pendaftaran peserta berhasil dibatalkan dan alokasi kursi lab telah dilepas", nil)
}

func (h *IntegrationHandler) GetUnplottedQueueSummary(c *gin.Context) {
	summary, err := h.service.GetUnplottedQueueSummary(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat ringkasan antrean peserta belum ter-plot: "+err.Error())
		return
	}

	response.Success(c, "Ringkasan antrean peserta berhasil dimuat", summary)
}
