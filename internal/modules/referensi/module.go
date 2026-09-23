package referensi

import (
	"poltekkes-cat-backend/internal/modules/referensi/handler"
	"poltekkes-cat-backend/internal/modules/referensi/repository"
	"poltekkes-cat-backend/internal/modules/referensi/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Init(rg *gin.RouterGroup, db *sqlx.DB, siakadDBs ...*sqlx.DB) {
	repo := repository.NewReferensiRepository(db, siakadDBs...)
	svc := service.NewReferensiService(repo)
	hdl := handler.NewReferensiHandler(svc)
	hdl.RegisterRoutes(rg)
}
