package handler

import (
	"strconv"
	"poltekkes-cat-backend/internal/modules/scoring/service"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type ScoringHandler struct {
	service service.ScoringService
}

func NewScoringHandler(service service.ScoringService) *ScoringHandler {
	return &ScoringHandler{service: service}
}

func (h *ScoringHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/recap", h.GetResults)
	rg.GET("/recap/:schedule_id", h.GetResults)
	rg.GET("/dashboard", h.GetDashboardStats)
	rg.GET("/item-analysis", h.GetItemAnalysis)
	rg.GET("/item-analysis/:schedule_id", h.GetItemAnalysis)
	rg.GET("/berita-acara", h.GetBeritaAcara)
	rg.GET("/berita-acara/:schedule_id", h.GetBeritaAcara)
}

func (h *ScoringHandler) GetResults(c *gin.Context) {
	scheduleIDStr := c.Param("schedule_id")
	if scheduleIDStr == "" {
		scheduleIDStr = c.Query("schedule_id")
	}
	scheduleID, _ := strconv.Atoi(scheduleIDStr)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	results, pagination, err := h.service.GetResults(c.Request.Context(), scheduleID, page, perPage)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat rekapitulasi nilai: "+err.Error())
		return
	}

	response.SuccessWithPagination(c, "Rekapitulasi hasil nilai ujian berhasil dimuat", results, pagination)
}

func (h *ScoringHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat statistik dashboard: "+err.Error())
		return
	}

	response.Success(c, "Data statistik dashboard berhasil dimuat", stats)
}

func (h *ScoringHandler) GetItemAnalysis(c *gin.Context) {
	scheduleIDStr := c.Param("schedule_id")
	if scheduleIDStr == "" {
		scheduleIDStr = c.Query("schedule_id")
	}
	scheduleID, _ := strconv.Atoi(scheduleIDStr)
	if scheduleID == 0 {
		scheduleID = 1 // Default fallback schedule
	}

	analysis, err := h.service.GetItemAnalysis(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat analisis butir soal: "+err.Error())
		return
	}

	response.Success(c, "Analisis psikometri butir soal berhasil dihitung", analysis)
}

func (h *ScoringHandler) GetBeritaAcara(c *gin.Context) {
	scheduleIDStr := c.Param("schedule_id")
	if scheduleIDStr == "" {
		scheduleIDStr = c.Query("schedule_id")
	}
	scheduleID, _ := strconv.Atoi(scheduleIDStr)
	if scheduleID == 0 {
		scheduleID = 1 // Default fallback schedule
	}

	ba, err := h.service.GetBeritaAcara(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat Berita Acara Ujian: "+err.Error())
		return
	}

	response.Success(c, "Data Berita Acara Ujian berhasil dimuat", ba)
}
