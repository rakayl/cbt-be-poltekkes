package handler

import (
	"errors"
	"strings"

	"poltekkes-cat-backend/internal/modules/auth/dto"
	"poltekkes-cat-backend/internal/modules/auth/service"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.LoginAdmin)
	rg.POST("/switch-module", h.SwitchModule)
	rg.GET("/modules", h.GetModuleMenus)
	rg.GET("/menus", h.GetModuleMenus)
	rg.POST("/participant/login", h.LoginParticipant)
	rg.POST("/ep/login", h.LoginEPParticipant)
}

func (h *AuthHandler) LoginAdmin(c *gin.Context) {
	var req dto.AdminLoginRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Username dan password wajib diisi", nil)
		return
	}

	res, err := h.service.LoginAdmin(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrNoAccessRole) || strings.Contains(err.Error(), "tidak memiliki hak akses") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Autentikasi berhasil", res)
}

func (h *AuthHandler) LoginParticipant(c *gin.Context) {
	var req dto.ParticipantLoginRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Nomor peserta dan kata sandi wajib diisi", nil)
		return
	}

	res, err := h.service.LoginParticipant(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Login peserta berhasil", res)
}

func (h *AuthHandler) LoginEPParticipant(c *gin.Context) {
	var req dto.ParticipantLoginRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Identitas peserta dan kata sandi wajib diisi", nil)
		return
	}

	res, err := h.service.LoginEPParticipant(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Login peserta TOEFL berhasil", res)
}

func (h *AuthHandler) GetModuleMenus(c *gin.Context) {
	moduleID := c.DefaultQuery("module", "cat")
	roleID := c.DefaultQuery("role", "admin")

	menus, err := h.service.GetModuleMenus(c.Request.Context(), moduleID, roleID)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat menu modul: "+err.Error())
		return
	}

	response.Success(c, "Daftar menu modul berhasil dimuat", menus)
}

func (h *AuthHandler) SwitchModule(c *gin.Context) {
	var req dto.SwitchModuleRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data perpindahan modul tidak valid", nil)
		return
	}

	res, err := h.service.SwitchModule(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, "Token akses modul berhasil digenerate", res)
}
