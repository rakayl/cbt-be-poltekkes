package handler

import (
	"strconv"

	"poltekkes-cat-backend/internal/modules/bank_soal/dto"
	"poltekkes-cat-backend/internal/modules/bank_soal/service"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type DynamicBankSoalHandler struct {
	service service.DynamicBankSoalService
}

func NewDynamicBankSoalHandler(service service.DynamicBankSoalService) *DynamicBankSoalHandler {
	return &DynamicBankSoalHandler{service: service}
}

func (h *DynamicBankSoalHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// Stimulus & Multi-Items Routes
	rg.GET("/stimuli/:code", h.GetStimuliByBankCode)
	rg.GET("/stimuli/detail/:id", h.GetStimulusByID)
	rg.POST("/stimuli", h.CreateStimulus)
	rg.PUT("/stimuli/:id", h.UpdateStimulus)
	rg.DELETE("/stimuli/:id", h.DeleteStimulus)

	// Blueprint Composition Routes
	rg.GET("/blueprints", h.GetBlueprints)
	rg.GET("/blueprints/:id", h.GetBlueprintByID)
	rg.POST("/blueprints", h.CreateBlueprint)

	// Exam Schedule Extension (Time Window & Scoring Rules)
	rg.GET("/schedules/:schedule_id/ext", h.GetExamScheduleExt)
	rg.POST("/schedules/ext", h.SetExamScheduleExt)
}

func (h *DynamicBankSoalHandler) GetStimuliByBankCode(c *gin.Context) {
	code := c.Param("code")
	stimuli, err := h.service.GetStimuliByBankCode(c.Request.Context(), code)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat stimulus bank soal: "+err.Error())
		return
	}
	response.Success(c, "Daftar stimulus berhasil dimuat", stimuli)
}

func (h *DynamicBankSoalHandler) GetStimulusByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	stimulus, err := h.service.GetStimulusByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Stimulus tidak ditemukan: "+err.Error())
		return
	}
	response.Success(c, "Detail stimulus berhasil dimuat", stimulus)
}

func (h *DynamicBankSoalHandler) CreateStimulus(c *gin.Context) {
	var req dto.CreateStimulusDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload stimulus tidak valid: "+err.Error(), nil)
		return
	}

	res, err := h.service.CreateStimulus(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal membuat stimulus: "+err.Error())
		return
	}
	response.Success(c, "Stimulus kasus klinis dan sub-pertanyaan berhasil dibuat", res)
}

func (h *DynamicBankSoalHandler) UpdateStimulus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req dto.UpdateStimulusDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload update stimulus tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.UpdateStimulus(c.Request.Context(), id, &req)
	if err != nil {
		response.InternalServerError(c, "Gagal memperbarui stimulus: "+err.Error())
		return
	}
	response.Success(c, "Data stimulus berhasil diperbarui", nil)
}

func (h *DynamicBankSoalHandler) DeleteStimulus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := h.service.DeleteStimulus(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, "Gagal menghapus stimulus: "+err.Error())
		return
	}
	response.Success(c, "Stimulus berhasil dihapus", nil)
}

func (h *DynamicBankSoalHandler) GetBlueprints(c *gin.Context) {
	bps, err := h.service.GetBlueprints(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat blueprint: "+err.Error())
		return
	}
	response.Success(c, "Daftar blueprint ujian dinamis berhasil dimuat", bps)
}

func (h *DynamicBankSoalHandler) GetBlueprintByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	bp, err := h.service.GetBlueprintByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Blueprint tidak ditemukan: "+err.Error())
		return
	}
	response.Success(c, "Detail blueprint berhasil dimuat", bp)
}

func (h *DynamicBankSoalHandler) CreateBlueprint(c *gin.Context) {
	var req dto.CreateBlueprintDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload blueprint tidak valid: "+err.Error(), nil)
		return
	}

	res, err := h.service.CreateBlueprint(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal membuat blueprint: "+err.Error())
		return
	}
	response.Success(c, "Blueprint ujian dinamis berhasil dibuat", res)
}

func (h *DynamicBankSoalHandler) GetExamScheduleExt(c *gin.Context) {
	schedID, _ := strconv.Atoi(c.Param("schedule_id"))
	ext, err := h.service.GetExamScheduleExt(c.Request.Context(), schedID)
	if err != nil {
		response.NotFound(c, "Ekstensi jadwal tidak ditemukan / belum diset")
		return
	}
	response.Success(c, "Ekstensi jadwal berhasil dimuat", ext)
}

func (h *DynamicBankSoalHandler) SetExamScheduleExt(c *gin.Context) {
	var req dto.SetExamScheduleExtDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload ekstensi jadwal tidak valid: "+err.Error(), nil)
		return
	}

	err := h.service.SetExamScheduleExt(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menyimpan ekstensi jadwal: "+err.Error())
		return
	}
	response.Success(c, "Jendela waktu dan aturan skoring berhasil disimpan", nil)
}
