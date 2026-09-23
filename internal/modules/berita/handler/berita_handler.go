package handler

import (
	"poltekkes-cat-backend/internal/modules/berita/dto"
	"poltekkes-cat-backend/internal/modules/berita/service"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type BeritaHandler struct {
	service service.BeritaService
}

func NewBeritaHandler(service service.BeritaService) *BeritaHandler {
	return &BeritaHandler{service: service}
}

func (h *BeritaHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.GetAllBerita)
	rg.POST("", h.CreateBerita)
}

func (h *BeritaHandler) GetAllBerita(c *gin.Context) {
	data, err := h.service.GetAllBerita(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat berita: "+err.Error())
		return
	}

	response.Success(c, "Daftar berita & pengumuman berhasil dimuat", data)
}

func (h *BeritaHandler) CreateBerita(c *gin.Context) {
	var req dto.CreateBeritaRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data berita tidak valid", nil)
		return
	}

	id, err := h.service.CreateBerita(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menambahkan berita: "+err.Error())
		return
	}

	response.Created(c, "Berita berhasil diterbitkan", map[string]int{"berita_id": id})
}
