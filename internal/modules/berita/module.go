package berita

import (
	"poltekkes-cat-backend/internal/modules/berita/handler"
	"poltekkes-cat-backend/internal/modules/berita/repository"
	"poltekkes-cat-backend/internal/modules/berita/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(rg *gin.RouterGroup, db *sqlx.DB) {
	repo := repository.NewBeritaRepository(db)
	svc := service.NewBeritaService(repo)
	hdl := handler.NewBeritaHandler(svc)
	hdl.RegisterRoutes(rg)
}
