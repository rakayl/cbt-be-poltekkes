package scoring

import (
	"poltekkes-cat-backend/internal/modules/scoring/handler"
	"poltekkes-cat-backend/internal/modules/scoring/repository"
	"poltekkes-cat-backend/internal/modules/scoring/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(rg *gin.RouterGroup, db *sqlx.DB) {
	repo := repository.NewScoringRepository(db)
	svc := service.NewScoringService(repo)
	hdl := handler.NewScoringHandler(svc)
	hdl.RegisterRoutes(rg)
}
