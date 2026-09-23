package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/bank_soal/dto"
	"poltekkes-cat-backend/internal/modules/bank_soal/service"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BankSoalHandler struct {
	service service.BankSoalService
}

func NewBankSoalHandler(service service.BankSoalService) *BankSoalHandler {
	return &BankSoalHandler{service: service}
}

func (h *BankSoalHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// Bank Soal Header Routes
	rg.GET("", h.GetQuestionBanks)
	rg.POST("", h.CreateQuestionBank)
	rg.GET("/:code", h.GetQuestionBankByCode)
	rg.PUT("/:code", h.UpdateQuestionBank)
	rg.DELETE("/:code", h.DeleteQuestionBank)
	rg.POST("/:code/copy", h.CopyQuestionBank)

	// Butir Soal & Options Routes
	rg.GET("/:code/questions", h.GetQuestionsByCode)
	rg.POST("/:code/questions", h.CreateQuestionItem)
	rg.PUT("/:code/questions/:nourut", h.UpdateQuestionItem)
	rg.DELETE("/:code/questions/:nourut", h.DeleteQuestionItem)

	// Multimedia Upload
	rg.POST("/upload-media", h.UploadMedia)
}

func (h *BankSoalHandler) GetQuestionBanks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	search := c.DefaultQuery("search", "")

	banks, pagination, err := h.service.GetQuestionBanks(c.Request.Context(), page, perPage, search)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar bank soal: "+err.Error())
		return
	}

	response.SuccessWithPagination(c, "Daftar bank soal berhasil dimuat", banks, pagination)
}

func (h *BankSoalHandler) GetQuestionBankByCode(c *gin.Context) {
	code := c.Param("code")
	bank, err := h.service.GetQuestionBankByCode(c.Request.Context(), code)
	if err != nil {
		response.NotFound(c, "Bank soal tidak ditemukan: "+err.Error())
		return
	}

	response.Success(c, "Detail bank soal berhasil dimuat", bank)
}

func (h *BankSoalHandler) CreateQuestionBank(c *gin.Context) {
	var req dto.CreateQuestionBankDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data bank soal tidak valid: "+err.Error(), nil)
		return
	}

	res, err := h.service.CreateQuestionBank(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal membuat bank soal baru: "+err.Error())
		return
	}

	response.Success(c, "Paket bank soal berhasil dibuat", res)
}

func (h *BankSoalHandler) UpdateQuestionBank(c *gin.Context) {
	code := c.Param("code")
	var req dto.UpdateQuestionBankDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data perubahan bank soal tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.UpdateQuestionBank(c.Request.Context(), code, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui bank soal: "+err.Error())
		return
	}

	response.Success(c, "Data bank soal berhasil diperbarui", nil)
}

func (h *BankSoalHandler) DeleteQuestionBank(c *gin.Context) {
	code := c.Param("code")
	err := h.service.DeleteQuestionBank(c.Request.Context(), code)
	if err != nil {
		response.InternalServerError(c, "Gagal menghapus bank soal: "+err.Error())
		return
	}

	response.Success(c, "Bank soal berhasil dihapus", nil)
}

func (h *BankSoalHandler) CopyQuestionBank(c *gin.Context) {
	sourceCode := c.Param("code")
	var req dto.CopyQuestionBankDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data duplikasi bank soal tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.CopyQuestionBank(c.Request.Context(), sourceCode, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menduplikasi bank soal: "+err.Error())
		return
	}

	response.Success(c, "Bank soal berhasil diduplikasi", nil)
}

func (h *BankSoalHandler) GetQuestionsByCode(c *gin.Context) {
	code := c.Param("code")
	questions, err := h.service.GetQuestionsByCode(c.Request.Context(), code)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat butir soal: "+err.Error())
		return
	}

	response.Success(c, "Daftar butir pertanyaan berhasil dimuat", questions)
}

func (h *BankSoalHandler) CreateQuestionItem(c *gin.Context) {
	code := c.Param("code")
	var req dto.SaveQuestionItemDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data butir soal tidak valid: "+err.Error(), nil)
		return
	}

	res, err := h.service.CreateQuestionItem(c.Request.Context(), code, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menambahkan butir soal: "+err.Error())
		return
	}

	response.Success(c, "Butir soal baru berhasil ditambahkan", res)
}

func (h *BankSoalHandler) UpdateQuestionItem(c *gin.Context) {
	code := c.Param("code")
	noUrut, _ := strconv.Atoi(c.Param("nourut"))

	var req dto.SaveQuestionItemDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data edit butir soal tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.UpdateQuestionItem(c.Request.Context(), code, noUrut, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui butir soal: "+err.Error())
		return
	}

	response.Success(c, "Butir soal berhasil diperbarui", nil)
}

func (h *BankSoalHandler) DeleteQuestionItem(c *gin.Context) {
	code := c.Param("code")
	noUrut, _ := strconv.Atoi(c.Param("nourut"))

	err := h.service.DeleteQuestionItem(c.Request.Context(), code, noUrut)
	if err != nil {
		response.InternalServerError(c, "Gagal menghapus butir soal: "+err.Error())
		return
	}

	response.Success(c, "Butir soal berhasil dihapus", nil)
}

func (h *BankSoalHandler) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File tidak ditemukan dalam request: "+err.Error(), nil)
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".mp3": true, ".wav": true, ".ogg": true, ".m4a": true, ".aac": true, ".flac": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true,
		".mp4": true, ".webm": true, ".mov": true, ".mkv": true,
	}

	if !allowed[ext] {
		response.BadRequest(c, fmt.Sprintf("Format file '%s' tidak didukung", ext), nil)
		return
	}

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
		response.InternalServerError(c, "Gagal menyimpan file: "+err.Error())
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
