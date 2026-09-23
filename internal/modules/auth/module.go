package auth

import (
	"poltekkes-cat-backend/internal/modules/auth/handler"
	"poltekkes-cat-backend/internal/modules/auth/repository"
	"poltekkes-cat-backend/internal/modules/auth/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(rg *gin.RouterGroup, cbtDB *sqlx.DB, siakadDBs ...*sqlx.DB) {
	repo := repository.NewAuthRepository(cbtDB, siakadDBs...)
	svc := service.NewAuthService(repo)
	hdl := handler.NewAuthHandler(svc)
	hdl.RegisterRoutes(rg)
}

