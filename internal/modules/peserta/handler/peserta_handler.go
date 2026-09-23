package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"poltekkes-cat-backend/internal/modules/peserta/dto"
	"poltekkes-cat-backend/internal/modules/peserta/service"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type PesertaHandler struct {
	service service.PesertaService
}

func NewPesertaHandler(service service.PesertaService) *PesertaHandler {
	return &PesertaHandler{service: service}
}

func (h *PesertaHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.GetParticipants)
	rg.GET("/search", h.SearchParticipants)
	rg.GET("/:code", h.GetParticipantByCode)
	rg.POST("", h.CreateParticipant)
	rg.PUT("/:code", h.UpdateParticipant)
	rg.DELETE("/:code", h.DeleteParticipant)
	rg.POST("/reset-login", h.ResetLogin)
	rg.POST("/:code/reset-session", h.ResetSessionByCode)
	rg.PUT("/:code/change-password", h.ChangePassword)
	rg.POST("/upload-payment", h.UploadPaymentProof)

	// Sipenmaru Import Routes
	rg.GET("/sipenmaru/options", h.GetSipenmaruOptions)
	rg.POST("/sipenmaru/preview", h.PreviewSipenmaru)
	rg.POST("/sipenmaru/import", h.ImportSipenmaru)
}

func (h *PesertaHandler) GetParticipants(c *gin.Context) {
	periodID, _ := strconv.Atoi(c.DefaultQuery("period_id", "0"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")

	participants, pagination, err := h.service.GetParticipants(c.Request.Context(), periodID, page, perPage, search)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat data peserta: "+err.Error())
		return
	}

	response.SuccessWithPagination(c, "Daftar peserta ujian berhasil dimuat", participants, pagination)
}

func (h *PesertaHandler) GetParticipantByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Kode peserta wajib diisi", nil)
		return
	}

	participant, err := h.service.GetParticipantByCode(c.Request.Context(), code)
	if err != nil {
		response.NotFound(c, "Data peserta dengan kode "+code+" tidak ditemukan")
		return
	}

	response.Success(c, "Detail biodata peserta berhasil dimuat", participant)
}

func (h *PesertaHandler) CreateParticipant(c *gin.Context) {
	var req dto.SavePesertaRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data peserta tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.CreateParticipant(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menyimpan biodata peserta: "+err.Error())
		return
	}

	response.Created(c, "Biodata peserta berhasil disimpan", map[string]string{"kodepeserta": req.KodePeserta})
}

func (h *PesertaHandler) UpdateParticipant(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Kode peserta wajib diisi", nil)
		return
	}

	var req dto.SavePesertaRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data peserta tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.UpdateParticipant(c.Request.Context(), code, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui biodata peserta: "+err.Error())
		return
	}

	response.Success(c, "Biodata peserta berhasil diperbarui", map[string]string{"kodepeserta": code})
}

func (h *PesertaHandler) DeleteParticipant(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Kode peserta wajib diisi", nil)
		return
	}

	err := h.service.DeleteParticipant(c.Request.Context(), code)
	if err != nil {
		response.InternalServerError(c, "Gagal menghapus data peserta: "+err.Error())
		return
	}

	response.Success(c, "Data peserta berhasil dihapus", nil)
}

func (h *PesertaHandler) ResetLogin(c *gin.Context) {
	var req dto.ResetParticipantLoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Nomor peserta wajib diisi", nil)
		return
	}

	err := h.service.ResetLogin(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal mereset sesi login peserta: "+err.Error())
		return
	}

	response.Success(c, "Sesi login peserta berhasil direset", nil)
}

func (h *PesertaHandler) ResetSessionByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Kode peserta wajib diisi", nil)
		return
	}

	err := h.service.ResetLogin(c.Request.Context(), &dto.ResetParticipantLoginDTO{KodePeserta: code})
	if err != nil {
		response.InternalServerError(c, "Gagal mereset sesi login peserta: "+err.Error())
		return
	}

	response.Success(c, "Sesi login peserta berhasil direset", nil)
}

func (h *PesertaHandler) ChangePassword(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Kode peserta wajib diisi", nil)
		return
	}

	var req dto.ChangePasswordRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Password baru wajib diisi (minimal 4 karakter): "+err.Error(), nil)
		return
	}

	err := h.service.ChangePassword(c.Request.Context(), code, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal mengubah password peserta: "+err.Error())
		return
	}

	response.Success(c, "Password peserta berhasil diubah", map[string]string{"kodepeserta": code})
}

func (h *PesertaHandler) SearchParticipants(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		q = c.Query("search")
	}
	if len(q) < 2 {
		response.Success(c, "Hasil pencarian", []interface{}{})
		return
	}

	results, err := h.service.SearchParticipants(c.Request.Context(), q)
	if err != nil {
		response.InternalServerError(c, "Gagal mencari peserta: "+err.Error())
		return
	}

	response.Success(c, "Hasil pencarian peserta", results)
}

func (h *PesertaHandler) UploadPaymentProof(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File bukti pembayaran tidak ditemukan: "+err.Error(), nil)
		return
	}

	if file.Size > 5*1024*1024 {
		response.BadRequest(c, "Ukuran file maksimal 5MB", nil)
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".pdf" {
		response.BadRequest(c, "Format file harus berupa JPG, PNG, atau PDF", nil)
		return
	}

	uploadDir := "/app/uploads/pembayaran"
	_ = os.MkdirAll(uploadDir, 0755)

	filename := fmt.Sprintf("pembayaran_%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.InternalServerError(c, "Gagal menyimpan file: "+err.Error())
		return
	}

	fileURL := "/uploads/pembayaran/" + filename
	response.Success(c, "File bukti pembayaran berhasil diupload", map[string]string{
		"filename": filename,
		"url":      fileURL,
	})
}

func (h *PesertaHandler) GetSipenmaruOptions(c *gin.Context) {
	opts, err := h.service.GetSipenmaruFilterOptions(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat opsi filter Sipenmaru: "+err.Error())
		return
	}
	response.Success(c, "Opsi filter pendaftaran Sipenmaru berhasil dimuat", opts)
}

func (h *PesertaHandler) PreviewSipenmaru(c *gin.Context) {
	var req dto.SipenmaruPreviewRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Parameter filter tidak valid: "+err.Error(), nil)
		return
	}

	result, err := h.service.PreviewSipenmaruCandidates(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat preview pendaftar: "+err.Error())
		return
	}

	response.Success(c, "Preview pendaftar Sipenmaru berhasil dimuat", result)
}

func (h *PesertaHandler) ImportSipenmaru(c *gin.Context) {
	var req dto.ImportSipenmaruRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Parameter impor tidak valid: "+err.Error(), nil)
		return
	}

	if req.CBTPeriodID <= 0 {
		response.BadRequest(c, "Periode CBT tujuan (cbt_period_id) wajib dipilih", nil)
		return
	}

	result, err := h.service.ExecuteImportSipenmaru(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal mengeksekusi impor pendaftar: "+err.Error())
		return
	}

	msg := fmt.Sprintf("Impor selesai: %d berhasil, %d dilewati", result.ImportedCount, result.SkippedCount)
	if result.ErrorCount > 0 {
		msg += fmt.Sprintf(", %d gagal", result.ErrorCount)
	}

	response.Success(c, msg, result)
}

