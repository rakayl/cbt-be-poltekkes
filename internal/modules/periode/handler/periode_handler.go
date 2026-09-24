package handler

import (
	"poltekkes-cat-backend/internal/modules/periode/dto"
	"poltekkes-cat-backend/internal/modules/periode/service"
	"poltekkes-cat-backend/internal/shared/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PeriodeHandler struct {
	service service.PeriodeService
}

func NewPeriodeHandler(service service.PeriodeService) *PeriodeHandler {
	return &PeriodeHandler{service: service}
}

func (h *PeriodeHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.GetPeriods)
	rg.POST("", h.CreatePeriod)
	rg.PUT("/:id", h.UpdatePeriod)
	rg.DELETE("/:id", h.DeletePeriod)
}

func (h *PeriodeHandler) GetPeriods(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")

	periods, pagination, err := h.service.GetPeriods(c.Request.Context(), page, perPage, search)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar periode: "+err.Error())
		return
	}

	response.SuccessWithPagination(c, "Daftar periode ujian berhasil dimuat", periods, pagination)
}

func (h *PeriodeHandler) CreatePeriod(c *gin.Context) {
	var req dto.CreatePeriodRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data periode tidak valid", nil)
		return
	}

	id, err := h.service.CreatePeriod(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal membuat periode: "+err.Error())
		return
	}

	response.Created(c, "Periode ujian berhasil dibuat", map[string]int{"period_id": id})
}

func (h *PeriodeHandler) UpdatePeriod(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID periode tidak valid", nil)
		return
	}

	var req dto.UpdatePeriodRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data periode tidak valid", nil)
		return
	}

	if err := h.service.UpdatePeriod(c.Request.Context(), id, &req); err != nil {
		response.InternalServerError(c, "Gagal memperbarui periode: "+err.Error())
		return
	}

	response.Success(c, "Periode ujian berhasil diperbarui", map[string]int{"period_id": id})
}

func (h *PeriodeHandler) DeletePeriod(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID periode tidak valid", nil)
		return
	}

	if err := h.service.DeletePeriod(c.Request.Context(), id); err != nil {
		response.InternalServerError(c, "Gagal menghapus periode: "+err.Error())
		return
	}

	response.Success(c, "Periode ujian berhasil dihapus", map[string]int{"period_id": id})
}
