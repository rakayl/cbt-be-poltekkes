package peserta

import (
	"poltekkes-cat-backend/internal/modules/peserta/handler"
	"poltekkes-cat-backend/internal/modules/peserta/repository"
	"poltekkes-cat-backend/internal/modules/peserta/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(rg *gin.RouterGroup, db *sqlx.DB, siakadDBs ...*sqlx.DB) {
	repo := repository.NewPesertaRepository(db, siakadDBs...)
	svc := service.NewPesertaService(repo)
	hdl := handler.NewPesertaHandler(svc)
	hdl.RegisterRoutes(rg)
}
