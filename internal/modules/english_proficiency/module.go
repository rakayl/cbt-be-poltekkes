package english_proficiency

import (
	"poltekkes-cat-backend/internal/modules/english_proficiency/handler"
	"poltekkes-cat-backend/internal/modules/english_proficiency/repository"
	"poltekkes-cat-backend/internal/modules/english_proficiency/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(rg *gin.RouterGroup, db *sqlx.DB) {
	repo := repository.NewEPRepository(db)
	svc := service.NewEPService(repo)
	hdl := handler.NewEPHandler(svc)
	hdl.RegisterRoutes(rg)
}
