package handler

import (
	"poltekkes-cat-backend/internal/modules/referensi/dto"
	"poltekkes-cat-backend/internal/modules/referensi/service"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type ReferensiHandler struct {
	service service.ReferensiService
}

func NewReferensiHandler(service service.ReferensiService) *ReferensiHandler {
	return &ReferensiHandler{service: service}
}

func (h *ReferensiHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// Ruang
	rg.GET("/ruang", h.GetRuangList)
	rg.POST("/ruang", h.CreateRuang)
	rg.PUT("/ruang/:kode_ruang", h.UpdateRuang)
	rg.DELETE("/ruang/:kode_ruang", h.DeleteRuang)

	// Skor
	rg.GET("/skor", h.GetSkorList)

	// Jenis Ujian
	rg.GET("/jenis-ujian", h.GetJenisUjianList)
	rg.POST("/jenis-ujian", h.CreateJenisUjian)
	rg.PUT("/jenis-ujian/:kode_jenis", h.UpdateJenisUjian)
	rg.DELETE("/jenis-ujian/:kode_jenis", h.DeleteJenisUjian)

	// Wilayah
	rg.GET("/wilayah", h.GetWilayahList)

	// Panitia
	rg.GET("/panitia", h.GetPanitiaList)
	rg.POST("/panitia", h.CreatePanitia)
	rg.PUT("/panitia/:nip", h.UpdatePanitia)
	rg.DELETE("/panitia/:nip", h.DeletePanitia)

	// Jenis Periode
	rg.GET("/jenis-periode", h.GetJenisPeriodeList)
	rg.POST("/jenis-periode", h.CreateJenisPeriode)
	rg.PUT("/jenis-periode/:jenis_periode", h.UpdateJenisPeriode)
	rg.DELETE("/jenis-periode/:jenis_periode", h.DeleteJenisPeriode)

	// Unit & Salin Soal
	rg.GET("/unit", h.GetUnitList)
	rg.POST("/salin-soal", h.SalinSoal)
}

func (h *ReferensiHandler) GetRuangList(c *gin.Context) {
	data, err := h.service.GetRuangList(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat master ruang: "+err.Error())
		return
	}
	response.Success(c, "Master ruang ujian berhasil dimuat", data)
}

func (h *ReferensiHandler) GetSkorList(c *gin.Context) {
	data, err := h.service.GetSkorList(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat master skor: "+err.Error())
		return
	}
	response.Success(c, "Master konfigurasi skor berhasil dimuat", data)
}

func (h *ReferensiHandler) GetJenisUjianList(c *gin.Context) {
	data, err := h.service.GetJenisUjianList(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat master jenis ujian: "+err.Error())
		return
	}
	response.Success(c, "Master jenis ujian berhasil dimuat", data)
}

func (h *ReferensiHandler) GetWilayahList(c *gin.Context) {
	parentID := c.Query("parent_id")
	data, err := h.service.GetWilayahList(c.Request.Context(), parentID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat master wilayah: "+err.Error())
		return
	}
	response.Success(c, "Master wilayah berhasil dimuat", data)
}

func (h *ReferensiHandler) GetPanitiaList(c *gin.Context) {
	data, err := h.service.GetPanitiaList(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat master panitia: "+err.Error())
		return
	}
	response.Success(c, "Master panitia berhasil dimuat", data)
}

func (h *ReferensiHandler) GetJenisPeriodeList(c *gin.Context) {
	data, err := h.service.GetJenisPeriodeList(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat master jenis periode: "+err.Error())
		return
	}
	response.Success(c, "Master jenis periode berhasil dimuat", data)
}

func (h *ReferensiHandler) GetUnitList(c *gin.Context) {
	data, err := h.service.GetUnitList(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal memuat master unit/satker: "+err.Error())
		return
	}
	response.Success(c, "Master unit/satker berhasil dimuat", data)
}

func (h *ReferensiHandler) SalinSoal(c *gin.Context) {
	var req dto.SalinSoalRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data salin soal tidak valid", nil)
		return
	}

	res, err := h.service.SalinSoal(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menyalin butir soal: "+err.Error())
		return
	}

	response.Success(c, "Butir soal berhasil disalin ke bank soal baru", res)
}

func (h *ReferensiHandler) CreateRuang(c *gin.Context) {
	var req dto.CreateRuangRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data ruang tidak valid: "+err.Error(), nil)
		return
	}
	kode, err := h.service.CreateRuang(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menambah ruang ujian: "+err.Error())
		return
	}
	response.Success(c, "Ruang ujian berhasil ditambahkan", map[string]string{"kode_ruang": kode})
}

func (h *ReferensiHandler) UpdateRuang(c *gin.Context) {
	kodeRuang := c.Param("kode_ruang")
	var req dto.UpdateRuangRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data ruang tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.UpdateRuang(c.Request.Context(), kodeRuang, &req); err != nil {
		response.InternalServerError(c, "Gagal memperbarui ruang ujian: "+err.Error())
		return
	}
	response.Success(c, "Ruang ujian berhasil diperbarui", nil)
}

func (h *ReferensiHandler) DeleteRuang(c *gin.Context) {
	kodeRuang := c.Param("kode_ruang")
	if err := h.service.DeleteRuang(c.Request.Context(), kodeRuang); err != nil {
		response.InternalServerError(c, "Gagal menghapus ruang ujian: "+err.Error())
		return
	}
	response.Success(c, "Ruang ujian berhasil dihapus", nil)
}

func (h *ReferensiHandler) CreateJenisUjian(c *gin.Context) {
	var req dto.CreateJenisUjianRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data jenis ujian tidak valid: "+err.Error(), nil)
		return
	}
	kode, err := h.service.CreateJenisUjian(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Gagal menambah jenis ujian: "+err.Error())
		return
	}
	response.Success(c, "Jenis ujian berhasil ditambahkan", map[string]string{"kode_jenis": kode})
}

func (h *ReferensiHandler) UpdateJenisUjian(c *gin.Context) {
	kodeJenis := c.Param("kode_jenis")
	var req dto.UpdateJenisUjianRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data jenis ujian tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.UpdateJenisUjian(c.Request.Context(), kodeJenis, &req); err != nil {
		response.InternalServerError(c, "Gagal memperbarui jenis ujian: "+err.Error())
		return
	}
	response.Success(c, "Jenis ujian berhasil diperbarui", nil)
}

func (h *ReferensiHandler) DeleteJenisUjian(c *gin.Context) {
	kodeJenis := c.Param("kode_jenis")
	if err := h.service.DeleteJenisUjian(c.Request.Context(), kodeJenis); err != nil {
		response.InternalServerError(c, "Gagal menghapus jenis ujian: "+err.Error())
		return
	}
	response.Success(c, "Jenis ujian berhasil dihapus", nil)
}

func (h *ReferensiHandler) CreatePanitia(c *gin.Context) {
	var req dto.CreatePanitiaRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data panitia tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.CreatePanitia(c.Request.Context(), &req); err != nil {
		response.InternalServerError(c, "Gagal menambah panitia: "+err.Error())
		return
	}
	response.Success(c, "Panitia berhasil ditambahkan", nil)
}

func (h *ReferensiHandler) UpdatePanitia(c *gin.Context) {
	nip := c.Param("nip")
	var req dto.UpdatePanitiaRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data panitia tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.UpdatePanitia(c.Request.Context(), nip, &req); err != nil {
		response.InternalServerError(c, "Gagal memperbarui panitia: "+err.Error())
		return
	}
	response.Success(c, "Panitia berhasil diperbarui", nil)
}

func (h *ReferensiHandler) DeletePanitia(c *gin.Context) {
	nip := c.Param("nip")
	if err := h.service.DeletePanitia(c.Request.Context(), nip); err != nil {
		response.InternalServerError(c, "Gagal menghapus panitia: "+err.Error())
		return
	}
	response.Success(c, "Panitia berhasil dihapus", nil)
}

func (h *ReferensiHandler) CreateJenisPeriode(c *gin.Context) {
	var req dto.CreateJenisPeriodeRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data jenis periode tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.CreateJenisPeriode(c.Request.Context(), &req); err != nil {
		response.InternalServerError(c, "Gagal menambah jenis periode: "+err.Error())
		return
	}
	response.Success(c, "Jenis periode berhasil ditambahkan", nil)
}

func (h *ReferensiHandler) UpdateJenisPeriode(c *gin.Context) {
	jenisPeriode := c.Param("jenis_periode")
	var req dto.UpdateJenisPeriodeRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data jenis periode tidak valid: "+err.Error(), nil)
		return
	}
	if err := h.service.UpdateJenisPeriode(c.Request.Context(), jenisPeriode, &req); err != nil {
		response.InternalServerError(c, "Gagal memperbarui jenis periode: "+err.Error())
		return
	}
	response.Success(c, "Jenis periode berhasil diperbarui", nil)
}

func (h *ReferensiHandler) DeleteJenisPeriode(c *gin.Context) {
	jenisPeriode := c.Param("jenis_periode")
	if err := h.service.DeleteJenisPeriode(c.Request.Context(), jenisPeriode); err != nil {
		response.InternalServerError(c, "Gagal menghapus jenis periode: "+err.Error())
		return
	}
	response.Success(c, "Jenis periode berhasil dihapus", nil)
}
